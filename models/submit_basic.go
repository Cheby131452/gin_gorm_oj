package models

import "gorm.io/gorm"

type SubmitBasic struct {
	gorm.Model
	Identity        string        `gorm:"column:identity;type:varchar(36);" json:"identity"`                 //唯一标识
	ProblemIdentity string        `gorm:"column:problem_identity;type:varchar(36);" json:"problem_identity"` //问题的唯一标识
	UserIdentity    string        `gorm:"column:user_identity;type:varchar(36);" json:"user_identity"`       //用户的唯一标识
	Path            string        `gorm:"column:path;type:varchar(255);" json:"path"`                        //代码路径
	Status          int           `gorm:"column:status;type:tinyint(1);" json:"status"`                      //【-1-待判断，1-答案正确，2-答案错误，3-运行超时，4-运行超内存，5-编译错误】
	ProblemBasic    *ProblemBasic `gorm:"foreignKey:identity;references:problem_identity"`                   //关联问题基础表
	UserBasic       *UserBasic    `gorm:"foreignKey:identity;references:user_identity"`                      //关联用户基础表
}

func (table *SubmitBasic) TableName() string {
	return "submit_basic"
}

func GetSubmitList(problemIdentity, userIdentity string, status int) *gorm.DB {
	tx := DB.Model(new(SubmitBasic)).Preload("ProblemBasic", func(db *gorm.DB) *gorm.DB {
		return db.Omit("content") //省略掉problem_basic表中的content字段内容
	}).Preload("UserBasic")
	if problemIdentity != "" {
		tx.Where("problem_identity = ?", problemIdentity)
	}
	if userIdentity != "" {
		tx.Where("user_identity = ?", userIdentity)
	}
	if status != 0 {
		tx.Where("status = ?", status)
	}
	return tx
}
