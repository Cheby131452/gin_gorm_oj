package service

import (
	"encoding/json"
	"errors"
	"gin_gorm_oj/define"
	"gin_gorm_oj/helper"
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

// ProblemCreate
// @Summary 问题创建
// @Tags 管理员私有方法
// @Param authorization header string true "authorization"
// @Param title formData string true "title"
// @Param content formData string true "content"
// @Param max_runtime formData int false "max_runtime"
// @Param max_mem formData int false "max_mem"
// @Param category_ids formData array false "category_ids"
// @Param test_cases formData array true "test_cases"
// @Success 200 {string} json "{"code":"200","msg","","data":""}"
// @Router /admin/problem-create [post]
func ProblemCreate(c *gin.Context) {
	title := c.PostForm("title")
	content := c.PostForm("content")
	maxRuntime, _ := strconv.Atoi(c.PostForm("max_runtime"))
	maxMem, _ := strconv.Atoi(c.PostForm("max_mem"))
	categoryIds := c.PostFormArray("category_ids")
	testCases := c.PostFormArray("test_cases")
	if len(testCases) == 0 || title == "" || content == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数不能为空",
		})
		return
	}

	identity := helper.GetUUID()
	data := &models.ProblemBasic{
		Title:      title,
		Identity:   identity,
		Content:    content,
		MaxRuntime: maxRuntime,
		MaxMem:     maxMem,
	}

	//处理分类
	problemCategories := make([]*models.ProblemCategory, 0)
	for _, id := range categoryIds {
		categoryId, _ := strconv.Atoi(id)
		problemCategories = append(problemCategories, &models.ProblemCategory{
			ProblemId:  data.ID,
			CategoryId: uint(categoryId),
		})
	}
	data.ProblemCategories = problemCategories

	//处理测试用例
	testCaseBasics := make([]*models.TestCase, 0)
	for _, t := range testCases {
		//example: {"input":"1 2\n","output":"3\n"}
		caseMap := make(map[string]string)
		err := json.Unmarshal([]byte(t), &caseMap)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "测试用例格式错误",
			})
			return
		}
		if _, ok := caseMap["input"]; !ok {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "测试用例格式错误",
			})
			return
		}
		if _, ok := caseMap["output"]; !ok {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "测试用例格式错误",
			})
			return
		}
		testCaseBasics = append(testCaseBasics, &models.TestCase{
			Identity:        helper.GetUUID(),
			ProblemIdentity: identity,
			Input:           caseMap["input"],
			Output:          caseMap["output"],
		})
	}
	data.TestCases = testCaseBasics

	//创建问题
	err := models.DB.Create(data).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "ProblemCreate error:" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"identity": identity,
		},
	})
}

// ProblemModify
// @Summary 问题修改
// @Tags 管理员私有方法
// @Param authorization header string true "authorization"
// @Param identity formData string true "identity"
// @Param title formData string true "title"
// @Param content formData string true "content"
// @Param max_runtime formData int false "max_runtime"
// @Param max_mem formData int false "max_mem"
// @Param category_ids formData []string false "category_ids" collectionFormat(multi)
// @Param test_cases formData []string true "test_cases" collectionFormat(multi)
// @Success 200 {string} json "{"code":"200","msg","","data":""}"
// @Router /admin/problem-modify [put]
func ProblemModify(c *gin.Context) {
	identity := c.PostForm("identity")
	title := c.PostForm("title")
	content := c.PostForm("content")
	maxRuntime, _ := strconv.Atoi(c.PostForm("max_runtime"))
	maxMem, _ := strconv.Atoi(c.PostForm("max_mem"))
	categoryIds := c.PostFormArray("category_ids")
	testCases := c.PostFormArray("test_cases")
	if len(testCases) == 0 || title == "" || content == "" || identity == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数不能为空",
		})
		return
	}

	//利用事务进行更新
	if err := models.DB.Transaction(func(tx *gorm.DB) error {
		//问题基础信息的保存
		problemBasic := &models.ProblemBasic{
			Identity:   identity,
			Title:      title,
			Content:    content,
			MaxRuntime: maxRuntime,
			MaxMem:     maxMem,
		}
		err := tx.Where("identity = ?", identity).Updates(problemBasic).Error
		if err != nil {
			return err
		} //返回错误以触发回滚

		//查询问题详情
		err = tx.Where("identity = ?", identity).Find(problemBasic).Error
		if err != nil {
			return err
		}

		//关联问题分类的更新
		//1.删除已存在的关联关系
		err = tx.Where("problem_id = ?", problemBasic.ID).Delete(new(models.ProblemCategory)).Error
		if err != nil {
			return err
		}
		//2.新增关联关系
		pcs := make([]*models.ProblemCategory, 0)
		for _, id := range categoryIds {
			intId, _ := strconv.Atoi(id)
			pcs = append(pcs, &models.ProblemCategory{
				ProblemId:  problemBasic.ID,
				CategoryId: uint(intId),
			})
		}
		err = tx.Create(pcs).Error
		if err != nil {
			return err
		}

		//关联测试案例的更新
		//1.删除已存在的关联关系
		err = tx.Where("problem_identity = ?", problemBasic.ID).Delete(&models.ProblemBasic{}).Error
		if err != nil {
			return err
		}
		//2.新增关联关系
		tcs := make([]*models.TestCase, 0)
		for _, t := range testCases {
			caseMap := make(map[string]string)
			caseIdentity := helper.GetUUID()
			err = json.Unmarshal([]byte(t), &caseMap)
			if err != nil {
				return err
			}
			if _, ok := caseMap["input"]; !ok {
				return errors.New("测试案例input格式错误")
			}
			if _, ok := caseMap["output"]; !ok {
				return errors.New("测试案例output格式错误")
			}
			tcs = append(tcs, &models.TestCase{
				Identity:        caseIdentity,
				ProblemIdentity: problemBasic.Identity,
				Input:           caseMap["input"],
				Output:          caseMap["output"],
			})
		}
		err = tx.Create(tcs).Error
		if err != nil {
			return err
		}
		return nil //返回nil以提交事务
	}); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "ProblemModify error:" + err.Error(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "问题修改成功",
	})
}
