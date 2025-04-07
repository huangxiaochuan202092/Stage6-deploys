package middleware

import (
	"fmt"
	"log"
	"net/http"
	"proapp/config"
	"proapp/utils"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// 错误信息常量
const (
	errNoToken         = "未提供有效的认证令牌"
	errInvalidToken    = "无效的认证令牌: %s"
	errNoUserID        = "未找到用户ID信息，权限验证失败"
	errNoUserRole      = "未获取到用户角色信息"
	errAdminRequired   = "需要管理员权限"
	errNoResourceID    = "请求URL必须包含资源ID"
	errInvalidResID    = "资源ID必须是数字"
	errDBError         = "系统错误"
	errUnsupportedRes  = "不支持的资源类型"
	errVerifyOwnership = "无法验证资源所有权"
	errInvalidUserID   = "用户ID格式错误"
	errNoPermission    = "您没有权限操作此资源"
)

// 从请求中提取令牌
func extractToken(c *gin.Context) string {
	// 从Authorization头提取
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	// 从查询参数提取
	return c.Query("token")
}

// 专门处理jwt.Token类型，获取其中的claims信息
func extractClaimsFromToken(tokenObj interface{}) map[string]interface{} {
	// 先记录详细的类型信息用于调试
	log.Printf("提取claims - 令牌类型: %T", tokenObj)

	// 尝试类型断言为*jwt.Token
	if token, ok := tokenObj.(*jwt.Token); ok && token != nil {
		log.Printf("成功识别为*jwt.Token")

		// 从token中获取claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			log.Printf("成功获取MapClaims: %v", claims)
			return claims
		}

		// 如果无法直接获取MapClaims，尝试从Claims字段获取
		v := reflect.ValueOf(token).Elem()
		if v.IsValid() {
			claimsField := v.FieldByName("Claims")
			if claimsField.IsValid() {
				if claims, ok := claimsField.Interface().(jwt.MapClaims); ok {
					log.Printf("通过反射获取MapClaims: %v", claims)
					return claims
				}

				// 尝试获取任何类型的映射
				if claimsMap, ok := claimsField.Interface().(map[string]interface{}); ok {
					log.Printf("通过反射获取map[string]interface{}: %v", claimsMap)
					return claimsMap
				}
			}
		}

		log.Printf("无法从*jwt.Token获取claims")
	}

	// 尝试直接类型断言为map
	if claims, ok := tokenObj.(map[string]interface{}); ok {
		log.Printf("令牌直接是map[string]interface{}: %v", claims)
		return claims
	}

	// 尝试使用utils包特定的类型（如果有的话）
	// 简化为打印token的详细结构，以便进一步分析
	log.Printf("令牌结构详细内容: %#v", tokenObj)

	// 最后尝试提取已知位置的Claims
	// 基于日志观察到的结构
	v := reflect.ValueOf(tokenObj)
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}

	// 检查是否有Claims字段
	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			fieldName := v.Type().Field(i).Name
			fieldValue := v.Field(i)

			// 寻找可能的claims字段，重点关注map类型的字段
			if fieldValue.Kind() == reflect.Map &&
				(strings.Contains(strings.ToLower(fieldName), "claim") ||
					i == 3) { // 从日志观察，第4个字段(索引3)很可能是claims
				if mapClaims, ok := fieldValue.Interface().(map[string]interface{}); ok {
					log.Printf("在字段%s中发现claims: %v", fieldName, mapClaims)
					return mapClaims
				}
			}
		}
	}

	return nil
}

