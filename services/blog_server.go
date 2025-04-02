package services

import (
	"errors"
	"log"
	"proapp/config"
	"proapp/models"
	"time"
)

// 获取所有博客
func GetAllBlogs() ([]models.Blog, error) {
	var blogs []models.Blog

	// 使用模型结构体对应的确切列名，包含user_email字段
	query := `SELECT id, title, content, category, tags, status, likes, user_id, user_email,
	         created_at, updated_at, deleted_at FROM blogs 
	         WHERE deleted_at IS NULL ORDER BY created_at DESC`
	result := config.DB.Raw(query).Scan(&blogs)

	if result.Error != nil {
		log.Printf("获取博客失败: %v", result.Error)
		return nil, result.Error
	}

	return blogs, nil
}

// 根据ID获取博客
func GetBlogById(id int) (*models.Blog, error) {
	var blog models.Blog

	// 明确列出所有需要的字段，包含user_email字段
	query := `SELECT id, title, content, category, tags, status, likes, user_id, user_email,
	         created_at, updated_at, deleted_at FROM blogs 
	         WHERE id = ? AND deleted_at IS NULL`
	result := config.DB.Raw(query, id).Scan(&blog)

	if result.Error != nil {
		log.Printf("根据ID获取博客失败: %v", result.Error)
		return nil, result.Error
	}

	// 如果没有找到记录
	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &blog, nil
}

// 创建博客
func CreateBlog(blog *models.Blog) error {
	blog.CreatedAt = time.Now()
	blog.UpdatedAt = time.Now()

	// 获取用户邮箱
	var userEmail string
	err := config.DB.Raw("SELECT email FROM users WHERE id = ? AND deleted_at IS NULL", blog.UserID).Scan(&userEmail).Error
	if err != nil {
		log.Printf("获取用户邮箱失败: %v", err)
		return err
	}

	// 设置博客的用户邮箱
	blog.UserEmail = userEmail

	// 使用SQL插入语句确保字段名匹配，包含user_email字段
	sql := `INSERT INTO blogs (
		title, content, category, tags, status, likes, user_id, user_email, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result := config.DB.Exec(sql,
		blog.Title,
		blog.Content,
		blog.Category,
		blog.Tags,
		blog.Status,
		blog.Likes,
		blog.UserID,
		blog.UserEmail, // 包含UserEmail值
		blog.CreatedAt,
		blog.UpdatedAt)

	if result.Error != nil {
		log.Printf("创建博客失败: %v", result.Error)
		return result.Error
	}

	return nil
}

// 更新博客
func UpdateBlog(id int, updates map[string]interface{}) error {
	// 确保更新时间
	updates["updated_at"] = time.Now()

	// 记录要更新的字段，便于调试
	log.Printf("更新博客 ID=%d，字段：%+v", id, updates)

	// 使用原始SQL查询，避免ORM映射问题
	query := "UPDATE blogs SET "
	values := []interface{}{}

	for field, value := range updates {
		query += field + " = ?, "
		values = append(values, value)
	}

	// 移除最后的逗号和空格
	query = query[:len(query)-2]

	// 添加WHERE条件
	query += " WHERE id = ? AND deleted_at IS NULL"
	values = append(values, id)

	result := config.DB.Exec(query, values...)

	if result.Error != nil {
		log.Printf("更新博客失败: %v", result.Error)
		return result.Error
	}

	// 记录更新结果
	log.Printf("博客更新结果: 影响行数=%d", result.RowsAffected)

	// 如果没有更新任何行，可能是因为博客不存在或已被删除
	if result.RowsAffected == 0 {
		return errors.New("未找到博客或博客已被删除")
	}

	return nil
}

// 删除博客
func DeleteBlog(id int) error {
	// 执行软删除
	sql := "UPDATE blogs SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL"
	result := config.DB.Exec(sql, time.Now(), id)

	if result.Error != nil {
		log.Printf("删除博客失败: %v", result.Error)
		return result.Error
	}

	return nil
}

// 点赞博客
func LikeBlog(id int) error {
	// 增加点赞数
	result := config.DB.Exec("UPDATE blogs SET likes = likes + 1 WHERE id = ?", id)

	if result.Error != nil {
		log.Printf("点赞博客失败: %v", result.Error)
		return result.Error
	}

	return nil
}

// 取消点赞
func DislikeBlog(id int) error {
	// 减少点赞数，但确保不会小于0
	result := config.DB.Exec("UPDATE blogs SET likes = GREATEST(likes - 1, 0) WHERE id = ?", id)

	if result.Error != nil {
		log.Printf("取消点赞失败: %v", result.Error)
		return result.Error
	}

	return nil
}

// 搜索博客
func SearchBlogs(keyword string) ([]models.Blog, error) {
	var blogs []models.Blog

	// 使用原始SQL查询，避免ORM映射问题，包含user_email字段
	query := `SELECT id, title, content, category, tags, status, likes, user_id, user_email,
	         created_at, updated_at, deleted_at FROM blogs 
	         WHERE deleted_at IS NULL AND title LIKE ? ORDER BY created_at DESC`

	result := config.DB.Raw(query, "%"+keyword+"%").Scan(&blogs)

	if result.Error != nil {
		log.Printf("搜索博客失败: %v", result.Error)
		return nil, result.Error
	}

	return blogs, nil
}
