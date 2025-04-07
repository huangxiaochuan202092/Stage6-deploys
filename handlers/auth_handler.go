package handlers

import (
	"log"
	"net/http"
	"proapp/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// ValidateToken 验证令牌有效性并返回状态
func ValidateToken(c *gin.Context) {
	// 由于使用了AuthMiddleware，如果代码执行到这里
	// 说明令牌有效并且已通过验证

	// 获取中间件设置的用户信息
	userID, exists := c.Get("userID")
	if !exists {
		log.Printf("未找到用户ID信息")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "令牌无效"})
		return
	}

	userRole, exists := c.Get("userRole")
	if !exists {
		log.Printf("未找到用户角色信息")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "令牌无效"})
		return
	}

	// 返回令牌验证成功
	c.JSON(http.StatusOK, gin.H{
		"valid": true,
		"user": gin.H{
			"id":   userID,
			"role": userRole,
		},
	})
}

// RefreshToken 处理令牌刷新请求
func RefreshToken(c *gin.Context) {
	// 从授权头获取令牌
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供授权令牌"})
		return
	}

	// 移除Bearer前缀
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	// 刷新令牌
	newToken, err := utils.RefreshToken(tokenString)
	if err != nil {
		log.Printf("刷新令牌失败: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "刷新令牌失败: " + err.Error()})
		return
	}

	// 返回新令牌
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"token":  newToken,
	})
}
