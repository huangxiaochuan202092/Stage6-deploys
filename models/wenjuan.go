package models

import (
	"time"

	"gorm.io/gorm"
)

// Wenjuan 表示问卷模型
type Wenjuan struct {
	ID         uint            `gorm:"primarykey" json:"id"`                                            // 问卷的主键ID
	Title      string          `gorm:"not null" json:"title"`                                           // 问卷的标题
	Content    string          `gorm:"type:text;not null" json:"content"`                               // 修改为text类型
	Deadline   *time.Time      `json:"deadline"`                                                        // 问卷的截止时间
	IsPinned   bool            `gorm:"default:false" json:"is_pinned"`                                  // 问卷是否置顶
	Status     string          `gorm:"not null;default:'draft'" json:"status"`                          // 问卷的状态，默认是草稿状态
	CreatedAt  time.Time       `gorm:"autoCreateTime" json:"created_at"`                                // 问卷的创建时间
	UpdatedAt  time.Time       `gorm:"autoUpdateTime" json:"updated_at"`                                // 问卷的更新时间
	Answers    []WenjuanAnswer `gorm:"foreignKey:WenjuanID;constraint:OnDelete:CASCADE" json:"answers"` // 问卷对应的答案，一对多关系
	Categories []Category      `gorm:"many2many:wenjuan_categories" json:"categories"`                  // 问卷对应的分类，多对多关系
	DeletedAt  gorm.DeletedAt  `gorm:"index" json:"-"`
}

// WenjuanAnswer 表示问卷答案模型
type WenjuanAnswer struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	WenjuanID uint           `gorm:"not null;index" json:"wenjuan_id"`
	Answer    string         `gorm:"type:text;not null" json:"answer"` // 存储JSON格式的答案字符串
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Category 表示问卷分类模型
type Category struct {
	ID          uint           `gorm:"primarykey" json:"id"`        // 修改为小写id
	Name        string         `gorm:"not null;unique" json:"name"` // 添加unique约束
	Description string         `json:"description"`                 // 描述可以为空
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Wenjuans    []Wenjuan      `gorm:"many2many:wenjuan_categories;" json:"wenjuans,omitempty"`
}

// WenjuanCategory 表示问卷和分类的中间表模型
type WenjuanCategory struct {
	WenjuanID  uint      `gorm:"primaryKey;constraint:OnDelete:CASCADE" json:"wenjuan_id"`  // 关联的问卷ID
	CategoryID uint      `gorm:"primaryKey;constraint:OnDelete:CASCADE" json:"category_id"` // 关联的分类ID
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`                          // 创建时间
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`                          // 更新时间
}
