package controller

import (
	"bluebell/dao/mysql"
	"bluebell/logic"
	"bluebell/models"
	"bluebell/pkg/jwt"
	"fmt"
	"github.com/go-playground/validator/v10"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SignUpHandler 注册业务
func SignUpHandler(c *gin.Context) {
	// 1.获取请求参数
	var fo *models.RegisterForm
	// 2.参数校验
	if err := c.ShouldBindJSON(&fo); err != nil {
		// 通过zap记录错误
		zap.L().Error("SignUp with invalid param", zap.Error(err))
		// 判断err是否为 validator.ValidationErrors类型的errors
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			// 若不是，直接返回错误
			ResponseError(c, CodeInvalidParams)
			return
		}
		// 若是，则翻译后返回错误
		ResponseErrorWithMsg(c, CodeInvalidParams, removeTopStruct(errs.Translate(trans)))
		return
	}
	fmt.Printf("fo: %v\n", fo)
	// 3.业务逻辑处理——注册
	if err := logic.SignUp(fo); err != nil {
		zap.L().Error("logic.signup failed", zap.Error(err))
		if err.Error() == mysql.ErrorUserExit {
			ResponseError(c, CodeUserExist)
			return
		}
		ResponseError(c, CodeServerBusy)
		return
	}
	// 4.返回响应
	ResponseSuccess(c, nil)
}

// LoginHandler 登录业务
func LoginHandler(c *gin.Context) {
	// 1.获取请求参数
	var u *models.LoginForm
	// 2.参数校验
	if err := c.ShouldBindJSON(&u); err != nil {
		// 通过zap记录错误
		zap.L().Error("Login with invalid param", zap.Error(err))
		// 判断err是否为 validator.ValidationErrors类型的errors
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			// 若不是，直接返回错误
			ResponseError(c, CodeInvalidParams)
			return
		}
		// 若是，则翻译后返回错误
		ResponseErrorWithMsg(c, CodeInvalidParams, removeTopStruct(errs.Translate(trans)))
		return
	}

	// 3.业务逻辑处理——登录
	user, err := logic.Login(u)
	if err != nil {
		zap.L().Error("logic.Login failed", zap.String("username", u.UserName), zap.Error(err))
		if err.Error() == mysql.ErrorUserNotExit {
			ResponseError(c, CodeUserNotExist)
			return
		}
		ResponseError(c, CodeInvalidParams)
		return
	}
	// 4.返回响应
	ResponseSuccess(c, gin.H{
		"user_id":       fmt.Sprintf("%d", user.UserID), // js识别的最大值：id值大于1<<53-1  int64: i<<63-1
		"user_name":     user.UserName,
		"access_token":  user.AccessToken,
		"refresh_token": user.RefreshToken,
	})
}

// RefreshTokenHandler 刷新accessToken
func RefreshTokenHandler(c *gin.Context) {
	rt := c.Query("refresh_token")
	// 客户端携带Token有三种方式 1.放在请求头 2.放在请求体 3.放在URI
	// 这里假设Token放在Header的 Authorization 中，并使用 Bearer 开头
	// 这里的具体实现方式要依据你的实际业务情况决定
	authHeader := c.Request.Header.Get("Authorization")
	if authHeader == "" {
		ResponseErrorWithMsg(c, CodeInvalidToken, "请求头缺少Auth Token")
		c.Abort()
		return
	}
	// 按空格分割
	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		ResponseErrorWithMsg(c, CodeInvalidToken, "Token格式不对")
		c.Abort()
		return
	}
	aToken, rToken, err := jwt.RefreshToken(parts[1], rt)
	zap.L().Error("jwt.RefreshToken failed", zap.Error(err))
	c.JSON(http.StatusOK, gin.H{
		"access_token":  aToken,
		"refresh_token": rToken,
	})
}
