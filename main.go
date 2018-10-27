package main

import (
	"github.com/op/go-logging"
	"os"
	"myproj/try/testspider/model"
	"myproj/try/testspider/config"
	"myproj/try/testspider/core/spider"
)

var logger *logging.Logger

func init()  {
	stdoutBackend := logging.NewBackendFormatter(
		logging.NewLogBackend(os.Stdout, "", 0),
		logging.MustStringFormatter(`%{color}[%{time:2006-01-02 15:04:05.000}] [%{module}] <%{shortfile}> %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`),
	)
	logging.SetBackend(stdoutBackend)
	logger = logging.MustGetLogger("main")
}


func main()  {
	var err error
	config.Init("","./testspider/config/config.yaml")
	model.AutoMigrate()
	err = spider.Start()
	if err != nil{
		logger.Info(err)
	}

}