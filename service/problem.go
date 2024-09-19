package service

import (
	"errors"
	"gin_gorm_oj/define"
	"gin_gorm_oj/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strconv"
)

// GetProblemList
// @Summary 问题列表
// @Tags 公共方法
// @Param page query int false "page"
// @Param size query int false "size"
// @Param keyword query string false "keyword"
// @Param category_identity query string false "category_identity"
// @Success 200 {string} json "{"code":"200","msg","","data":""}"
// @Router /problem-list [get]
func GetProblemList(c *gin.Context) {
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
	categoryIdentity := c.Query("category_identity")
	offset := (page - 1) * size
	keyword := c.Query("keyword")

	var count int64
	list := make([]*models.ProblemBasic, 0)
	tx := models.GetProblemList(keyword, categoryIdentity)
	err = tx.Count(&count).Offset(offset).Limit(size).Find(&list).Error
	if err != nil {
		log.Println("GetProblemList error:", err)
		return
	}
	c.JSONP(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"list":  list,
			"count": count,
		},
	})
}

// GetProblemDetail
// @Summary 问题详情
// @Tags 公共方法
// @Param identity query string false "problem identity"
// @Success 200 {string} json "{"code":"200","msg","","data":""}"
// @Router /problem-detail [get]
func GetProblemDetail(c *gin.Context) {
	identity := c.Query("identity")
	if identity == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "问题的唯一标识不能为空",
		})
		return
	}

	problemBasic := new(models.ProblemBasic)
	err := models.DB.Where("identity = ?", identity).Preload("ProblemCategories").
		Preload("ProblemCategories.CategoryBasic").First(&problemBasic).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "问题不存在",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "GetProblemDetail error:" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": problemBasic,
	})

}
