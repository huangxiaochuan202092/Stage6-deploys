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
	db.AutoMigrate(&models.User{},
		&models.Task{},
		&models.Blog{},
		// &models.Wenjuan{}, &models.WenjuanAnswer{},
	)

}

func GetDB() *gorm.DB {
	if DB == nil {
		InitDB()
	}
	log.Println("mysql connected")
	return DB
}