// JwtAuth JWT认证中间件 - 修正实现
func JwtAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取JWT令牌
		tokenString := extractToken(c)
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": errNoToken})
			c.Abort()
			return
		}

		log.Printf("开始验证令牌: %s", tokenString)

		// 验证令牌
		tokenObj, _, err := utils.VerifyToken(tokenString)
		if err != nil {
			log.Printf("令牌验证失败: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf(errInvalidToken, err.Error())})
			c.Abort()
			return
		}

		log.Printf("令牌验证成功，获取到tokenObj类型: %T", tokenObj)

		// 从Token中提取claims
		claims := extractClaimsFromToken(tokenObj)
		if claims == nil {
			log.Printf("无法从令牌中提取claims")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无法解析令牌内容"})
			c.Abort()
			return
		}

		log.Printf("成功提取claims: %v", claims)

		// 直接从claims map中提取数据
		var userID uint = 0
		var userEmail, userRole string = "", ""

		// 提取ID - 支持多种可能的键名
		for _, key := range []string{"id", "ID", "sub", "userId", "user_id"} {
			if idValue, exists := claims[key]; exists && idValue != nil {
				switch v := idValue.(type) {
				case float64:
					userID = uint(v)
				case int:
					userID = uint(v)
				case uint:
					userID = v
				case string:
					if id, err := strconv.ParseUint(v, 10, 32); err == nil {
						userID = uint(id)
					}
				}
				if userID > 0 {
					break
				}
			}
		}

		// 提取Email
		for _, key := range []string{"email", "Email"} {
			if emailValue, exists := claims[key]; exists && emailValue != nil {
				if email, ok := emailValue.(string); ok {
					userEmail = email
					break
				}
			}
		}

		// 提取Role
		for _, key := range []string{"role", "Role"} {
			if roleValue, exists := claims[key]; exists && roleValue != nil {
				if role, ok := roleValue.(string); ok {
					userRole = role
					break
				}
			}
		}

		// 验证是否获取到有效的用户ID
		if userID == 0 {
			log.Printf("未能获取有效的用户ID")
			c.JSON(http.StatusUnauthorized, gin.H{"error": errInvalidUserID})
			c.Abort()
			return
		}

		log.Printf("解析成功: userID=%d, email=%s, role=%s", userID, userEmail, userRole)

		// 将用户信息添加到上下文
		c.Set("userID", userID)
		c.Set("userEmail", userEmail)
		c.Set("userRole", userRole)

		c.Next()
	}
}

// OptionalJwtAuth 可选的JWT认证 - 修正实现
func OptionalJwtAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 防止崩溃
		defer func() {
			if r := recover(); r != nil {
				log.Printf("OptionalJwtAuth恢复自崩溃: %v", r)
				c.Next()
			}
		}()

		tokenString := extractToken(c)
		if tokenString == "" {
			log.Println("可选JWT认证: 未提供令牌")
			c.Next()
			return
		}

		log.Printf("可选JWT认证: 开始验证令牌")

		// 尝试验证令牌
		tokenObj, _, err := utils.VerifyToken(tokenString)
		if err != nil {
			log.Printf("可选JWT认证: 令牌验证失败: %v", err)
			c.Next()
			return
		}

		// 从Token中提取claims
		claims := extractClaimsFromToken(tokenObj)
		if claims == nil {
			log.Printf("可选JWT认证: 无法从令牌中提取claims")
			c.Next()
			return
		}

		// 使用与JwtAuth相同的方式提取信息
		var userID uint = 0
		var userEmail, userRole string = "", ""

		// 提取ID
		for _, key := range []string{"id", "ID", "sub", "userId", "user_id"} {
			if idValue, exists := claims[key]; exists && idValue != nil {
				switch v := idValue.(type) {
				case float64:
					userID = uint(v)
				case int:
					userID = uint(v)
				case uint:
					userID = v
				case string:
					if id, err := strconv.ParseUint(v, 10, 32); err == nil {
						userID = uint(id)
					}
				}
				if userID > 0 {
					break
				}
			}
		}

		// 提取Email和Role
		for _, key := range []string{"email", "Email"} {
			if emailValue, exists := claims[key]; exists && emailValue != nil {
				if email, ok := emailValue.(string); ok {
					userEmail = email
					break
				}
			}
		}

		for _, key := range []string{"role", "Role"} {
			if roleValue, exists := claims[key]; exists && roleValue != nil {
				if role, ok := roleValue.(string); ok {
					userRole = role
					break
				}
			}
		}

		// 只有当成功提取到userID时才设置上下文变量
		if userID > 0 {
			c.Set("userID", userID)
			c.Set("userEmail", userEmail)
			c.Set("userRole", userRole)
			log.Printf("可选JWT认证成功: 用户ID=%d, Email=%s, Role=%s", userID, userEmail, userRole)
		} else {
			log.Println("可选JWT认证: 未能提取有效的用户ID")
		}

		c.Next()
	}
}

