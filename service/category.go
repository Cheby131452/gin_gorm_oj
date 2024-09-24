package service

import (
	"gin_gorm_oj/define"
	"gin_gorm_oj/helper"
	"gin_gorm_oj/models"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)

// GetCategoryList
// @Summary 分类列表
// @Tags 管理员私有方法
// @Param authorization header string true "authorization token"
// @Param page query int false "page"
// @Param size query int false "size"
// @Param keyword query string false "keyword"
// @Success 200 {string} json "{"code":"200","msg","","data":""}"
// @Router /admin/category-list [get]
func GetCategoryList(c *gin.Context) {
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
	keyword := c.Query("keyword")
	var count int64
	categoryList := make([]*models.CategoryBasic, 0)
	err = models.DB.Model(new(models.CategoryBasic)).Where("name like ?", "%"+keyword+"%").Count(&count).
		Offset(offset).Find(&categoryList).Error
	if err != nil {
		log.Println("GetCategoryList error:", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "获取分类列表失败",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"list":  categoryList,
			"count": count,
		},
	})
}

// CategoryCreate
// @Summary 分类创建
// @Tags 管理员私有方法
// @Param authorization header string true "authorization token"
// @Param name formData string true "name"
// @Param parentId formData int false "parentId"
// @Success 200 {string} json "{"code":"200","msg","","data":""}"
// @Router /admin/category-create [post]
func CategoryCreate(c *gin.Context) {
	name := c.PostForm("name")
	parentId, _ := strconv.Atoi(c.PostForm("parentId"))
	identity := helper.GetUUID()
	category := &models.CategoryBasic{
		Name:     name,
		ParentID: uint(parentId),
		Identity: identity,
	}
	err := models.DB.Create(category).Error
	if err != nil {
		log.Println("CategoryCreate error:", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "创建分类失败",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "分类创建成功",
	})
}

// CategoryDelete
// @Summary 分类删除
// @Tags 管理员私有方法
// @Param authorization header string true "authorization token"
// @Param identity query string true "identity"
// @Success 200 {string} json "{"code":"200","msg","","data":""}"
// @Router /admin/category-delete [delete]
func CategoryDelete(c *gin.Context) {
	identity := c.Query("identity")
	if identity == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数不正确",
		})
		return
	}

	var count int64
	err := models.DB.Model(new(models.ProblemCategory)).
		Where("category_id = (SELECT id FROM category_basic WHERE identity = ? LIMIT 1)", identity).Count(&count).Error
	if err != nil {
		log.Println("Get ProblemCategory error:", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "获取分类关联的问题失败",
		})
		return
	}
	if count > 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "该分类下存在问题，不可删除",
		})
		return
	}

	err = models.DB.Where("identity = ?", identity).Unscoped().Delete(new(models.CategoryBasic)).Error
	if err != nil {
		log.Printf("Delete CategoryBasic error: %v\n", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "分类删除失败",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "分类删除成功",
	})
}

// CategoryModify
// @Summary 分类修改
// @Tags 管理员私有方法
// @Param authorization header string true "authorization token"
// @Param identity formData string true "identity"
// @Param name formData string true "name"
// @Param parentId formData int false "parentId"
// @Success 200 {string} json "{"code":"200","msg","","data":""}"
// @Router /admin/category-modify [put]
func CategoryModify(c *gin.Context) {
	identity := c.PostForm("identity")
	name := c.PostForm("name")
	parentId, _ := strconv.Atoi(c.PostForm("parentId"))
	if name == "" || identity == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数不正确",
		})
		return
	}
	category := &models.CategoryBasic{
		Name:     name,
		ParentID: uint(parentId),
		Identity: identity,
	}
	err := models.DB.Model(new(models.CategoryBasic)).Where("identity = ?", identity).Updates(category).Error
	if err != nil {
		log.Println("CategoryCreate error:", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "分类修改失败",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "分类修改成功",
	})
}
