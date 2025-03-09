package jwt

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

/**
 * 需要额外定义UserID和UserName，因此自定义 MyClaims，并内嵌官方字段
 * 如果想要保存更多信息，都可以添加到这个结构体中，切记不要保存敏感信息
 **/
type MyClaims struct {
	UserID             uint64 `json:"user_id"`
	Username           string `json:"username"`
	jwt.StandardClaims        // JWT规定的7个官方字段
}

// 定义Secret 用于加密的字符串
var mySecret = []byte("bluebell-plus")

// 获取密钥
func keyFunc(_ *jwt.Token) (i interface{}, err error) {
	return mySecret, nil
}

// TokenExpireDuration 定义JWT的过期时间
const TokenExpireDuration = time.Hour * 24

// rtoken解决atoken过期快的问题
const AccessTokenExpireDuration = time.Hour * 24      // access_token 过期时间
const RefreshTokenExpireDuration = time.Hour * 24 * 7 // refresh_token 过期时间

// GenToken 生成JWT：生成access_token 和 refresh_token
func GenToken(userID uint64, username string) (aToken, rToken string, err error) {
	// 创建声明实例 Token负载
	c := MyClaims{
		userID,
		username,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(AccessTokenExpireDuration).Unix(), // 过期时间
			Issuer:    "bluebell-plus",                                  // 签发人
		},
	}
	// access_token 加密并获得完整的编码后的字符串token：加密算法+Token负载+密钥
	aToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(mySecret)

	// refresh_token 无需存储数据，只为了刷新access_token
	rToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: time.Now().Add(RefreshTokenExpireDuration).Unix(), // 过期时间
		Issuer:    "bluebell-plus",                                   // 签发人
	}).SignedString(mySecret)
	// 使用指定的secret签名并获得完整的编码后的字符串token
	return
}

// ParseToken 解析并验证JWT
func ParseToken(tokenString string) (claims *MyClaims, err error) {
	// 解析TokenString
	var token *jwt.Token
	claims = new(MyClaims)
	token, err = jwt.ParseWithClaims(tokenString, claims, keyFunc)
	if err != nil {
		return
	}
	if !token.Valid { // 校验token
		err = errors.New("invalid token")
	}
	return
}

// RefreshToken 刷新access_token
func RefreshToken(aToken, rToken string) (newAToken, newRToken string, err error) {
	// 若refresh token也无效则直接返回
	if _, err = jwt.Parse(rToken, keyFunc); err != nil {
		return
	}

	// 从过期access token中解析出claims数据，即解析出payload负载信息
	var claims MyClaims
	_, err = jwt.ParseWithClaims(aToken, &claims, keyFunc)
	v, _ := err.(*jwt.ValidationError)

	// 当access token是过期错误并且refresh token没有过期时，就创建一个新的access token
	if v.Errors == jwt.ValidationErrorExpired {
		return GenToken(claims.UserID, claims.Username)
	}
	return
}
