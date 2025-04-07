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
	Status    string         `json:"status" gorm:"default:draft"`
	Likes     int            `json:"likes" gorm:"default:0"`
	UserID    uint           `json:"user_id" gorm:"column:user_id"`
	UserEmail string         `json:"user_email" gorm:"column:user_email"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// BlogResponse 定义博客列表响应结构
type BlogResponse struct {
	Total int    `json:"total"`
	Blogs []Blog `json:"blogs"`
}

// TableName 显式指定表名
func (Blog) TableName() string {
	return "blogs"
}
