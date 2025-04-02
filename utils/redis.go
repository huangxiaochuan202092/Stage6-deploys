package utils

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
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
