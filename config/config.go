package config

import (
	"log"
	"proapp/models"

	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	db, err := gorm.Open(mysql.Open("root:12345678@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	sqlDB, err := db.DB()
	if err != nil {
		panic("failed to connect database")
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	DB = db
	db.AutoMigrate(&models.User{}, //用户表
		&models.Task{},                             //任务表
		&models.Blog{},                             //博客
		&models.Wenjuan{}, &models.WenjuanAnswer{}, //答案和问卷
		&models.Category{},        //分类表
		&models.WenjuanCategory{}, //中间表
	)

}

func GetDB() *gorm.DB {
	if DB == nil {
		InitDB()
	}
	log.Println("mysql connected")
	return DB
}
