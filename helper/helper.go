package helper

import (
	"crypto/md5"
	"fmt"
	"github.com/dgrijalva/jwt-go"
)

type userClaims struct {
	jwt.StandardClaims
	Username string `json:"name"`
	Identity string `json:"identity"`
}

// 密钥
var myKey = []byte("gin-gorm-oj")

// MD5加密
func GetMd5(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

// 生成token
func GenerateToken(identity string, username string) (string, error) {
	userClaims := &userClaims{
		Username:       username,
		Identity:       identity,
		StandardClaims: jwt.StandardClaims{},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, userClaims)
	tokenString, err := token.SignedString(myKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// 解析token
func ParseToken(tokenString string) (*userClaims, error) {
	userClaim := new(userClaims)
	claims, err := jwt.ParseWithClaims(tokenString, userClaim, func(token *jwt.Token) (interface{}, error) {
		return myKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !claims.Valid {
		return nil, fmt.Errorf("Parse Token Error: %v", err)
	}
	return userClaim, nil
}
