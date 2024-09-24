package service

import (
	"errors"
	"gin_gorm_oj/define"
	"gin_gorm_oj/helper"
	"gin_gorm_oj/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strconv"
	"time"
)

// GetUserDetail
// @Summary 用户详情
// @Tags 公共方法
// @Param identity query string false "user identity"
// @Success 200 {string} json "{"code":"200","msg","","data":""}"
// @Router /user-detail [get]
func GetUserDetail(c *gin.Context) {
	identity := c.Query("identity")
	if identity == "" {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "用户标识不能为空",
		})
		return
	}

	userBasic := models.UserBasic{}
	err := models.DB.Omit("password").Where("identity = ?", identity).First(&userBasic).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(200, gin.H{
				"code": -1,
				"msg":  "没有该用户",
			})
			return
		}

		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "GetUserDetail error:" + err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"code": 200,
		"data": userBasic,
	})
}

// Login
// @Summary 用户登录
// @Tags 公共方法
// @Param username formData string false "username"
// @Param password formData string false "password"
// @Success 200 {string} json "{"code":"200","msg","","data":""}"
// @Router /login [post]
func Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	if username == "" || password == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "必填信息为空",
		})
		return
	}
	//md5加密
	password = helper.GetMd5(password)

	user := new(models.UserBasic)
	err := models.DB.Where("username = ? AND password = ?", username, password).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "用户名或密码错误",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Get UserBasic error:" + err.Error(),
		})
		return
	}
	//生成token
	token, err := helper.GenerateToken(user.Identity, user.Username, user.IsAdmin)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "GenerateToken error:" + err.Error(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"token": token,
		},
	})
}

// SendCode
// @Summary 发送验证码
// @Tags 公共方法
// @Param mail formData string true "email"
// @Success 200 {string} json "{"code":"200","msg","","data":""}"
// @Router /send-code [post]
func SendCode(c *gin.Context) {
	mail := c.PostForm("mail")
	if mail == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "邮箱不能为空",
		})
		return
	}
	code := helper.GenValidateCode()
	//将生成的验证码保存在Redis中，有效期60s
	models.Redis.Set(c, mail, code, time.Second*60)
	err := helper.SendCode(mail, code)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "SendCode error:" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "验证码发送成功",
	})
}

// Register
// @Summary 用户注册
// @Tags 公共方法
// @Param mail formData string true "mail"
// @Param username formData string true "username"
// @Param password formData string true "password"
// @Param phone formData string false "phone"
// @Param code formData string true "code"
// @Success 200 {string} json "{"code":"200","msg","","data":""}"
// @Router /register [post]
func Register(c *gin.Context) {
	userCode := c.PostForm("code")
	mail := c.PostForm("mail")
	username := c.PostForm("username")
	password := c.PostForm("password")
	phone := c.PostForm("phone")
	if mail == "" || username == "" || password == "" || userCode == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数不正确",
		})
	}
	//验证码是否正确
	sysCode, err := models.Redis.Get(c, mail).Result()
	if err != nil {
		log.Printf("Get Code error: %v\n", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "验证码不正确，请重新获取验证码",
		})
		return
	}

	if sysCode != userCode {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "验证码不正确",
		})
		return
	}

	//判断邮箱是否已存在
	var count int64
	err = models.DB.Model(new(models.UserBasic)).Where("mail = ?", mail).Count(&count).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Get User error:" + err.Error(),
		})
		return
	}
	if count > 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "该邮箱已经注册",
		})
		return
	}

	//将用户信息插入数据库
	userIdentity := helper.GetUUID()
	data := &models.UserBasic{
		Identity: userIdentity,
		Username: username,
		Password: helper.GetMd5(password),
		Phone:    phone,
		Mail:     mail,
	}
	err = models.DB.Create(data).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Create User error:" + err.Error(),
		})
		return
	}

	//生成token,返回给前端
	token, err := helper.GenerateToken(userIdentity, username, data.IsAdmin)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "GenerateToken error:" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"token": token,
		},
	})
}

// GetRankList
// @Summary 用户排行榜
// @Tags 公共方法
// @Param page query int false "page"
// @Param size query int false "size"
// @Success 200 {string} json "{"code":"200","msg","","data":""}"
// @Router /rank-list [get]
func GetRankList(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", define.DefaultPage))
	if err != nil {
		log.Println("GetProblemList page error:", err)
		return
	}
	size, err := strconv.Atoi(c.DefaultQuery("size", define.DefaultSize))
	if err != nil {
		log.Println("GetProblemList size error:", err)
		return
	}
	offset := (page - 1) * size
	var count int64
	list := make([]*models.UserBasic, 0)
	err = models.DB.Model(new(models.UserBasic)).Count(&count).Order("finish_problem_num DESC, submit_num ASC").
		Offset(offset).Limit(size).Find(&list).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "GetRankList error:" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"list":  list,
			"count": count,
		},
	})
}
