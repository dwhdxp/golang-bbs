package logic

import (
	"bluebell_backend/dao/mysql"
	"bluebell_backend/models"
	"bluebell_backend/pkg/jwt"
	"bluebell_backend/pkg/snowflake"
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
	user = &models.User{
		UserName: p.UserName,
		Password: p.Password,
	}
	if err := mysql.Login(user); err != nil {
		return nil, err
	}
	// 生成JWT
	//return jwt.GenToken(user.UserID,user.UserName)
	accessToken, refreshToken, err := jwt.GenToken(user.UserID, user.UserName)
	if err != nil {
		return
	}
	user.AccessToken = accessToken
	user.RefreshToken = refreshToken
	return
}
