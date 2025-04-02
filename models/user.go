package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	Email     string         `json:"email" gorm:"size:100;not null;unique"`
	Role      string         `json:"role" gorm:"default:user"` // 新增 Role 字段，默认为 user
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	Tasks     []Task         `json:"-" gorm:"foreignKey:CreatorID;references:ID"` // 修正外键关系定义
}
