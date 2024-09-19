package service

import (
	"errors"
	"gin_gorm_oj/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
