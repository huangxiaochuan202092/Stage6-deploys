package utils

import (
	"encoding/base64"
	"errors"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// 从环境变量获取密钥，如果不存在则使用默认值
func getJwtKeyFromEnv() []byte {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	return []byte(secretKey)
}

// 密钥应该从配置中获取
var jwtKey = getJwtKeyFromEnv()

// GetJwtKey 返回用于签名和验证JWT的密钥
func GetJwtKey() []byte {
	log.Printf("获取JWT密钥: %s", base64.StdEncoding.EncodeToString(jwtKey[:8])+"...") // 只打印部分密钥，安全考虑
	return jwtKey
}

// TokenExpiration 定义令牌有效期，更改为7天
const TokenExpiration = 7 * 24 * time.Hour

// GenerateToken 生成JWT令牌
func GenerateToken(userID uint, email string, role string) (string, error) {
	// 设置令牌有效期为7天
	expirationTime := time.Now().Add(TokenExpiration)

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
	tokenString, err := token.SignedString(GetJwtKey())
	if err != nil {
		log.Printf("生成token失败: %v", err)
		return "", err
	}

	log.Printf("成功生成token，用户ID=%d，邮箱=%s，角色=%s", userID, email, role)
	return tokenString, nil
}

// VerifyToken 验证JWT令牌并返回claims
func VerifyToken(tokenString string) (*jwt.Token, jwt.MapClaims, error) {
	claims := jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// 确保签名算法正确
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Printf("意外的签名方法: %v", token.Header["alg"])
			return nil, jwt.ErrSignatureInvalid
		}

		return GetJwtKey(), nil
	})

	return token, claims, err
}

// RefreshToken 刷新JWT令牌
func RefreshToken(tokenString string) (string, error) {
	token, claims, err := VerifyToken(tokenString)

	// 如果令牌无效但不是因为过期导致的，则返回错误
	if err != nil && !errors.Is(err, jwt.ErrTokenExpired) {
		return "", err
	}

	// 确保token有效或仅是过期
	if !token.Valid && !errors.Is(err, jwt.ErrTokenExpired) {
		return "", errors.New("无效的令牌")
	}

	// 从claims中提取必要信息
	var userID uint
	if id, ok := claims["id"]; ok {
		// 根据实际类型进行转换
		switch v := id.(type) {
		case float64:
			userID = uint(v)
		case int:
			userID = uint(v)
		case uint:
			userID = v
		}
	}

	email, _ := claims["email"].(string)
	role, _ := claims["role"].(string)

	// 生成新令牌
	return GenerateToken(userID, email, role)
}

// GetTokenRemainingTime 获取令牌剩余有效时间
func GetTokenRemainingTime(tokenString string) (time.Duration, error) {
	token, claims, err := VerifyToken(tokenString)
	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, errors.New("无效的令牌")
	}

	// 从claims中获取过期时间
	if exp, ok := claims["exp"]; ok {
		switch v := exp.(type) {
		case float64:
			expTime := time.Unix(int64(v), 0)
			return time.Until(expTime), nil
		case int64:
			expTime := time.Unix(v, 0)
			return time.Until(expTime), nil
		default:
			return 0, errors.New("无效的过期时间格式")
		}
	}

	return 0, errors.New("令牌没有过期时间")
}
