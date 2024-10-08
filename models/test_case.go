package models

import "gorm.io/gorm"

// 测试用例
type TestCase struct {
	gorm.Model
	Identity        string `gorm:"column:identity;type:varchar(36)" json:"identity"`
	ProblemIdentity string `gorm:"column:problem_identity;type:varchar(36)" json:"problem_identity"`
	Input           string `gorm:"column:input;type:text" json:"input"`
	Output          string `gorm:"column:output;type:text" json:"output"`
}

func (table *TestCase) TableName() string {
	return "test_case"
}
