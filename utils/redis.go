package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func InitRedis() {
	// 从环境变量获取Redis连接参数
	redisHost := os.Getenv("REDIS_HOST")
	log.Printf("连接Redis使用主机: %s", redisHost)
	// 确保有默认值
	if redisHost == "" {
		redisHost = "host.docker.internal"
	}

	redisPort := os.Getenv("REDIS_PORT")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	// 密码可以为空，不需要默认值

	redisDB := 0 // 默认值
	redisDBStr := os.Getenv("REDIS_DB")
	if redisDBStr != "" {
		var err error
		redisDB, err = strconv.Atoi(redisDBStr)
		if err != nil {
			log.Printf("Invalid REDIS_DB value, using default: %v", err)
			redisDB = 0
		}
	}

	// 构建Redis地址
	redisAddr := fmt.Sprintf("%s:%s", redisHost, redisPort)

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})
	ctx := context.Background()
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("failed to connect redis: %v", err)
	}
	log.Println("redis connected success")
}

// RedisClient 初始化 Redis 客户端

// SetVerificationCode 设置验证码到 Redis
func SetVerificationCode(email, code string) error {
	ctx := context.Background()
	key := fmt.Sprintf("verify:%s", email)
	return RedisClient.Set(ctx, key, code, 10*time.Minute).Err()
}

// GetVerificationCode 从 Redis 获取验证码
func GetVerificationCode(email string) (string, error) {
	ctx := context.Background()
	key := fmt.Sprintf("verify:%s", email)
	code, err := RedisClient.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("获取验证码失败: %v", err)
	}
	return code, nil
}

// DelVerificationCode 删除验证码 redis
func DelVerificationCode(email string) error {
	ctx := context.Background()
	key := fmt.Sprintf("verify:%s", email)
	return RedisClient.Del(ctx, key).Err()
}
