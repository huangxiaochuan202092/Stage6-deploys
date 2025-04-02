package utils

import (
	"encoding/base64"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// 密钥应该从配置中获取
var jwtKey = []byte("tOunsUhUhkAZvoTEj6ySe2JaJ+yPfeyehUmikyZ0aGw") // 请在生产环境中更改为安全的密钥

// GetJwtKey 返回用于签名和验证JWT的密钥
func GetJwtKey() []byte {
	log.Printf("获取JWT密钥: %s", base64.StdEncoding.EncodeToString(jwtKey[:8])+"...") // 只打印部分密钥，安全考虑
	return jwtKey
}

// GenerateToken 生成JWT令牌
func GenerateToken(userID uint, email string, role string) (string, error) {
	// 设置令牌有效期，例如24小时
	expirationTime := time.Now().Add(24 * time.Hour)

	// 设置令牌的声明
	claims := jwt.MapClaims{
		"id":    userID,
		"email": email,
		"role":  role,
		"exp":   expirationTime.Unix(),
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 使用密钥签名令牌
	tokenString, err := token.SignedString(GetJwtKey()) // 使用GetJwtKey()保证一致性
	if err != nil {
		log.Printf("生成token失败: %v", err)
		return "", err
	}

	log.Printf("成功生成token，用户ID=%d，邮箱=%s，角色=%s", userID, email, role)
	return tokenString, nil
}

// VerifyToken 验证JWT令牌
func VerifyToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 确保签名算法正确
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Printf("意外的签名方法: %v", token.Header["alg"])
			return nil, jwt.ErrSignatureInvalid
		}

		return GetJwtKey(), nil // 使用GetJwtKey()保证一致性
	})

	return token, err
}
