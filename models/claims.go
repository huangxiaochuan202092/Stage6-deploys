package models

import "github.com/golang-jwt/jwt/v4"

// Claims 用于JWT令牌的声明结构体
type Claims struct {
	jwt.RegisteredClaims
	ID        uint   `json:"id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	ExpiresAt int64  `json:"exp"`
}
