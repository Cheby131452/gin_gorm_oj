package service

import (
	"bytes"
	"errors"
	"gin_gorm_oj/define"
	"gin_gorm_oj/helper"
	"gin_gorm_oj/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// GetSubmitList
// @Summary 提交列表
// @Tags 公共方法
// @Param page query int false "page"
// @Param size query int false "size"
// @Param problem_identity query string false "problem_identity"
// @Param user_identity query string false "user_identity"
// @Param status query int false "status"
// @Success 200 {string} json "{"code":"200","msg","","data":""}"
// @Router /submit-list [get]
func GetSubmitList(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", define.DefaultPage))
	if err != nil {
		log.Println("GetSubmitList page error:", err)
		return
	}
	size, err := strconv.Atoi(c.DefaultQuery("size", define.DefaultSize))
	if err != nil {
		log.Println("GetSubmitList size error:", err)
		return
	}
	offset := (page - 1) * size
	var count int64
	list := make([]models.SubmitBasic, 0)

	problemIdentity := c.Query("problem_identity")
	userIdentity := c.Query("user_identity")
	status, _ := strconv.Atoi(c.Query("status"))
	tx := models.GetSubmitList(problemIdentity, userIdentity, status)
	err = tx.Count(&count).Offset(offset).Limit(size).Find(&list).Error
	if err != nil {
		log.Println("GetSubmitList error:", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "GetSubmitList error:" + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"list":  list,
			"count": count},
	})
}

// Submit
// @Summary 代码提交
// @Tags 用户私有方法
// @Param authorization header string true "authorization token"
// @Param problem_identity query string true "problem_identity"
// @Param code body string true "code"
// @Success 200 {string} json "{"code":"200","msg","","data":""}"
// @Router /user/submit [post]
func Submit(c *gin.Context) {
	problemIdentity := c.Query("problem_identity")
	codeByte, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Get Request Body error:" + err.Error(),
		})
		return
	}
	//保存代码
	path, err := helper.CodeSave(codeByte)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "CodeSave error:" + err.Error(),
		})
		return
	}

	u, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Get User Claim error:" + err.Error(),
		})
		return
	}
	userClaim := u.(*helper.UserClaims)
	sb := &models.SubmitBasic{
		Identity:        helper.GetUUID(),
		ProblemIdentity: problemIdentity,
		UserIdentity:    userClaim.Identity,
		Path:            path,
	}

	//代码判断
	pb := new(models.ProblemBasic)
	err = models.DB.Where("identity = ?", problemIdentity).Preload("TestCases").First(pb).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Get ProblemBasic error:" + err.Error(),
		})
		return
	}
	//答案错误的channel
	wa := make(chan int)
	//超内存的channel
	oom := make(chan int)
	//编译错误的channel
	ce := make(chan int)
	//通过个数
	passCount := 0
	var lock sync.Mutex
	//提示信息
	var msg string = "测试通过"
	for _, v := range pb.TestCases {
		testCase := v
		go func() {
			//执行测试
			cmd := exec.Command("go", "run", path)
			var out, stderr bytes.Buffer
			cmd.Stderr = &stderr
			cmd.Stdout = &out
			stdinPipe, err := cmd.StdinPipe()
			if err != nil {
				log.Fatal(err)
			}
			io.WriteString(stdinPipe, testCase.Input)
			var bm runtime.MemStats
			runtime.ReadMemStats(&bm)
			//根据测试的输入案例进行运行，拿到输出的结果和标准的输出结果比对
			if err = cmd.Run(); err != nil {
				log.Println(err, stderr.String())
				if err.Error() == "exit status 2" {
					msg = stderr.String()
					ce <- 1
					return
				}
			}
			var em runtime.MemStats
			runtime.ReadMemStats(&em)
			//答案错误
			if testCase.Output != out.String() {
				msg = "答案错误"
				wa <- 1
				return
			}
			//运行超内存
			if (em.Alloc-bm.Alloc)/1024 > uint64(pb.MaxMem) {
				msg = "运行超内存"
				oom <- 1
			}

			lock.Lock()
			passCount++
			lock.Unlock()
		}()
	}
	select {
	//-1-待判断，1-答案正确，2-答案错误，3-运行超时，4-运行超内存，5-编译错误
	case <-wa:
		sb.Status = 2
	case <-oom:
		sb.Status = 4
	case <-ce:
		sb.Status = 5
	case <-time.After(time.Millisecond * time.Duration(pb.MaxRuntime)):
		if passCount == len(pb.TestCases) {
			sb.Status = 1
		} else {
			sb.Status = 3
		}
	}

	//开启事务，更新用户完成问题个数和提交次数
	err = models.DB.Transaction(func(tx *gorm.DB) error {
		err = tx.Create(sb).Error
		if err != nil {
			return errors.New("SubmitBasic Save  error:" + err.Error())
		}
		ub := new(models.UserBasic)
		err = tx.Where("identity = ?", userClaim.Identity).First(ub).Error
		if err != nil {
			return errors.New("UserBasic Find  error:" + err.Error())
		}
		ub.SubmitNum++
		if sb.Status == 1 {
			ub.FinishProblemNum++
		}
		err = tx.Save(ub).Error
		if err != nil {
			return errors.New("UserBasic Save  error:" + err.Error())
		}
		return nil //commit
	})

	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Submit error:" + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"status": sb.Status,
			"msg":    msg,
		},
	})
}
