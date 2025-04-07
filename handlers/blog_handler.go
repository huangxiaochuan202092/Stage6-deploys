package handlers

import (
	"log"
	"net/http"
	"proapp/models"
	"proapp/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetAllBlogs 获取所有博客
func GetAllBlogs(c *gin.Context) {
	// 获取分页参数
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize <= 0 {
		pageSize = 20
	}

	// 获取搜索关键词
	keyword := c.Query("keyword")

	var blogs []models.Blog
	var total int64 // 总数

	if keyword != "" {
		// 如果有搜索关键词，调用搜索函数
		blogs, total, err = services.SearchBlogs(keyword, page, pageSize)
		log.Printf("搜索博客，关键词：%s，页码：%d，每页数量：%d", keyword, page, pageSize)
	} else {
		// 否则获取所有博客
		blogs, total, err = services.GetAllBlogs(page, pageSize)
		log.Printf("获取所有博客，页码：%d，每页数量：%d", page, pageSize)
	}

	if err != nil {
		log.Printf("获取博客列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取博客列表失败"})
		return
	}

	// 计算总页数
	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"blogs":  blogs, // 确保直接返回博客数组
		"total":  total, // 添加总数
		"pagination": gin.H{
			"current_page": page,
			"page_size":    pageSize,
			"total_pages":  totalPages,
			"total_count":  total,
		},
	})
}

// GetBlogById 获取单个博客
func GetBlogById(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的博客ID"})
		return
	}

	blog, err := services.GetBlogById(id)
	if err != nil {
		log.Printf("获取博客详情失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取博客详情失败"})
		return
	}

	if blog == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "博客不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"blog": blog,
		},
	})
}

// CreateBlog 创建博客 - 修复日志打印语句
func CreateBlog(c *gin.Context) {
	var input struct {
		Title    string `json:"title" binding:"required"`
		Content  string `json:"content" binding:"required"`
		Category string `json:"category"`
		Tags     string `json:"tags"`
		Status   string `json:"status"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 从JWT中获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无法获取用户信息"})
		return
	}

	// 创建博客对象
	blog := &models.Blog{
		Title:    input.Title,
		Content:  input.Content,
		Category: input.Category,
		Tags:     input.Tags,
		Status:   input.Status,
	}

	// 直接将UserID设置为正确类型，避免类型转换错误
	if uid, ok := userID.(uint); ok {
		blog.UserID = uid
	} else if uid, ok := userID.(float64); ok {
		blog.UserID = uint(uid)
	} else if uid, ok := userID.(int); ok {
		blog.UserID = uint(uid)
	} else {
		// 修复这里的语法错误 - 确保字符串不会被换行符中断
		log.Printf("userID类型错误: 类型=%T, 值=%v", userID, userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户ID类型错误"})
		return
	}

	// 继续执行创建博客的逻辑
	if err := services.CreateBlog(blog); err != nil {
		log.Printf("创建博客失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建博客失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "博客创建成功",
		"blog":    blog,
	})
}

// LikeBlog 改进点赞处理函数，确保正确处理错误和返回值
func LikeBlog(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		log.Printf("无效的博客ID: %s, 错误: %v", idParam, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的博客ID"})
		return
	}

	// 检查博客是否存在
	blog, err := services.GetBlogById(id)
	if err != nil {
		log.Printf("获取博客失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取博客失败"})
		return
	}
	if blog == nil {
		log.Printf("博客不存在，ID: %d", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "博客不存在"})
		return
	}

	// 从JWT中获取用户ID - 记录用户ID但暂时不使用
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无法获取用户信息"})
		return
	}

	// 记录用户ID，用于将来扩展功能
	log.Printf("用户ID=%v正在点赞博客ID=%d", userID, id)

	// 执行点赞操作，并获取更新后的点赞数
	newLikes, err := services.LikeBlog(id)
	if err != nil {
		log.Printf("点赞博客失败 ID=%d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "点赞失败: " + err.Error()})
		return
	}

	log.Printf("博客点赞成功，ID=%d，当前点赞数=%d", id, newLikes)
	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"message":     "点赞成功",
		"liked":       true,     // 添加liked状态
		"total_likes": newLikes, // 重命名为total_likes以匹配前端期望
	})
}

// 添加获取点赞状态的API
func GetLikeStatus(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的博客ID"})
		return
	}

	// 检查博客是否存在
	blog, err := services.GetBlogById(id)
	if err != nil {
		log.Printf("获取博客失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取博客失败"})
		return
	}
	if blog == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "博客不存在"})
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无法获取用户信息"})
		return
	}

	// 记录用户ID，用于将来扩展功能
	log.Printf("正在检查用户ID=%v是否点赞了博客ID=%d", userID, id)

	// 这里假设已点赞，实际应该查询数据库
	// 在实际实现中，您应该查询点赞记录表确认用户是否已点赞
	liked := true

	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"liked":       liked,
		"total_likes": blog.Likes,
	})
}

// DislikeBlog 取消点赞博客
func DislikeBlog(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的博客ID"})
		return
	}

	// 检查博客是否存在
	blog, err := services.GetBlogById(id)
	if err != nil {
		log.Printf("获取博客失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取博客失败"})
		return
	}
	if blog == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "博客不存在"})
		return
	}

	// 执行取消点赞操作
	if err := services.DislikeBlog(id); err != nil {
		log.Printf("取消点赞失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "取消点赞失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "取消点赞成功",
	})
}

// UpdateBlog 更新博客
func UpdateBlog(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的博客ID"})
		return
	}

	// 检查博客是否存在
	blog, err := services.GetBlogById(id)
	if err != nil {
		log.Printf("获取博客失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取博客失败"})
		return
	}
	if blog == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "博客不存在"})
		return
	}

	// 从请求中获取更新数据
	var input struct {
		Title    string `json:"title"`
		Content  string `json:"content"`
		Category string `json:"category"`
		Tags     string `json:"tags"`
		Status   string `json:"status"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("绑定更新请求数据失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 构建更新数据
	updates := make(map[string]interface{})

	if input.Title != "" {
		updates["title"] = input.Title
	}
	if input.Content != "" {
		updates["content"] = input.Content
	}
	if input.Category != "" {
		updates["category"] = input.Category
	}
	if input.Tags != "" {
		updates["tags"] = input.Tags
	}
	if input.Status != "" {
		updates["status"] = input.Status
	}

	// 如果没有数据更新
	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "没有提供更新数据"})
		return
	}

	// 执行更新
	if err := services.UpdateBlog(id, updates); err != nil {
		log.Printf("更新博客失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新博客失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "博客更新成功",
	})
}

// DeleteBlog 删除博客
func DeleteBlog(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的博客ID"})
		return
	}

	// 检查博客是否存在
	blog, err := services.GetBlogById(id)
	if err != nil {
		log.Printf("获取博客失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取博客失败"})
		return
	}
	if blog == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "博客不存在"})
		return
	}

	// 执行删除
	if err := services.DeleteBlog(id); err != nil {
		log.Printf("删除博客失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除博客失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "博客删除成功",
	})
}

// 注意：ValidateToken 和 RefreshToken 函数已移除，请使用 auth_handler.go 中的实现
