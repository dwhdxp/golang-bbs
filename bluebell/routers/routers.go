package routers

import (
	"bluebell/controller"
	"bluebell/logger"
	"bluebell/middlewares"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/gin-contrib/pprof"
)

// SetupRouter 设置路由
func SetupRouter(mode string) *gin.Engine {
	if mode == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	// 设置中间件
	r.Use(logger.GinLogger(),
		logger.GinRecovery(true),                           // Recovery recover掉项目可能出现的panic，并使用zap记录相关日志
		middlewares.RateLimitMiddleware(2*time.Second, 40), // 每两秒钟添加十个令牌  全局限流
	)

	v1 := r.Group("/api/v1") // 创建API v1版本路由组
	// 登录注册业务
	v1.POST("/login", controller.LoginHandler)
	v1.POST("/signup", controller.SignUpHandler)
	v1.GET("/refresh_token", controller.RefreshTokenHandler) // 刷新accessToken

	// 帖子业务
	v1.GET("/post/:id", controller.PostDetailHandler) // 根据帖子id查询帖子详情
	v1.GET("/posts", controller.PostListHandler)      // 分页展示帖子列表
	v1.GET("/posts2", controller.PostList2Handler)    // 根据发布时间或者分数排序分页展示(所有/某社区)帖子列表
	v1.GET("/search", controller.PostSearchHandler)   // 搜索业务-搜索帖子

	// 社区业务
	v1.GET("/community", controller.CommunityHandler)           // 获取分类社区列表
	v1.GET("/community/:id", controller.CommunityDetailHandler) // 根据社区id查找社区详情

	// JWT认证中间件
	v1.Use(middlewares.JWTAuthMiddleware())
	{
		v1.POST("/post", controller.CreatePostHandler) // 创建帖子

		v1.POST("/vote", controller.VoteHandler) // 投票

		v1.POST("/comment", controller.CommentHandler)    // 评论
		v1.GET("/comment", controller.CommentListHandler) // 评论列表

		v1.GET("/ping", func(c *gin.Context) {
			c.String(http.StatusOK, "ping success")
		})
	}

	pprof.Register(r) // 注册pprof相关路由

	// 处理其他路由
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"msg": "404",
		})
	})
	return r
}
