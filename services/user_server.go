package services

import (
	"errors"
	"log"
	"proapp/config"
	"proapp/models"

	"gorm.io/gorm"
)

// 根据邮箱获取用户
func GetUserByEmail(email string) (*models.User, error) {
	var user models.User

	// 使用First前检查DB是否已初始化
	if config.DB == nil {
		return nil, errors.New("数据库连接未初始化")
	}

	// 使用新会话确保查询不受之前查询的影响
	db := config.DB.Session(&gorm.Session{NewDB: true})

	// 使用Find代替First，因为First在找不到记录时会返回error
	result := db.Unscoped().Where("email = ?", email).Find(&user)

	if result.Error != nil {
		// 如果错误是 "记录未找到"，则返回 nil, nil
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // 用户不存在，不是错误
		}
		// 对于其他数据库错误，记录日志并返回错误
		log.Printf("数据库查询错误: %v", result.Error)
		return nil, result.Error
	}

	// 检查是否找到记录
	if result.RowsAffected == 0 {
		return nil, nil // 返回nil,nil表示未找到用户但没有错误
	}

	return &user, nil
}

// 创建用户
func CreateUser(email string) (*models.User, error) {
	log.Printf("开始创建用户: email=%s", email)

	// 检查数据库连接
	if config.DB == nil {
		log.Printf("严重错误: 数据库连接未初始化")
		return nil, errors.New("数据库连接未初始化")
	}

	// 创建新会话避免查询条件累积
	db := config.DB.Session(&gorm.Session{NewDB: true})

	// 检查邮箱是否已存在，使用新的会话
	var existingUser models.User
	result := db.Unscoped().Where("email = ?", email).First(&existingUser)
	if result.Error == nil {
		// 用户已存在
		log.Printf("邮箱 %s 已存在", email)
		return &existingUser, nil // 返回已存在用户，而不是报错
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// 发生了非"记录不存在"的错误
		log.Printf("检查邮箱是否存在时发生错误: %v", result.Error)
		return nil, result.Error
	}

	// 创建新用户，使用新的会话防止条件累积
	newUser := models.User{
		Email: email,
		Role:  "user", // 确保设置默认角色
	}

	if err := db.Create(&newUser).Error; err != nil {
		log.Printf("创建用户失败: %v", err)
		return nil, err
	}

	log.Printf("成功创建用户: ID=%d, Email=%s", newUser.ID, newUser.Email)
	return &newUser, nil
}

// 根据id获取用户
func GetUserByID(id uint) (*models.User, error) {
	var user models.User
	result := config.DB.Unscoped().Where("id = ?", id).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

// 获取所有用户
func GetAllUsers() ([]models.User, error) {
	log.Printf("开始获取所有用户数据")

	// 检查数据库连接
	if config.DB == nil {
		log.Printf("严重错误: 数据库连接未初始化")
		return nil, errors.New("数据库连接未初始化")
	}

	var users []models.User

	// 创建新的查询会话，避免之前的查询状态影响
	db := config.DB.Session(&gorm.Session{NewDB: true})

	// 明确指定表名为"users"，避免ORM映射错误，并按ID排序
	query := db.Unscoped().Table("users").Order("id")
	log.Printf("执行查询: %v", query.Statement.SQL.String())

	result := query.Find(&users)
	if result.Error != nil {
		// 如果错误是 "记录未找到"，则返回空切片和 nil 错误
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			log.Printf("未找到任何用户记录")
			return []models.User{}, nil // 返回空切片表示没有用户
		}
		// 对于其他数据库错误，记录日志并返回错误
		log.Printf("数据库查询错误: %v", result.Error)
		return nil, result.Error
	}

	log.Printf("成功查询到 %d 名用户", len(users))
	return users, nil
}

// 更新用户
func UpdateUser(userID uint, newEmail string) error {
	result := config.DB.Unscoped().Model(&models.User{}).Where("id = ?", userID).Update("email", newEmail)
	return result.Error
}

// 更新用户字段
func UpdateUserFields(userID uint, updates map[string]interface{}) error {
	// 使用新会话确保不影响其他查询
	db := config.DB.Session(&gorm.Session{NewDB: true})

	// 明确指定更新条件
	result := db.Unscoped().Model(&models.User{}).Where("id = ?", userID).Updates(updates)

	// 清除可能存在的全局查询状态
	config.DB = config.DB.Session(&gorm.Session{NewDB: true})

	return result.Error
}

// 删除用户
func DeleteUserByID(userID uint) error {
	// 使用软删除
	result := config.DB.Unscoped().Where("id = ?", userID).Delete(&models.User{})
	if result.Error != nil {
		return result.Error
	}

	// 检查是否删除成功
	if result.RowsAffected == 0 {
		return errors.New("未找到用户或用户已被删除")
	}

	return nil
}