// RequireAdmin 管理员权限检查中间件
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("userRole")

		// 检查角色是否存在
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": errNoUserRole})
			c.Abort()
			return
		}

		// 检查是否是管理员
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": errAdminRequired})
			c.Abort()
			return
		}

		c.Next()
	}
}

// 将任意类型的userID转换为uint
func convertToUint(userID interface{}) (uint, error) {
	switch v := userID.(type) {
	case uint:
		return v, nil
	case int:
		return uint(v), nil
	case float64:
		return uint(v), nil
	case int64:
		return uint(v), nil
	case string:
		idInt, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return 0, err
		}
		return uint(idInt), nil
	default:
		return 0, fmt.Errorf("不支持的ID类型: %T", userID)
	}
}

// CheckResourcePermission 检查用户是否有权限操作特定资源
func CheckResourcePermission(resourceType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文获取用户ID和角色
		userID, exists := c.Get("userID")
		if !exists {
			log.Println(errNoUserID)
			c.JSON(http.StatusUnauthorized, gin.H{"error": errNoUserID})
			c.Abort()
			return
		}

		// 获取用户角色 - 管理员直接放行
		role, _ := c.Get("userRole")
		userRole := fmt.Sprintf("%v", role)
		if userRole == "admin" {
			log.Printf("用户ID %v 是管理员，拥有全部权限", userID)
			c.Next()
			return
		}

		// 获取资源ID
		resourceIDStr := c.Param("id")
		if resourceIDStr == "" {
			log.Println("请求URL未包含资源ID参数")
			c.JSON(http.StatusBadRequest, gin.H{"error": errNoResourceID})
			c.Abort()
			return
		}

		resourceID, err := strconv.Atoi(resourceIDStr)
		if err != nil {
			log.Printf("资源ID格式错误: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": errInvalidResID})
			c.Abort()
			return
		}

		// 获取数据库连接
		db := config.GetDB()
		if db == nil {
			log.Printf("无法获取数据库连接")
			c.JSON(http.StatusInternalServerError, gin.H{"error": errDBError})
			c.Abort()
			return
		}

		// 根据资源类型确定表和字段名
		var tableName, idField, creatorField string
		switch resourceType {
		case "blog":
			tableName = "blogs"
			idField = "id"
			creatorField = "user_id"
		case "task":
			tableName = "tasks"
			idField = "id"
			creatorField = "creator_id"
		default:
			log.Printf("不支持的资源类型: %s", resourceType)
			c.JSON(http.StatusInternalServerError, gin.H{"error": errUnsupportedRes})
			c.Abort()
			return
		}

		// 查询资源的创建者ID
		var creatorID uint
		query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", creatorField, tableName, idField)
		err = db.Raw(query, resourceID).Scan(&creatorID).Error
		if err != nil {
			log.Printf("查询资源所有者失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": errVerifyOwnership})
			c.Abort()
			return
		}

		// 转换用户ID到uint类型
		userIDUint, err := convertToUint(userID)
		if err != nil {
			log.Printf("转换用户ID失败: %v (类型: %T, 值: %v)", err, userID, userID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": errInvalidUserID})
			c.Abort()
			return
		}

		// 检查当前用户是否为资源创建者
		if creatorID != userIDUint {
			log.Printf("权限拒绝: 用户ID %v 不是资源 %s ID %d 的所有者 (所有者ID: %d)",
				userIDUint, resourceType, resourceID, creatorID)
			c.JSON(http.StatusForbidden, gin.H{"error": errNoPermission})
			c.Abort()
			return
		}

		log.Printf("权限检查通过: 用户ID %v 是资源 %s ID %d 的所有者", userIDUint, resourceType, resourceID)
		c.Next()
	}
}
