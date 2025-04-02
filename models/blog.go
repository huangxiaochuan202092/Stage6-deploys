package models

import (
	"time"

	"gorm.io/gorm"
)

// Blog 博客模型
type Blog struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	Title     string         `json:"title" gorm:"size:100;not null"`
	Content   string         `json:"content" gorm:"type:text"`
	Category  string         `json:"category" gorm:"size:50"`
	Tags      string         `json:"tags" gorm:"size:200"`
	Status    string         `json:"status" gorm:"default:draft"` // draft, published
	Likes     int            `json:"likes" gorm:"default:0"`
	UserID    uint           `json:"user_id" gorm:"not null"`
	UserEmail string         `json:"user_email" gorm:"size:100;not null"` // 添加用户邮箱字段
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (Blog) TableName() string {
	return "blogs"
}
