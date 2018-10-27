package model

import (
	"github.com/jinzhu/gorm"
	"time"
	"myproj/try/testspider/config"
	"github.com/op/go-logging"
)

func init()  {
	logger = logging.MustGetLogger("db")
}

var db *gorm.DB

func AutoMigrate() *gorm.DB {
	db := CreateConn()
	logger.Debugf("Start migrating database automatically")
	if err := db.Debug().Exec("set transaction isolation level serializable").AutoMigrate(
		&Question{},
	).Error; err != nil {
		logger.Panicf("Error auto-migrating database : %s", err)
	}
	return db
}

func CreateConn() *gorm.DB {
	var err error
	if db == nil {
		dbc := config.Globalconfig.DBConfig
		db, err = gorm.Open("mysql", dbc.User+":"+dbc.Password+"@tcp("+dbc.Address+":"+dbc.Port+")/"+dbc.DBName+"?charset=utf8&parseTime=true")
		if err != nil {
			logger.Error(err)
			return nil
		}
		db.DB().SetMaxOpenConns(0)
		db.DB().SetMaxIdleConns(0)
		db.DB().SetConnMaxLifetime(10 * time.Minute)
	}
	return db
}