package middlewares

import (
	"gin_gorm_oj/helper"
	"github.com/gin-gonic/gin"
	"net/http"
)

//func AuthAdminCheck() gin.HandlerFunc {
//	return func(c *gin.Context) {
//		//检验用户是否是管理员
//		auth := c.GetHeader("Authorization")
//		userClaim, err := helper.ParseToken(auth)
//		if err != nil {
//			c.Abort()
//			c.JSON(http.StatusOK, gin.H{
//				"code": http.StatusUnauthorized,
//				"msg":  "Unauthorized Authorization",
//			})
//			return
//		}
//		if userClaim == nil || userClaim.IsAdmin != 1 {
//			c.Abort()
//			c.JSON(http.StatusOK, gin.H{
//				"code": http.StatusUnauthorized,
//				"msg":  "Unauthorized Authorization",
//			})
//			return
//		}
//		c.Next()
//	}
//}

func AuthUserCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		//检验用户是否是管理员
		auth := c.GetHeader("Authorization")
		userClaim, err := helper.ParseToken(auth)
		if err != nil {
			c.Abort()
			c.JSON(http.StatusOK, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  "Unauthorized Authorization",
			})
			return
		}
		if userClaim == nil {
			c.Abort()
			c.JSON(http.StatusOK, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  "Unauthorized Authorization",
			})
			return
		}
		c.Set("user", userClaim)
		c.Next()
	}
}
