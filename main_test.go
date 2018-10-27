package main_test

import (
	"testing"
	"github.com/op/go-logging"
)

var logger *logging.Logger

func init()  {
	logger = logging.MustGetLogger("test")
}

func Test_main(t *testing.T) {

	//config.Init("","./config/config.yaml")
	////model.AutoMigrate()
	//logger.Infof("%v",config.GetConfig())
	//spider.PostFile(config.GetConfig().XcxUrl,"/Users/leemike/go/src/myproj/try/testspider/images/一年级_20181027/154350_5b582a3631ba8.png","test")

}
