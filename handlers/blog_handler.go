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
	// 获取搜索关键词
	keyword := c.Query("keyword")

	var blogs []models.Blog
	var err error

	if keyword != "" {
		// 如果有搜索关键词，调用搜索函数
		blogs, err = services.SearchBlogs(keyword)
		log.Printf("搜索博客，关键词：%s", keyword)
	} else {
		// 否则获取所有博客
		blogs, err = services.GetAllBlogs()
	}

	if err != nil {
		log.Printf("获取博客列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取博客列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"blogs":  blogs,
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
		"blog":   blog,
	})
}

// CreateBlog 创建博客
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
		UserID:   uint(userID.(float64)),
	}

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

	// 记录请求数据，便于调试
	log.Printf("收到博客更新请求: ID=%d, 标题=%s, 分类=%s, 标签=%s, 状态=%s",
		id, input.Title, input.Category, input.Tags, input.Status)

	// 构建更新数据，只包含非空字段
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
		log.Printf("更新请求未提供任何更新字段")
		c.JSON(http.StatusBadRequest, gin.H{"error": "没有提供更新数据"})
		return
	}

	// 执行更新
	if err := services.UpdateBlog(id, updates); err != nil {
		log.Printf("更新博客失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新博客失败: " + err.Error()})
		return
	}

	// 返回成功响应
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

// LikeBlog 点赞博客
func LikeBlog(c *gin.Context) {
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

	// 点赞
	if err := services.LikeBlog(id); err != nil {
		log.Printf("点赞博客失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "点赞失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "点赞成功",
	})
}

// DislikeBlog 取消点赞
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

	// 取消点赞
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
