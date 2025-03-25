package routes

import (
	"fmt"
	"net/http"
	"proapp/config"
	"proapp/handlers"
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
		// 发送验证码的处理器
		userGroup.POST("/send-code", handlers.SendCodeHandler)

		// 登录或注册的处理器
		userGroup.POST("/login-or-register", handlers.LoginOrRegisterHandler)

		// 用户管理
		// 获取所有用户
		userGroup.GET("/", handlers.GetAllUsersHandler)

		// 根据 ID 获取用户
		userGroup.GET("/:id", handlers.GetUserByIdHandler)

		// 更新用户信息
		userGroup.PUT("/:id", handlers.UpdateUserHandler)

		// 删除用户
		userGroup.DELETE("/:id", handlers.DeleteUserHandler)

		//任务管理
		// 创建任务
		userGroup.POST("/tasks", handlers.CreateTask)

		// 获取所有任务
		userGroup.GET("/tasks", handlers.GetAllTasks)

		// 获取单个任务
		userGroup.GET("/tasks/:id", handlers.GetTask)

		// 更新任务
		userGroup.PUT("/tasks/:id", handlers.UpdateTask)

		// 删除任务
		userGroup.DELETE("/tasks/:id", handlers.DeleteTask)

		//博客管理
		// 创建博客
		userGroup.POST("/blog", handlers.CreateBlog)

		// 获取所有博客
		userGroup.GET("/blog", handlers.GetAllBlogs)

		// 获取单个博客
		userGroup.GET("/blog/:id", handlers.GetBlogById)

		// 更新博客
		userGroup.PUT("/blog/:id", handlers.UpdateBlog)

		// 删除博客
		userGroup.DELETE("/blog/:id", handlers.DeleteBlog)

		// 点赞博客
		userGroup.POST("/blog/:id/like", handlers.LikeBlog)

		// 取消点赞博客
		userGroup.POST("/blog/:id/dislike", handlers.DislikeBlog)

		// 问卷管理路由组
		wenjuanGroup := userGroup.Group("/wenjuans")
		{
			// 把搜索路由放在最前面，避免被其他路由匹配
			wenjuanGroup.GET("/search", handlers.SearchWenjuanByTitle)

			wenjuanGroup.POST("", handlers.CreateWenjuan)
			wenjuanGroup.GET("", handlers.GetAllWenjuans)
			wenjuanGroup.GET("/:id", handlers.GetWenjuanById)
			wenjuanGroup.PUT("/:id", handlers.UpdateWenjuan)
			wenjuanGroup.DELETE("/:id", handlers.DeleteWenjuan)
			//根据标题搜索问卷

			// 置顶问卷
			wenjuanGroup.POST("/:id/pin", handlers.PinWenjuan)
			// 取消置顶问卷
			wenjuanGroup.POST("/:id/unpin", handlers.UnpinWenjuan)

			// 答案管理
			wenjuanGroup.POST("/:id/submit", handlers.SubmitWenjuanAnswer)
			wenjuanGroup.PUT("/:id/answers/:answerId", handlers.UpdateWenjuanAnswer)
			wenjuanGroup.DELETE("/:id/answers/:answerId", handlers.DeleteWenjuanAnswer)
			wenjuanGroup.GET("/:id/answers/:answerId", handlers.GetWenjuanAnswer)

			// PDF下载
			wenjuanGroup.GET("/:id/download", handlers.DownloadWenjuanAndAnswers)

			// 分类管理
			wenjuanGroup.GET("/categories", handlers.GetAllCategories)
			wenjuanGroup.POST("/categories", handlers.CreateCategory)
			wenjuanGroup.PUT("/categories/:id", handlers.UpdateCategory)
			wenjuanGroup.DELETE("/categories/:id", handlers.DeleteCategory)

			// 问卷分类关联
			wenjuanGroup.POST("/:id/categories/:categoryId", handlers.AddCategoryToWenjuan)
		}
	}

	return r
}
