package model

import (
	_ "database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/op/go-logging"
)
var logger *logging.Logger

func init()  {
	logger = logging.MustGetLogger("question")
}

type Question struct {
	//gorm.Model
	ID        uint `gorm:"primary_key"`
	//题目的文字描述
	Content string
	//题目的图片地址
	QuesPic string
	//答案的图片地址
	AnsPic string
	//年级
	Grade string
	//类型
	QuesType string
	//来源页
	SourcePageUrl string `gorm:"index"`
	AnsContent string
}

func (*Question) TableName() string {
	return "question"
}