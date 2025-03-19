package middlewares

import (
	"bluebell_backend/controller"
	"bluebell_backend/pkg/jwt"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware JWT的认证中间件，实现鉴权
func JWTAuthMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		// client携带Token有三种方式 1.Header 2.Body 3.URI
		// 将Token放在Header的Authorization字段，并确保格式为Bearer <token>
		// 1.从请求Header获取Authorization字段
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			controller.ResponseErrorWithMsg(c, controller.CodeInvalidToken, "请求头缺少Auth Token")
			c.Abort()
			return
		}

		// 2.解析得到tokenString
		// parts[0]=="Bearer", parts[1]==tokenString
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			controller.ResponseErrorWithMsg(c, controller.CodeInvalidToken, "Token格式错误，应为Bearer <token>")
			c.Abort()
			return
		}

		// 3.解析并验证JWT
		mc, err := jwt.ParseToken(parts[1])
		if err != nil {
			fmt.Println(err)
			controller.ResponseError(c, controller.CodeInvalidToken)
			c.Abort()
			return
		}

		/*
			限制账号在同一时间只能在一个设备上登录
			* 从Redis中获取 userID 对应的 Token
			* 比较 当前请求Token 与 Redis中的Token 是否一致
		*/

		// 4.将claims中的userID信息保存到请求的上下文c
		c.Set(controller.ContextUserIDKey, mc.UserID)
		c.Next() // 后续的处理函数可以用过c.Get(ContextUserIDKey)来获取当前请求的用户信息
	}
}
