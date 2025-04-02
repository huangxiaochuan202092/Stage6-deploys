package handlers

import (
	"log"
	"net/http"
	"proapp/config"
	"proapp/services"
	"proapp/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 发送验证码
func SendCodeHandler(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
	}

	// 绑定 JSON 输入
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 生成验证码
	code := utils.GenerateVerificationCode()

	// 发送验证码邮件
	if err := utils.SendVerificationEmail(input.Email, code); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 存储验证码
	if err := utils.SetVerificationCode(input.Email, code); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "验证码发送成功"})
}

// 登录或注册
func LoginOrRegisterHandler(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
		Code  string `json:"code" binding:"required"`
	}

	// 绑定 JSON 输入
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("绑定 JSON 失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("处理登录/注册请求: email=%s", input.Email)

	// 获取存储的验证码
	storedCode, err := utils.GetVerificationCode(input.Email)
	if err != nil {
		log.Printf("获取验证码失败: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "验证码获取失败"})
		return
	}

	// 验证码匹配
	if storedCode != input.Code {
		log.Printf("验证码不匹配: 输入验证码=%s, 存储验证码=%s", input.Code, storedCode)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "验证码不匹配"})
		return
	}

	log.Printf("验证码验证通过，开始检查用户是否存在: email=%s", input.Email)

	// 检查数据库连接
	if config.DB == nil {
		log.Printf("严重错误: 数据库连接未初始化")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库服务不可用"})
		return
	}

	// 检查用户是否存在
	user, err := services.GetUserByEmail(input.Email)
	if err != nil {
		log.Printf("获取用户失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误: " + err.Error()})
		return
	}

	log.Printf("获取用户结果: %+v", user)

	// 如果用户不存在，则创建用户
	if user == nil {
		log.Printf("用户不存在，尝试创建新用户: email=%s", input.Email)
		user, err = services.CreateUser(input.Email)
		if err != nil {
			log.Printf("创建用户失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败: " + err.Error()})
			return
		}
	}

	// 删除验证码
	if err := utils.DelVerificationCode(input.Email); err != nil {
		log.Printf("删除验证码失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "验证码删除失败"})
		return
	}

	// 生成token
	token, err := utils.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成token失败"})
		return
	}

	// 返回用户信息
	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
		"token":   token,
		"user": gin.H{
			"email": user.Email,
			"id":    user.ID,
			"role":  user.Role,
		},
	})
}

// 获取所有用户
func GetAllUsersHandler(c *gin.Context) {
	// 记录请求来源和认证信息
	authHeader := c.GetHeader("Authorization")
	log.Printf("GetAllUsersHandler 被调用, 认证头: %v", authHeader != "")

	// 检查用户权限
	role, exists := c.Get("userRole")
	if !exists {
		log.Printf("错误: 未找到用户角色信息")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问，缺少角色信息"})
		return
	}
	log.Printf("当前用户角色: %v", role)

	users, err := services.GetAllUsers()
	if err != nil {
		log.Printf("获取所有用户数据失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户失败: " + err.Error()})
		return
	}

	// 添加调试日志，检查返回的用户数据
	log.Printf("返回用户数据，共 %d 条记录", len(users))
	for i, user := range users {
		log.Printf("用户 %d: ID=%d, Email=%s, Role=%s", i+1, user.ID, user.Email, user.Role)
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

// 根据id获取用户
func GetUserByIdHandler(c *gin.Context) {
	id := c.Param("id")
	idUint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	user, err := services.GetUserByID(uint(idUint)) // 修改这里：GetUserById -> GetUserByID
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// 更新用户
func UpdateUserHandler(c *gin.Context) {
	targetID := c.Param("id")
	targetIDUint, err := strconv.ParseUint(targetID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 管理员可以更新任何用户
	var adminInput struct {
		Email string `json:"email" binding:"required,email"`
		Role  string `json:"role" binding:"omitempty,oneof=user admin"`
	}
	if err := c.ShouldBindJSON(&adminInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{
		"email": adminInput.Email,
	}
	if adminInput.Role != "" {
		updates["role"] = adminInput.Role
	}

	// 直接调用更新，不做额外权限检查
	if err := services.UpdateUserFields(uint(targetIDUint), updates); err != nil {
		log.Printf("更新用户失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户失败"})
		return
	}

	// 清除可能的查询缓存或实例状态
	config.DB = config.DB.Session(&gorm.Session{NewDB: true})

	c.JSON(http.StatusOK, gin.H{
		"message": "更新用户成功",
		"user": gin.H{
			"id":    targetIDUint,
			"email": adminInput.Email,
			"role":  adminInput.Role,
		},
	})
}

// 删除用户
func DeleteUserHandler(c *gin.Context) {
	// 获取路径参数 id
	id := c.Param("id")
	idUint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 调用服务层删除用户
	if err := services.DeleteUserByID(uint(idUint)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除用户失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除用户成功"})
}

// 获取用户信息
func GetUserInfo(c *gin.Context) {
	// 从 token 中获取用户 ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	user, err := services.GetUserByID(uint(userID.(float64)))
	if err != nil {
		log.Printf("获取用户信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"role":       user.Role,
			"created_at": user.CreatedAt,
		},
	})
}

// 获取用户的任务列表
func GetUserTasks(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	tasks, err := services.GetTasksByUserID(uint(userID.(float64)))
	if err != nil {
		log.Printf("获取用户任务失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取任务列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"tasks":  tasks,
	})
}

func UpdateUserSelfHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	var input struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := services.UpdateUserEmail(uint(userID.(float64)), input.Email); err != nil {
		log.Printf("更新用户信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新用户信息失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新用户信息成功"})
}
