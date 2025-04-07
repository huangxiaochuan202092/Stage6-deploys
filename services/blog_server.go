package services

import (
	"errors"
	"fmt"
	"log"
	"proapp/config"
	"proapp/models"
	"time"

	"gorm.io/gorm" // Import gorm package
)

// 获取所有博客
func GetAllBlogs(page, pageSize int) ([]models.Blog, int64, error) {
	return GetPaginatedBlogs(page, pageSize, "")
}

// 根据ID获取博客 - 使用明确的SQL查询，指定表名
func GetBlogById(id int) (*models.Blog, error) {
	var blog models.Blog

	// 使用明确的SQL查询，指定表名
	sql := "SELECT * FROM blogs WHERE id = ? AND deleted_at IS NULL"
	result := config.DB.Raw(sql, id).Scan(&blog)

	if result.Error != nil {
		log.Printf("根据ID获取博客失败: %v", result.Error)
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, nil // 博客不存在
	}

	return &blog, nil
}

// 创建博客 - 增强错误处理，使之更健壮
func CreateBlog(blog *models.Blog) error {
	if blog == nil {
		return errors.New("博客对象不能为空")
	}

	blog.CreatedAt = time.Now()
	blog.UpdatedAt = time.Now()

	// 验证用户ID
	if blog.UserID == 0 {
		return errors.New("用户ID不能为空")
	}

	// 获取用户邮箱
	var userEmail string
	err := config.DB.Raw("SELECT email FROM users WHERE id = ? AND deleted_at IS NULL", blog.UserID).Scan(&userEmail).Error
	if err != nil {
		log.Printf("获取用户邮箱失败: %v", err)
		return fmt.Errorf("获取用户邮箱失败: %w", err)
	}

	if userEmail == "" {
		return errors.New("找不到有效的用户邮箱")
	}

	// 设置博客的用户邮箱
	blog.UserEmail = userEmail

	// 验证必填字段
	if blog.Title == "" {
		return errors.New("博客标题不能为空")
	}

	if blog.Content == "" {
		return errors.New("博客内容不能为空")
	}

	// 使用原始SQL插入数据
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
		blog.UserEmail,
		blog.CreatedAt,
		blog.UpdatedAt)

	if result.Error != nil {
		log.Printf("创建博客失败: %v", result.Error)
		return fmt.Errorf("创建博客失败: %w", result.Error)
	}

	// 获取插入的ID
	var lastID int
	err = config.DB.Raw("SELECT LAST_INSERT_ID()").Scan(&lastID).Error
	if err == nil && lastID > 0 {
		blog.ID = uint(lastID)
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

// 增强点赞功能

// 点赞博客 - 确保正确更新点赞数
func LikeBlog(id int) (int, error) {
	// 验证博客是否存在
	blog, err := GetBlogById(id)
	if err != nil {
		log.Printf("获取博客失败: %v", err)
		return 0, err
	}

	if blog == nil {
		return 0, errors.New("博客不存在")
	}

	// 增加点赞数
	sql := "UPDATE blogs SET likes = likes + 1 WHERE id = ? AND deleted_at IS NULL"
	result := config.DB.Exec(sql, id)
	if result.Error != nil {
		log.Printf("点赞博客失败: %v", result.Error)
		return 0, result.Error
	}

	// 查询更新后的点赞数
	var newLikes int
	sql = "SELECT likes FROM blogs WHERE id = ? AND deleted_at IS NULL"
	err = config.DB.Raw(sql, id).Scan(&newLikes).Error
	if err != nil {
		log.Printf("获取更新后的点赞数失败: %v", err)
		return 0, err
	}

	log.Printf("博客ID=%d 点赞成功，新的点赞数: %d", id, newLikes)
	return newLikes, nil
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
func SearchBlogs(keyword string, page, pageSize int) ([]models.Blog, int64, error) {
	return GetPaginatedBlogs(page, pageSize, keyword)
}

// GetPaginatedBlogs 获取分页博客列表
func GetPaginatedBlogs(page, pageSize int, keyword string) ([]models.Blog, int64, error) {
	var blogs []models.Blog
	var total int64

	// 明确使用blogs表，避免与tasks表混淆
	tableName := "blogs"

	// 1. 获取总数
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE deleted_at IS NULL", tableName)
	if keyword != "" {
		countSQL += " AND title LIKE ?"
		if err := config.DB.Raw(countSQL, "%"+keyword+"%").Count(&total).Error; err != nil {
			log.Printf("获取博客总数失败: %v", err)
			return nil, 0, err
		}
	} else {
		if err := config.DB.Raw(countSQL).Count(&total).Error; err != nil {
			log.Printf("获取博客总数失败: %v", err)
			return nil, 0, err
		}
	}

	// 2. 如果总数为0，直接返回空结果
	if total == 0 {
		log.Printf("找到0条匹配的博客记录 (关键词: %s)", keyword)
		return []models.Blog{}, 0, nil
	}

	// 3. 获取分页数据
	offset := (page - 1) * pageSize

	// 明确使用SQL查询，避免GORM模型混淆
	dataSQL := fmt.Sprintf("SELECT * FROM %s WHERE deleted_at IS NULL", tableName)
	if keyword != "" {
		dataSQL += " AND title LIKE ?"
		dataSQL += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
		if err := config.DB.Raw(dataSQL, "%"+keyword+"%", pageSize, offset).Scan(&blogs).Error; err != nil {
			log.Printf("获取分页博客数据失败: %v", err)
			return nil, 0, err
		}
	} else {
		dataSQL += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
		if err := config.DB.Raw(dataSQL, pageSize, offset).Scan(&blogs).Error; err != nil {
			log.Printf("获取分页博客数据失败: %v", err)
			return nil, 0, err
		}
	}

	// 4. 查询成功日志
	log.Printf("成功获取分页博客 (关键词: %s, 页码: %d, 大小: %d)，返回 %d 条记录",
		keyword, page, pageSize, len(blogs))

	return blogs, total, nil
}

// BlogServer 结构体定义
type BlogServer struct {
	db *gorm.DB
}

// NewBlogServer 创建一个新的博客服务实例
func NewBlogServer() *BlogServer {
	return &BlogServer{
		db: config.DB,
	}
}

// 修改博客列表获取函数，修复类型错误
func GetBlogsWithPagination(page, pageSize int, keyword string) (*models.BlogResponse, error) {
	// 获取总数
	var total int64 // 修改为int64类型
	query := config.DB.Model(&models.Blog{})

	if keyword != "" {
		query = query.Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("获取博客总数失败: %w", err)
	}

	// 获取分页数据
	var blogs []models.Blog
	offset := (page - 1) * pageSize

	log.Printf("获取所有博客，页码：%d，每页数量：%d", page, pageSize)

	query = config.DB.Model(&models.Blog{})
	if keyword != "" {
		query = query.Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&blogs).Error; err != nil {
		log.Printf("获取博客列表失败: %v", err)
		return nil, fmt.Errorf("获取博客列表失败: %w", err)
	}

	return &models.BlogResponse{
		Total: int(total),
		Blogs: blogs,
	}, nil
}
