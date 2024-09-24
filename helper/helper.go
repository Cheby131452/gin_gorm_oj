package helper

import (
	"crypto/md5"
	"crypto/tls"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/jordan-wright/email"
	uuid "github.com/satori/go.uuid"
	"math/rand"
	"net/smtp"
	"strconv"
	"time"
)

type userClaims struct {
	jwt.StandardClaims
	Username string `json:"name"`
	Identity string `json:"identity"`
	IsAdmin  int    `json:"is_admin"`
}

// 密钥
var myKey = []byte("gin-gorm-oj")

// MD5加密
func GetMd5(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

// 生成token
func GenerateToken(identity string, username string, isAdmin int) (string, error) {
	userClaims := &userClaims{
		StandardClaims: jwt.StandardClaims{},
		Username:       username,
		Identity:       identity,
		IsAdmin:        isAdmin,
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

// 发送邮箱验证码
func SendCode(toUserEmail string, code string) error {
	e := email.NewEmail()
	e.From = "Lijr <1643804185@qq.com>"
	e.To = []string{toUserEmail}
	e.Subject = "验证码已发送，请查收"
	e.HTML = []byte("您的验证码是: <b>" + code + "</b>")
	//err := e.Send("smtp.qq.com:465", smtp.PlainAuth("", "1643804185@qq.com", "password123", "smtp.qq.com"))
	//返回 EOF 时，关闭SSL重试
	err := e.SendWithTLS("smtp.qq.com:465", smtp.PlainAuth("", "1643804185@qq.com", "czcukzmsqsifchag", "smtp.qq.com"),
		&tls.Config{InsecureSkipVerify: true, ServerName: "smtp.qq.com"})
	return err
}

// 生成UUID
func GetUUID() string {
	return uuid.NewV4().String()
}

// 生成验证码
func GenValidateCode() string {
	rand.Seed(time.Now().UnixNano())
	s := ""
	for i := 1; i <= 6; i++ {
		s += strconv.Itoa(rand.Intn(10))
	}
	return s
}
