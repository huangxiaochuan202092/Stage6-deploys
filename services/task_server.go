package services

import (
	"errors"
	"log"
	"proapp/config"
	"proapp/models"
	"time"

	"gorm.io/gorm"
)

// 创建任务
func CreateTask(task *models.Task) error {
	// 确保任务表字段正确
	if task.Title == "" {
		return errors.New("任务标题不能为空")
	}

	// 设置默认值
	if task.Priority == "" {
		task.Priority = "medium"
	}
	if task.Status == "" {
		task.Status = "pending"
	}

	// 明确设置创建时间和更新时间
	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now

	// 修复：使用纯SQL查询获取用户邮箱，完全避免ORM状态污染
	var userEmail string
	err := config.DB.Raw("SELECT email FROM users WHERE id = ? AND deleted_at IS NULL", task.CreatorID).Scan(&userEmail).Error
	if err != nil {
		log.Printf("获取用户邮箱失败: %v", err)
		return err
	}

	// 将邮箱赋值给任务
	task.UserEmail = userEmail

	// 日志记录插入的数据，便于调试
	log.Printf("准备创建任务，数据: title=%s, desc=%s, priority=%s, status=%s, creator_id=%d, email=%s",
		task.Title, task.Description, task.Priority, task.Status, task.CreatorID, task.UserEmail)

	// 使用原始SQL查询避免ORM字段映射问题
	sql := `INSERT INTO tasks (
		title, description, priority, status, deadline, 
		creator_id, user_email, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result := config.DB.Exec(sql,
		task.Title,
		task.Description,
		task.Priority,
		task.Status,
		task.Deadline,
		task.CreatorID,
		task.UserEmail,
		task.CreatedAt,
		task.UpdatedAt)

	if result.Error != nil {
		log.Printf("创建任务失败: %v", result.Error)
		return result.Error
	}

	return nil
}

// 获取所有任务
func GetAllTasks() ([]models.Task, error) {
	var tasks []models.Task

	// 使用原始SQL查询来避免ORM的自动关联和复杂逻辑
	result := config.DB.Raw("SELECT * FROM tasks WHERE deleted_at IS NULL ORDER BY created_at DESC").Scan(&tasks)

	if result.Error != nil {
		log.Printf("获取任务失败: %v", result.Error)
		return nil, result.Error
	}

	return tasks, nil
}

// 根据用户ID获取任务列表
func GetTasksByUserID(userID uint) ([]models.Task, error) {
	var tasks []models.Task

	// 明确使用tasks表，并添加用户ID筛选
	result := config.DB.Table("tasks").Where("creator_id = ?", userID).Order("created_at DESC").Find(&tasks)

	if result.Error != nil {
		log.Printf("获取用户任务失败: %v", result.Error)
		return nil, result.Error
	}

	return tasks, nil
}

// 获取单个任务
func GetTaskById(id int) (*models.Task, error) {
	var task models.Task

	// 使用原始SQL查询或明确指定查询字段，避免ORM自动映射错误
	result := config.DB.Raw("SELECT * FROM tasks WHERE id = ? AND deleted_at IS NULL LIMIT 1", id).Scan(&task)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		log.Printf("根据ID获取任务失败: %v", result.Error)
		return nil, result.Error
	}

	// 如果没有找到记录
	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &task, nil
}

// 更新任务
func UpdateTask(id int, updates map[string]interface{}) error {
	// 确保更新时间
	updates["updated_at"] = time.Now()

	// 明确使用tasks表，避免错误地更新categories表
	result := config.DB.Table("tasks").Where("id = ?", id).Updates(updates)

	if result.Error != nil {
		log.Printf("更新任务失败: %v", result.Error)
		return result.Error
	}

	return nil
}

// 删除任务
func DeleteTask(id int) error {
	// 使用原始SQL执行软删除，避免GORM软删除机制的问题
	sql := "UPDATE tasks SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL"
	result := config.DB.Exec(sql, time.Now(), id)

	if result.Error != nil {
		log.Printf("删除任务失败: %v", result.Error)
		return result.Error
	}

	// 如果没有行受影响（可能任务不存在或已被删除）
	if result.RowsAffected == 0 {
		log.Printf("未找到ID为%d的任务或任务已被删除", id)
		// 不返回错误，因为目标状态（任务被删除）已经达成
	}

	return nil
}

// 获取分页任务列表
func GetTasksWithPagination(page, pageSize int) ([]models.Task, int64, error) {
	var tasks []models.Task
	var total int64

	// 获取总记录数
	if err := config.DB.Table("tasks").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据，按创建时间倒序排序
	result := config.DB.Table("tasks").Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&tasks)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return tasks, total, nil
}

// 搜索任务
func SearchTasks(keyword string) ([]models.Task, error) {
	var tasks []models.Task

	// 使用新的DB会话避免状态污染
	result := config.DB.Session(&gorm.Session{NewDB: true}).
		Table("tasks").
		Where("title LIKE ?", "%"+keyword+"%").
		Order("created_at DESC").
		Find(&tasks)

	if result.Error != nil {
		return nil, result.Error
	}
	return tasks, nil
}
