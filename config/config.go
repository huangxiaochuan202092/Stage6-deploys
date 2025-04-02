package config

import (
	"log"
	"proapp/models"

	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() {
	db, err := gorm.Open(mysql.Open("root:12345678@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		panic("failed to get database instance")
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Starting database migration...")

	// 设置日志级别
	logLevel := logger.Silent
	if gin.Mode() != gin.ReleaseMode {
		logLevel = logger.Info
	}

	db.Config.Logger = logger.Default.LogMode(logLevel)

	// 设置MySQL表选项
	db = db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci")

	// 确保表存在
	err = db.Exec(`CREATE DATABASE IF NOT EXISTS test`).Error
	if err != nil {
		log.Printf("Create database failed: %v", err)
		panic(err)
	}

	// 执行迁移
	err = db.AutoMigrate(
		&models.User{},
		&models.Blog{},
		&models.Task{},
		&models.Wenjuan{},
		&models.WenjuanAnswer{},
		&models.Category{},
		&models.WenjuanCategory{},
	)
	if err != nil {
		log.Printf("数据库迁移失败: %v", err)
		panic(err)
	}

	// 添加日志确认表创建
	var count int64
	if err := db.Model(&models.Category{}).Count(&count).Error; err != nil {
		log.Printf("检查分类表失败: %v", err)
	} else {
		log.Printf("分类表存在，当前记录数: %d", count)
	}

	log.Println("Database migration completed successfully")
	DB = db
}

func GetDB() *gorm.DB {
	if DB == nil {
		InitDB()
	}
	log.Println("mysql connected")
	return DB
}
