package models

import "gorm.io/gorm"

type Problem struct {
	gorm.Model
	Identity   string `gorm:"column:identity;type:varchar(36);" json:"identity"`        //问题表的唯一标识
	CategoryId string `gorm:"column:category_id;type:varchar(255);" json:"category_id"` //分类ID，逗号分隔
	Title      string `gorm:"column:title;type:varchar(255);" json:"title"`             //文章标题
	Content    string `gorm:"column:content;type:text;" json:"content"`                 //文章正文
	MaxMem     string `gorm:"column:max_men;type:int(11);" json:"max_men"`              //最大运行内存
	MaxRuntime string `gorm:"column:max_runtime;type:int(11);" json:"max_runtime"`      //最大运行时间
}

func (table *Problem) TableName() string {
	return "problem"
}
