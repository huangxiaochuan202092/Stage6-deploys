package routes

import (
	"fmt"
	"net/http"
	"proapp/config"
	"proapp/handlers"
	"proapp/middleware"
	"proapp/utils"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.Default()

	// 初始化数据库连接
	config.GetDB()

	// 初始化Redis连接
	utils.InitRedis()

	// 修改模板路径设置
	templatePath := "./templates/*" // 修改这里
	fmt.Printf("Template path: %s\n", templatePath)
	r.LoadHTMLGlob(templatePath)

	// 添加静态文件服务
	r.Static("/static", "./static")

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

	// 问卷管理页面
	r.GET("/wenjuan_manager", func(c *gin.Context) {
		fmt.Println("访问问卷管理页面")
		c.HTML(http.StatusOK, "wenda.html", gin.H{
			"title": "问卷管理系统",
		})
	})

	// API 路由组
	userGroup := r.Group("/user")
	{
		// 不需要认证的接口
		userGroup.POST("/send-code", handlers.SendCodeHandler)
		userGroup.POST("/login-or-register", handlers.LoginOrRegisterHandler)

		// 需要认证的接口
		auth := userGroup.Group("")
		auth.Use(middleware.AuthMiddleware())
		{
			// 添加令牌验证路由
			auth.GET("/validate-token", handlers.ValidateToken)

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

			// 博客管理
			auth.GET("/blog", handlers.GetAllBlogs)
			auth.GET("/blog/:id", handlers.GetBlogById)
			auth.Use(middleware.CheckResourcePermission("blog"))
			{
				auth.POST("/blog", handlers.CreateBlog)
				auth.PUT("/blog/:id", handlers.UpdateBlog)
				auth.DELETE("/blog/:id", handlers.DeleteBlog)
				// 添加点赞相关路由
				auth.POST("/blog/:id/like", handlers.LikeBlog)
				auth.POST("/blog/:id/dislike", handlers.DislikeBlog)
			}

			// 任务管理
			auth.GET("/tasks", handlers.GetAllTasks)
			auth.GET("/tasks/:id", handlers.GetTask)
			auth.Use(middleware.CheckResourcePermission("task"))
			{
				auth.POST("/tasks", handlers.CreateTask)
				auth.PUT("/tasks/:id", handlers.UpdateTask)
				auth.DELETE("/tasks/:id", handlers.DeleteTask)
			}

			// 问卷管理
			wenjuan := auth.Group("/wenjuans")
			{
				// 公共查看接口
				wenjuan.GET("", handlers.GetAllWenjuans)
				wenjuan.GET("/search", handlers.SearchWenjuanByTitle)
				wenjuan.GET("/:id", handlers.GetWenjuanById)
				wenjuan.GET("/:id/answers/:answerId", handlers.GetWenjuanAnswer)
				wenjuan.GET("/categories", handlers.GetAllCategories)

				// 需要权限验证的操作接口
				authWenjuan := wenjuan.Group("")
				authWenjuan.Use(middleware.CheckResourcePermission("wenjuan"))
				{
					authWenjuan.POST("", handlers.CreateWenjuan)
					authWenjuan.PUT("/:id", handlers.UpdateWenjuan)
					authWenjuan.DELETE("/:id", handlers.DeleteWenjuan)
					authWenjuan.POST("/:id/pin", handlers.PinWenjuan)
					authWenjuan.POST("/:id/unpin", handlers.UnpinWenjuan)
					authWenjuan.POST("/categories", handlers.CreateCategory)
					authWenjuan.PUT("/categories/:id", handlers.UpdateCategory)
					authWenjuan.DELETE("/categories/:id", handlers.DeleteCategory)
					authWenjuan.PUT("/:id/answers/:answerId", handlers.UpdateWenjuanAnswer)
					authWenjuan.DELETE("/:id/answers/:answerId", handlers.DeleteWenjuanAnswer)
				}
			}
		}
	}

	return r
}
