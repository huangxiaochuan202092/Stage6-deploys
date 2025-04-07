package config

import (
	"fmt"
	"log"
	"os"
	"proapp/models"

	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() *gorm.DB {
	// 从环境变量获取数据库连接参数
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// 构建连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName)

	log.Printf("Using DSN: %s:%s@tcp(%s:%s)/%s", dbUser, "****", dbHost, dbPort, dbName)

	// 使用连接字符串连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
		return nil
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

	// 执行迁移前打印SQL
	db.Logger = db.Logger.LogMode(logger.Info)

	// 使用事务进行迁移，以便在出错时回滚
	err = db.Transaction(func(tx *gorm.DB) error {
		// 迁移表结构
		if err := tx.AutoMigrate(
			&models.User{},
			&models.Blog{},
			&models.Task{}, // 确保Task模型被迁移
		); err != nil {
			log.Printf("数据库迁移失败: %v", err)
			return err
		}

		return nil
	})

	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	log.Println("Database migration completed successfully")
	DB = db
	return db
}

// 修改 GetDB 函数，加入连接健康检查
func GetDB() *gorm.DB {
	if DB == nil {
		InitDB()
	} else {
		// 检查数据库连接是否健康
		sqlDB, err := DB.DB()
		if err != nil {
			log.Printf("获取SQL DB实例失败: %v，尝试重新初始化连接", err)
			InitDB()
		} else if err := sqlDB.Ping(); err != nil {
			log.Printf("数据库连接Ping失败: %v，尝试重新初始化连接", err)
			// 尝试关闭现有连接
			sqlDB.Close()
			InitDB()
		}
	}

	log.Println("数据库连接就绪")
	return DB
}

// 改进 ResetDBConnection 函数

// 增强数据库连接重置功能
func ResetDBConnection() {
	log.Println("尝试重置数据库连接...")

	if DB != nil {
		sqlDB, err := DB.DB()
		if err == nil {
			sqlDB.Close()
		}
	}

	// 强制重新初始化数据库连接
	DB = nil

	// 使用新的初始化代码，确保表迁移

	db := InitDB()

	// 验证连接是否成功
	if db != nil {
		sqlDB, err := db.DB()
		if err == nil {
			err = sqlDB.Ping()
			if err == nil {
				log.Println("数据库连接已成功重置")
				return
			}
			log.Printf("数据库Ping失败: %v", err)
		} else {
			log.Printf("获取SQL DB实例失败: %v", err)
		}
	}

	log.Println("数据库连接重置失败，使用备用方法")

	// 备用方法：直接使用连接字符串初始化
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		"root", "12345678", "localhost", "3306", "test")

	newDb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("备用方法初始化数据库失败: %v", err)
		return
	}

	DB = newDb
	log.Println("数据库连接已重置(备用方法)")
}
