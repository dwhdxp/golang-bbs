package mysql

import (
	"bluebell_backend/models"
	"bluebell_backend/pkg/snowflake"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"errors"
)

/**
 * 封装数据库操作
 * 供logic层根据业务需求调用
 **/

const secret = "huchao.vip"

// Md5算法加密用户密码
func encryptPassword(data []byte) (result string) {
	h := md5.New()
	h.Write([]byte(secret))
	return hex.EncodeToString(h.Sum(data))
}

// 注册业务：检查指定username的用户是否存在
func CheckUserExist(username string) (error error) {
	sqlStr := `select count(user_id) from user where username = ?`
	var count int
	if err := db.Get(&count, sqlStr, username); err != nil {
		return err
	}
	if count > 0 {
		return errors.New(ErrorUserExit)
	}
	return
}

// 注册业务：向数据库中插入一条新的用户
func InsertUser(user models.User) (error error) {
	// 加密密码
	user.Password = encryptPassword([]byte(user.Password))
	// 执行sql插入数据
	sqlstr := `insert into user(user_id,username,password,email,gender) values(?,?,?,?,?)`
	_, err := db.Exec(sqlstr, user.UserID, user.UserName, user.Password, user.Email, user.Gender)
	return err
}

func Register(user *models.User) (err error) {
	sqlStr := "select count(user_id) from user where username = ?"
	var count int64
	err = db.Get(&count, sqlStr, user.UserName)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if count > 0 {
		// 用户已存在
		return errors.New(ErrorUserExit)
	}
	// 生成user_id
	userID, err := snowflake.GetID()
	if err != nil {
		return ErrorGenIDFailed
	}
	// 生成加密密码
	password := encryptPassword([]byte(user.Password))
	// 把用户插入数据库
	sqlStr = "insert into user(user_id, username, password) values (?,?,?)"
	_, err = db.Exec(sqlStr, userID, user.UserName, password)
	return
}

// 登录业务：判断用户是否存在以及密码是否正确
func Login(user *models.User) (err error) {
	originPassword := user.Password // 记录下原始密码
	sqlStr := "select user_id, username, password from user where username = ?"
	err = db.Get(user, sqlStr, user.UserName)
	// 查询数据库出错
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	// 用户不存在
	if err == sql.ErrNoRows {
		return errors.New(ErrorUserNotExit)
	}
	// 生成加密密码与查询到的密码比较
	password := encryptPassword([]byte(originPassword))
	if user.Password != password {
		return errors.New(ErrorPasswordWrong)
	}
	return nil
}

// GetUserByID 根据user_id查询作者信息
func GetUserByID(id uint64) (user *models.User, err error) {
	user = new(models.User)
	sqlStr := `select user_id, username from user where user_id = ?`
	err = db.Get(user, sqlStr, id)
	return
}
