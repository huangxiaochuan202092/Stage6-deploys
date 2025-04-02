package models

import (
	"time"

	"gorm.io/gorm"
)

// Task 任务模型
type Task struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	Title       string         `json:"title" gorm:"size:100;not null"`
	Description string         `json:"description" gorm:"type:text"`
	Priority    string         `json:"priority" gorm:"default:medium"` // high, medium, low
	Status      string         `json:"status" gorm:"default:pending"`  // pending, in_progress, completed
	Deadline    *time.Time     `json:"deadline"`
	CreatorID   uint           `json:"creator_id"`
	UserEmail   string         `json:"user_email" gorm:"not null"` // 添加用户邮箱字段
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (Task) TableName() string {
	return "tasks"
}
