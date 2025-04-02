package middleware

import (
	"log"
	"net/http"
	"proapp/utils"
	"strconv"
	"strings"

	"proapp/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// AuthMiddleware 验证JWT令牌并设置用户信息到上下文
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 Authorization 头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Printf("缺少Authorization头")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供认证信息"})
			c.Abort()
			return
		}

		// 检查 Bearer 前缀
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Printf("Authorization头格式错误: %s", authHeader)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证格式无效"})
			c.Abort()
			return
		}

		// 解析 JWT 令牌
		tokenString := parts[1]

		log.Printf("开始验证令牌: %s...", tokenString[:10]) // 只打印令牌的前10个字符，安全考虑

		// 使用统一的验证方法
		token, err := utils.VerifyToken(tokenString)

		if err != nil {
			log.Printf("解析JWT令牌失败: %v", err)
			// 如果是签名错误，可能需要重新登录
			if err == jwt.ErrSignatureInvalid {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "令牌签名无效，请重新登录",
					"code":  "invalid_signature",
				})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的认证令牌"})
			}
			c.Abort()
			return
		}

		// 验证令牌声明
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// 将用户信息设置到上下文
			userId, exists := claims["id"]
			if !exists {
				log.Printf("JWT令牌缺少用户ID")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的令牌：缺少用户ID"})
				c.Abort()
				return
			}

			// 提取用户角色，确保始终存在
			userRole, exists := claims["role"]
			if !exists {
				log.Printf("JWT令牌缺少用户角色，将使用默认角色'user'")
				userRole = "user" // 设置默认角色
			}

			// 记录提取的信息
			log.Printf("认证成功，用户ID: %v, 角色: %v", userId, userRole)

			// 将信息设置到上下文
			c.Set("userID", userId)
			c.Set("userRole", userRole)

			c.Next()
		} else {
			log.Printf("无效的JWT声明")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的认证令牌"})
			c.Abort()
			return
		}
	}
}

// RequireAdmin 检查用户是否具有管理员权限
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("userRole")
		if !exists {
			log.Printf("未找到用户角色信息")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无法验证用户权限"})
			c.Abort()
			return
		}

		// 检查是否为管理员
		if userRole != "admin" {
			log.Printf("用户角色 '%v' 不是管理员", userRole)
			c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
			c.Abort()
			return
		}

		log.Printf("管理员权限验证通过")
		c.Next()
	}
}

// CheckResourcePermission 检查用户是否对特定资源有权限
func CheckResourcePermission(resourceType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("userRole")
		if !exists {
			log.Printf("未找到用户角色信息")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无法验证用户权限"})
			c.Abort()
			return
		}

		userID, exists := c.Get("userID")
		if !exists {
			log.Printf("未找到用户ID信息")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无法验证用户权限"})
			c.Abort()
			return
		}

		method := c.Request.Method

		// 如果是管理员，直接放行所有操作
		if userRole == "admin" {
			c.Next()
			return
		}

		// 普通用户权限控制
		if userRole == "user" {
			// GET请求允许访问所有资源
			if method == "GET" {
				c.Next()
				return
			}

			// POST/PUT/DELETE 需要验证资源所有权
			resourceID := c.Param("id")
			if resourceID != "" {
				// 将字符串 ID 转换为整数
				id, err := strconv.Atoi(resourceID)
				if err != nil {
					c.JSON(400, gin.H{"error": "无效的资源ID"})
					c.Abort()
					return
				}

				var hasPermission bool
				switch resourceType {
				case "blog":
					blog, err := services.GetBlogById(id)
					if err == nil && blog != nil {
						hasPermission = uint(blog.UserID) == userID
					}
				case "task":
					task, err := services.GetTaskById(id)
					if err == nil && task != nil {
						hasPermission = uint(task.CreatorID) == userID
					}
				case "wenjuan":
					wenjuan, err := services.GetWenjuanById(id)
					if err == nil && wenjuan != nil {
						hasPermission = uint(wenjuan.CreatorID) == userID
					}
				}

				if !hasPermission {
					c.JSON(403, gin.H{"error": "没有权限操作此资源"})
					c.Abort()
					return
				}
			}

			// 允许创建新资源
			if method == "POST" && c.Param("id") == "" {
				c.Next()
				return
			}

			c.Next()
			return
		}

		// 其他角色无权限
		c.JSON(403, gin.H{"error": "没有权限"})
		c.Abort()
	}
}
