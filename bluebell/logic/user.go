package logic

import (
	"bluebell/dao/mysql"
	"bluebell/models"
	"bluebell/pkg/jwt"
	"bluebell/pkg/snowflake"
)

// 注册业务逻辑代码
func SignUp(p *models.RegisterForm) (error error) {
	// 1.判断用户是否注册
	err := mysql.CheckUserExist(p.UserName)
	if err != nil {
		return err
	}

	// 2.雪花算法生成UID
	userId, err := snowflake.GetID()
	if err != nil {
		return mysql.ErrorGenIDFailed
	}

	// 3.写入数据库
	u := models.User{
		UserID:   userId,
		UserName: p.UserName,
		Password: p.Password,
		Email:    p.Email,
		Gender:   p.Gender,
	}
	return mysql.InsertUser(u)
}

// 登录业务逻辑代码
func Login(p *models.LoginForm) (user *models.User, error error) {
	// 1.判断用户是否存在以及密码是否输入正确
	user = &models.User{
		UserName: p.UserName,
		Password: p.Password,
	}
	if err := mysql.Login(user); err != nil {
		return nil, err
	}

	// 2.生成JWT：AccessToken和RefreshToken
	accessToken, refreshToken, err := jwt.GenToken(user.UserID, user.UserName)
	if err != nil {
		return
	}
	user.AccessToken = accessToken
	user.RefreshToken = refreshToken
	return
}

//
