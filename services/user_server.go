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
	log.Printf("开始查询邮箱为 %s 的用户", email)

	// 检查数据库连接
	if config.DB == nil {
		log.Printf("严重错误: 数据库连接未初始化")
		return nil, errors.New("数据库连接未初始化")
	}

	var user models.User
	// 使用明确的表名和字段，避免意外的表连接
	result := config.DB.Unscoped().Table("users").Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			log.Printf("未找到邮箱为 %s 的用户", email)
			return nil, nil
		}
		log.Printf("查询用户时发生数据库错误: %v", result.Error)
		return nil, result.Error
	}

	log.Printf("成功获取到用户: ID=%d, Email=%s", user.ID, user.Email)
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

	// 开启事务
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("发生panic，事务回滚: %v", r)
		}
	}()

	// 检查邮箱是否已存在（包括软删除的记录）
	var user models.User
	// 使用明确的表名查询，避免意外连接
	if err := tx.Unscoped().Table("users").Where("email = ?", email).First(&user).Error; err == nil {
		tx.Rollback()
		log.Printf("邮箱 %s 已存在", email)
		return nil, errors.New("邮箱已存在")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		log.Printf("检查邮箱是否存在时发生错误: %v", err)
		return nil, err
	}

	// 创建新用户
	newUser := models.User{Email: email, Role: "user"} // 确保设置默认角色
	if err := tx.Create(&newUser).Error; err != nil {
		tx.Rollback()
		log.Printf("创建用户失败: %v", err)
		return nil, err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		log.Printf("提交事务失败: %v", err)
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
	result := config.DB.Unscoped().Delete(&models.User{}, userID)
	return result.Error
}

func UpdateUserEmail(userID uint, newEmail string) error {
	result := config.DB.Model(&models.User{}).Where("id = ?", userID).Update("email", newEmail)
	return result.Error
}
