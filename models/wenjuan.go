package models

import (
	"time"

	"gorm.io/gorm"
)

// Wenjuan 表示问卷模型
type Wenjuan struct {
	ID         uint            `gorm:"primarykey" json:"id"`                            // 问卷的主键ID
	Title      string          `gorm:"not null;type:longtext" json:"title"`             // 修改为longtext类型
	Content    string          `gorm:"type:text;not null" json:"content"`               // 修改为text类型
	Deadline   *time.Time      `json:"deadline"`                                        // 问卷的截止时间
	IsPinned   bool            `gorm:"column:is_pinned;default:false" json:"is_pinned"` // 问卷是否置顶 (Explicit column name)
	Status     string          `gorm:"not null;default:'draft'" json:"status"`          // 问卷的状态，默认是草稿状态
	CreatedAt  time.Time       `gorm:"autoCreateTime" json:"created_at"`                // 问卷的创建时间
	UpdatedAt  time.Time       `gorm:"autoUpdateTime" json:"updated_at"`                // 问卷的更新时间
	DeletedAt  gorm.DeletedAt  `gorm:"index" json:"-"`                                  // 添加软删除字段
	UserEmail  string          `gorm:"not null;type:longtext" json:"user_email"`        // 添加用户邮箱字段
	CreatorID  uint            `json:"creator_id"`                                      // 问卷的创建者ID
	Answers    []WenjuanAnswer `gorm:"foreignKey:WenjuanID" json:"answers,omitempty"`   // 添加 Answers 关联字段
	Categories []Category      `gorm:"many2many:wenjuan_categories;" json:"categories,omitempty"`
}

// WenjuanAnswer 表示问卷答案模型
type WenjuanAnswer struct {
	gorm.Model
	WenjuanID uint           `json:"wenjuan_id"`
	UserEmail string         `json:"user_email" gorm:"type:varchar(255);not null;default:'anonymous@example.com'"` // 添加默认值
	Answer    string         `gorm:"type:text;not null" json:"answer"`                                             // 存储JSON格式的答案字符串
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (WenjuanAnswer) TableName() string {
	return "wenjuan_answers"
}

// Category 表示问卷分类模型
type Category struct {
	ID          uint           `gorm:"primarykey" json:"id"`                                                      // 修改为小写id
	Name        string         `gorm:"type:varchar(255);not null;unique" json:"name"`                             // 添加unique约束
	Description string         `gorm:"type:text" json:"description"`                                              // 描述可以为空
	UserEmail   string         `gorm:"type:varchar(255);not null;default:'system@example.com'" json:"user_email"` // 添加默认值
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Wenjuans    []Wenjuan      `gorm:"many2many:wenjuan_categories;" json:"wenjuans,omitempty"`
}

// TableName 指定表名
func (Category) TableName() string {
	return "categories"
}

// WenjuanCategory 表示问卷和分类的中间表模型
type WenjuanCategory struct {
	WenjuanID  uint           `gorm:"primaryKey;constraint:OnDelete:CASCADE" json:"wenjuan_id"`  // 关联的问卷ID
	CategoryID uint           `gorm:"primaryKey;constraint:OnDelete:CASCADE" json:"category_id"` // 关联的分类ID
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`                          // 创建时间
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`                          // 更新时间
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`                                            // 添加软删除字段
}

// TableName 指定表名
func (WenjuanCategory) TableName() string {
	return "wenjuan_categories"
}
