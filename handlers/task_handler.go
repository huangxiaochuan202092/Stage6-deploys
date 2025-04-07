package handlers

import (
	"log"
	"net/http"
	"proapp/models"
	"proapp/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GetAllTasks 获取所有任务
func GetAllTasks(c *gin.Context) {
	// 获取搜索关键词
	keyword := c.Query("keyword")

	var tasks []models.Task
	var err error

	if keyword != "" {
		// 如果有搜索关键词，调用搜索函数
		tasks, err = services.SearchTasks(keyword)
	} else {
		// 否则获取所有任务
		tasks, err = services.GetAllTasks()
	}

	if err != nil {
		log.Printf("获取任务列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取任务列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"tasks":  tasks,
	})
}

// GetTask 获取单个任务
func GetTask(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	task, err := services.GetTaskById(id)
	if err != nil {
		log.Printf("获取任务详情失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取任务详情失败"})
		return
	}

	if task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"task":   task,
	})
}

// CreateTask 创建新任务
func CreateTask(c *gin.Context) {
	var input struct {
		Title       string    `json:"title" binding:"required"`
		Description string    `json:"description"`
		Priority    string    `json:"priority"`
		Status      string    `json:"status"`
		DueDate     time.Time `json:"due_date"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("绑定请求数据失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 记录收到的数据
	log.Printf("收到创建任务请求: 标题=%s, 描述=%s, 优先级=%s, 状态=%s, 截止日期=%v",
		input.Title, input.Description, input.Priority, input.Status, input.DueDate)

	// 确保标题非空
	if input.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "标题不能为空"})
		return
	}

	// 从JWT中获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无法获取用户信息"})
		return
	}

	// 验证优先级和状态
	if input.Priority == "" {
		input.Priority = "medium"
	} else if input.Priority != "high" && input.Priority != "medium" && input.Priority != "low" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的优先级，可选值：high, medium, low"})
		return
	}

	if input.Status == "" {
		input.Status = "pending"
	} else if input.Status != "pending" && input.Status != "in_progress" && input.Status != "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的状态，可选值：pending, in_progress, completed"})
		return
	}

	// 创建任务对象
	var deadline *time.Time
	if !input.DueDate.IsZero() {
		deadline = &input.DueDate
	}

	// 根据类型分别处理userID
	var creatorID uint
	switch v := userID.(type) {
	case float64:
		creatorID = uint(v)
	case uint:
		creatorID = v
	case int:
		creatorID = uint(v)
	default:
		log.Printf("userID类型错误: %T, 值: %v", userID, userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户ID类型无效"})
		return
	}

	task := &models.Task{
		Title:       input.Title,
		Description: input.Description,
		Priority:    input.Priority,
		Status:      input.Status,
		Deadline:    deadline,
		CreatorID:   creatorID,
	}

	if err := services.CreateTask(task); err != nil {
		log.Printf("创建任务失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建任务失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "任务创建成功",
		"task":    task,
	})
}

// UpdateTask 更新任务
func UpdateTask(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	// 检查任务是否存在
	task, err := services.GetTaskById(id)
	if err != nil {
		log.Printf("获取任务失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取任务失败"})
		return
	}
	if task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	var input struct {
		Title       string    `json:"title"`
		Description string    `json:"description"`
		Priority    string    `json:"priority"`
		Status      string    `json:"status"`
		DueDate     time.Time `json:"due_date"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 构建更新数据
	updates := make(map[string]interface{})

	if input.Title != "" {
		updates["title"] = input.Title
	}
	if input.Description != "" {
		updates["description"] = input.Description
	}
	if input.Priority != "" {
		if input.Priority != "high" && input.Priority != "medium" && input.Priority != "low" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的优先级，可选值：high, medium, low"})
			return
		}
		updates["priority"] = input.Priority
	}
	if input.Status != "" {
		if input.Status != "pending" && input.Status != "in_progress" && input.Status != "completed" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的状态，可选值：pending, in_progress, completed"})
			return
		}
		updates["status"] = input.Status
	}
	if !input.DueDate.IsZero() {
		updates["deadline"] = input.DueDate
	}

	// 如果没有数据更新
	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "没有提供更新数据"})
		return
	}

	// 执行更新
	if err := services.UpdateTask(id, updates); err != nil {
		log.Printf("更新任务失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新任务失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "任务更新成功",
	})
}

// DeleteTask 删除任务
func DeleteTask(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	// 检查任务是否存在
	task, err := services.GetTaskById(id)
	if err != nil {
		log.Printf("获取任务失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取任务失败"})
		return
	}
	if task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	// 执行删除
	if err := services.DeleteTask(id); err != nil {
		log.Printf("删除任务失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除任务失败：" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "任务删除成功",
	})
}
