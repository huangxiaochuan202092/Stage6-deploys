package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"proapp/models"
	"proapp/services"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// 创建问卷
func CreateWenjuan(c *gin.Context) {
	// 从 token 中获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未登录或登录已过期",
		})
		return
	}

	var input struct {
		Title      string     `json:"title" binding:"required"`
		Content    string     `json:"content" binding:"required"`
		Status     string     `json:"status"`
		Deadline   *time.Time `json:"deadline"`
		CategoryId uint       `json:"category_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 验证状态值
	if input.Status != "" && input.Status != "draft" && input.Status != "published" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "状态值无效,只能是draft或published",
		})
		return
	}

	// 验证截止时间
	if input.Deadline != nil && input.Deadline.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "截止时间不能早于当前时间",
		})
		return
	}

	// 解析问题内容
	var questions []string
	if err := json.Unmarshal([]byte(input.Content), &questions); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "问题格式错误",
			"error":   err.Error(),
		})
		return
	}

	wenjuan := &models.Wenjuan{
		Title:     input.Title,
		Content:   input.Content,
		Status:    input.Status,
		Deadline:  input.Deadline,
		CreatorID: userID.(uint), // 添加创建者ID
	}

	// 创建问卷并关联分类
	if err := services.CreateWenjuanWithCategory(wenjuan, input.CategoryId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建问卷失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
		"data":    wenjuan,
	})
}

// 修改获取问卷列表的处理函数
func GetAllWenjuans(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	isPinnedStr := c.DefaultQuery("isPinned", "")
	var isPinned *bool
	if isPinnedStr != "" {
		val, err := strconv.ParseBool(isPinnedStr)
		if err == nil {
			isPinned = &val
		}
	}

	data, err := services.GetAllWenjuans(page, pageSize, isPinned)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取问卷列表失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    data,
	})
}

// 获取单个问卷
func GetWenjuanById(c *gin.Context) {
	id := c.Param("id")
	idUint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	wenjuan, err := services.GetWenjuanById(int(idUint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, wenjuan)
}

func GetWenjuanAnswers(c *gin.Context) {
	id := c.Param("id")
	idUint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	answers, err := services.GetWenjuanAnswers(int(idUint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, answers)
}

func GetWenjuanAnswerStats(c *gin.Context) {
	id := c.Param("id")
	idUint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	stats, err := services.GetWenjuanAnswerStats(int(idUint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// 修改更新问卷处理函数
func UpdateWenjuan(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的问卷ID",
		})
		return
	}

	var input struct {
		Title    string     `json:"title"`
		Content  string     `json:"content"`
		Status   string     `json:"status"`
		Deadline *time.Time `json:"deadline"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	updates := make(map[string]interface{})

	if input.Title != "" {
		updates["title"] = input.Title
	}
	if input.Content != "" {
		// 验证 content 格式
		var questions []string
		if err := json.Unmarshal([]byte(input.Content), &questions); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "问题格式错误",
				"error":   err.Error(),
			})
			return
		}
		updates["content"] = input.Content
	}
	if input.Status != "" {
		updates["status"] = input.Status
	}
	if input.Deadline != nil {
		updates["deadline"] = *input.Deadline
	}

	// 调用服务层更新
	if err := services.UpdateWenjuan(id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新失败",
			"error":   err.Error(),
		})
		return
	}

	// 返回更新后的问卷
	wenjuan, err := services.GetWenjuanById(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取更新后的问卷失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    wenjuan,
	})
}

// 删除问卷
func DeleteWenjuan(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的ID",
			"error":   err.Error(),
		})
		return
	}

	if err := services.DeleteWenjuan(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "问卷删除成功",
	})
}

