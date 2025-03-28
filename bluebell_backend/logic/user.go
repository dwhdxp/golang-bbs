package logic

import (
	"bluebell_backend/dao/mysql"
	"bluebell_backend/models"
	"bluebell_backend/pkg/jwt"
	"bluebell_backend/pkg/rabbitmq"
	"bluebell_backend/pkg/snowflake"
	"fmt"
	"go.uber.org/zap"
)

// 注册业务逻辑代码
func SignUp(p *models.RegisterForm) (error error) {
	//// 1.判断用户是否注册
	//err := mysql.CheckUserExist(p.UserName)
	//if err != nil {
	//	return err
	//}

	// 1.雪花算法生成UID
	userId, err := snowflake.GetID()
	if err != nil {
		return mysql.ErrorGenIDFailed
	}

	// 2.写入数据库
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
	/*
		限制账号在同一时间只能在一个设备上登录
		* 将 {userID : Access Token} 缓存至 Redis中
		* 设置过期时间与 Access Token 过期时间相同，避免Token过期但Redis中仍存在旧记录
	*/
	user.AccessToken = accessToken
	user.RefreshToken = refreshToken
	return
}

// SignUpNew 注册逻辑代码优化，将邮件发送任务异步发布到RabbitMQ队列中
func SignUpNew(p *models.RegisterForm) error {
	// 判断用户是否注册
	err := mysql.CheckUserExist(p.UserName)
	if err != nil {
		return err
	}

	var errs []error
	// 若用户提供电子邮件，异步发送邮件到用户邮箱
	if p.Email != "" {
		Ed := &models.RegisterEmailData{
			Email:    p.Email,
			UserName: p.UserName,
			Password: p.Password,
		}
		zap.L().Debug("emaildetail", zap.String("Username", Ed.UserName),
			zap.String("Email", Ed.Email))
		// 使用生产者发布邮件任务到队列
		err := rabbitmq.PublishEmailTask(Ed)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to publish email task: %w;", err))
		}
		//// 不使用消息队列
		//err := email.SendEmail(Ed)
		//if err != nil {
		//	errs = append(errs, err)
		//}
	}

	// 用户注册逻辑 保存到mysql
	if err := SignUp(p); err != nil {
		errs = append(errs, fmt.Errorf("signup error: %w", err))
	}
	zap.L().Debug("signup success", zap.String("email", p.Email),
		zap.String("username", p.UserName))
	if len(errs) > 0 {
		return fmt.Errorf("multiple errors occurred: %v", errs)
	}
	return nil
}
