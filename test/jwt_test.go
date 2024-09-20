package test

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"testing"
)

type userClaims struct {
	jwt.StandardClaims
	Username string `json:"name"`
	Identity string `json:"identity"`
}

// 密钥
var myKey = []byte("gin-gorm-oj")

// 生成token
func TestGenerateToken(t *testing.T) {
	userClaims := &userClaims{
		Username:       "liqin",
		Identity:       "user_1",
		StandardClaims: jwt.StandardClaims{},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, userClaims)
	signedString, err := token.SignedString(myKey)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(signedString)
}

// 解析token
func TestParseToken(t *testing.T) {
	tokenString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoibGlxaW4iLCJpZGVudGl0eSI6InVzZXJfMSJ9.Jk3mrlIapikcpZsqunjIvyTfbW9F40EovCGe5VCqoIE"
	userClaim := new(userClaims)
	claims, err := jwt.ParseWithClaims(tokenString, userClaim, func(token *jwt.Token) (interface{}, error) {
		return myKey, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if claims.Valid {
		fmt.Println(claims)
	}
}