// 修改提交答案处理函数
func SubmitWenjuanAnswer(c *gin.Context) {
	// 从 JWT claims 获取用户邮箱
	claims, exists := c.Get("userEmail")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "请先登录",
		})
		return
	}

	userEmail, ok := claims.(string)
	if !ok || userEmail == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "无效的用户信息",
		})
		return
	}

	wenjuanId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的问卷ID",
		})
		return
	}

	var input struct {
		Answer string `json:"answer" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的答案格式",
			"error":   err.Error(),
		})
		return
	}

	// 验证答案格式
	var answers []string
	if err := json.Unmarshal([]byte(input.Answer), &answers); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "答案必须是JSON数组格式",
			"error":   err.Error(),
		})
		return
	}

	// 提交答案，传入用户邮箱
	if err := services.SubmitWenjuanAnswer(wenjuanId, input.Answer, userEmail); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "提交答案失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "提交成功",
	})
}

// 修改更新答案的处理器
func UpdateWenjuanAnswer(c *gin.Context) {
	wenjuanId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的问卷ID",
		})
		return
	}

	answerId, err := strconv.Atoi(c.Param("answerId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的答案ID",
		})
		return
	}

	var input struct {
		Answer string `json:"answer" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 验证答案格式
	var answers []string
	if err := json.Unmarshal([]byte(input.Answer), &answers); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "答案格式错误，应为JSON数组字符串",
			"error":   err.Error(),
		})
		return
	}

	updatedAnswer, err := services.UpdateWenjuanAnswer(wenjuanId, answerId, input.Answer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新答案失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "答案更新成功",
		"data": gin.H{
			"wenjuan_id": wenjuanId,
			"answer_id":  answerId,
			"answers":    answers,
			"updated_at": updatedAnswer.UpdatedAt,
		},
	})
}

// 删除答案
func DeleteWenjuanAnswer(c *gin.Context) {
	wenjuanId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的问卷ID",
		})
		return
	}

	answerId, err := strconv.Atoi(c.Param("answerId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的答案ID",
		})
		return
	}

	if err := services.DeleteWenjuanAnswer(wenjuanId, answerId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除答案失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "答案删除成功",
	})
}

// 获取问卷答案
func GetWenjuanAnswer(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的问卷ID"})
		return
	}

	answerId, err := strconv.Atoi(c.Param("answerId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的答案ID"})
		return
	}

	answer, err := services.GetWenjuanAnswer(id, answerId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, answer)
}

// 置顶
func PinWenjuan(c *gin.Context) {
	wenjuanId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的问卷ID"})
		return
	}

	if err := services.PinWenjuan(wenjuanId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "置顶成功"})
}

// 取消置顶
func UnpinWenjuan(c *gin.Context) {
	wenjuanId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的问卷ID"})
		return
	}

	if err := services.UnpinWenjuan(wenjuanId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "取消置顶成功"})
}

// 获取所有分类
func GetAllCategories(c *gin.Context) {
	result, err := services.GetAllCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取分类列表失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// 修改创建分类处理函数
func CreateCategory(c *gin.Context) {
	var input struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)

	if input.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "分类名称不能为空",
		})
		return
	}

	err := services.CreateCategory(input.Name, input.Description)
	if err != nil {
		if strings.Contains(err.Error(), "已存在") {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建分类失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "分类创建成功",
		"data": gin.H{
			"name":        input.Name,
			"description": input.Description,
		},
	})
}

// 更新分类
func UpdateCategory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的分类ID"})
		return
	}

	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if input.Name != "" {
		updates["name"] = input.Name
	}
	if input.Description != "" {
		updates["description"] = input.Description
	}

	if err := services.UpdateCategory(id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	category, err := services.GetCategoryById(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, category)
}

// 删除分类
func DeleteCategory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的分类ID"})
		return
	}

	if err := services.DeleteCategory(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "分类删除成功"})
}

// 添加分类到问卷
func AddCategoryToWenjuan(c *gin.Context) {
	wenjuanId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的问卷ID"})
		return
	}

	categoryId, err := strconv.Atoi(c.Param("categoryId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的分类ID"})
		return
	}

	if err := services.AddCategoryToWenjuan(wenjuanId, categoryId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "分类添加成功"})
}

// 下载问卷和答案
// DownloadWenjuanAndAnswers 处理下载问卷及答案的 HTTP 请求
func DownloadWenjuanAndAnswers(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的问卷ID"})
		return
	}

	pdfData, err := services.ExportWenjuanAsPDF(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=wenjuan_%d.pdf", id))
	c.Data(http.StatusOK, "application/pdf", pdfData)
}

// 添加错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":  500,
				"error": err.Error(),
			})
		}
	}
}

// 根据标题搜索问卷
func SearchWenjuanByTitle(c *gin.Context) {
	title := c.Query("title")
	// 添加分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	// 限制页面大小，防止请求过大数据
	if pageSize > 100 {
		pageSize = 100
	}

	result, err := services.SearchWenjuanByTitle(title, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "搜索失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": result,
	})
}

func ExportWenjuanAnswers(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的问卷ID"})
		return
	}

	csvData, err := services.ExportWenjuanAnswersAsCSV(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=wenjuan_%d_answers.csv", id))
	c.Data(http.StatusOK, "text/csv", csvData)
}
