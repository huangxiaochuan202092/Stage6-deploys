package routes

import (
	"fmt"
	"net/http"
	"proapp/config"
	"proapp/handlers"
	"proapp/middleware"

	//导入新的模块化问卷包
	"proapp/utils"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.Default()

	// 添加CORS中间件
	r.Use(corsMiddleware())

	// 添加请求日志中间件
	r.Use(requestLogger())

	// 初始化数据库连接
	config.GetDB()

	// 初始化Redis连接
	utils.InitRedis()

	// 修改模板路径设置
	templatePath := "./templates/*"
	fmt.Printf("Template path: %s\n", templatePath)
	r.LoadHTMLGlob(templatePath)

	// 设置主页路由
	r.GET("/", func(c *gin.Context) {
		fmt.Println("Accessing root route /")
		c.HTML(http.StatusOK, "user_auth.html", gin.H{
			"title": "用户验证",
		})
	})
	// 后台管理页面
	r.GET("/admin", func(c *gin.Context) {
		fmt.Println("访问后台管理页面")
		c.HTML(http.StatusOK, "admin.html", nil)
	})

	// 用户管理页面
	r.GET("/user_manager", func(c *gin.Context) {
		fmt.Println("访问用户管理页面")
		c.HTML(http.StatusOK, "user.html", nil)
	})

	// 任务管理页面
	r.GET("/task_manager", func(c *gin.Context) {
		fmt.Println("访问任务管理页面")
		c.HTML(http.StatusOK, "task.html", nil)
	})

	// 博客管理页面
	r.GET("/blog_manager", func(c *gin.Context) {
		fmt.Println("访问博客管理页面")
		c.HTML(http.StatusOK, "blog.html", nil)
	})

	// API 路由组 - 用户管理、博客管理、任务管理保持不变
	userGroup := r.Group("/user")
	{
		// 不需要认证的接口
		userGroup.POST("/send-code", handlers.SendCodeHandler)
		userGroup.POST("/login-or-register", handlers.LoginOrRegisterHandler)

		// 需要认证的接口
		auth := userGroup.Group("")
		auth.Use(middleware.JwtAuth()) // Use the more flexible JwtAuth middleware
		{
			// 添加令牌验证路由
			auth.GET("/validate-token", handlers.ValidateToken)
			auth.POST("/refresh-token", handlers.RefreshToken)

			// 管理员接口
			adminAuth := auth.Group("")
			adminAuth.Use(middleware.RequireAdmin())
			{
				adminAuth.GET("/", handlers.GetAllUsersHandler)
				adminAuth.PUT("/:id", handlers.UpdateUserHandler) // 管理员可以更新任何用户
				adminAuth.DELETE("/:id", handlers.DeleteUserHandler)
			}

			// 普通用户接口
			auth.GET("/:id", handlers.GetUserByIdHandler)
			auth.PUT("/self", handlers.UpdateUserSelfHandler) // 用户只能更新自己的信息

			// 博客管理 - 基本操作
			// 所有用户都可以查看和点赞博客
			auth.GET("/blog", handlers.GetAllBlogs)
			auth.GET("/blog/:id", handlers.GetBlogById)
			auth.POST("/blog/:id/like", handlers.LikeBlog)
			auth.POST("/blog/:id/dislike", handlers.DislikeBlog)

			// 博客创建 - 不需要资源ID，任何登录用户都可以创建博客
			auth.POST("/blog", handlers.CreateBlog)

			// 博客修改和删除 - 需要权限检查中间件
			blogAuth := auth.Group("/blog")
			blogAuth.Use(middleware.CheckResourcePermission("blog"))
			{
				blogAuth.PUT("/:id", handlers.UpdateBlog)    // 只有博客所有者可以修改
				blogAuth.DELETE("/:id", handlers.DeleteBlog) // 只有博客所有者可以删除
			}

			// 任务管理 - 所有登录用户都可以查看任务列表和任务详情
			auth.GET("/tasks", handlers.GetAllTasks) // 所有用户可以查看任务列表
			auth.GET("/tasks/:id", handlers.GetTask) // 所有用户可以查看单个任务详情

			// 任务管理 - 创建任务不需要检查所有权，任何登录用户都可以创建
			auth.POST("/tasks", handlers.CreateTask) // 任何登录用户都可以创建任务

			// 任务管理 - 修改和删除操作需要检查资源所有权
			taskAuth := auth.Group("/tasks")
			taskAuth.Use(middleware.CheckResourcePermission("task")) // 使用资源所有权检查中间件
			{
				taskAuth.PUT("/:id", handlers.UpdateTask)    // 只有任务创建者才能更新任务
				taskAuth.DELETE("/:id", handlers.DeleteTask) // 只有任务创建者才能删除任务
			}
		}
	}

	return r
}

// CORS中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// 请求日志记录中间件
func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Printf("[%s] %s - 请求参数: %v\n",
			c.Request.Method,
			c.Request.URL.Path,
			c.Request.URL.Query())

		// 记录请求前
		c.Next()
		// 记录响应状态
		fmt.Printf("[%s] %s - 响应状态: %d\n",
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status())
	}
}

// 添加一个可选的JWT验证，允许URL参数中的token
func OptionalJwtAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 实现类似JwtAuth但更宽松的验证
		// ...
		c.Next()
	}
}
