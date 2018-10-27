package config

import (
	"github.com/spf13/viper"
	"strings"
	"fmt"
	"github.com/op/go-logging"
)

type Config struct {
	DBConfig DBConfig
	SavePath string
	RootUrl string
	XcxUrl string
	TestPageUrl string

}

type DBConfig struct {
	Address string
	Port string
	User string
	DBName string
	Password string
}



var logger *logging.Logger

var Globalconfig *Config

var defaultconfigpath = "./testspider/config/config.yaml"


func GetConfig() *Config {
	if Globalconfig != nil{
		return Globalconfig
	}
	Init("web",defaultconfigpath)
	return Globalconfig
}

func init()  {
	logger = logging.MustGetLogger("config")
}

func Init(envPrefix, path string) {
	Globalconfig = &Config{}
	if err := ParseConfig(envPrefix, path, Globalconfig); err != nil {
		logger.Panic("Error parsing config file [%s] : %s", path, err)
	}
}

func ParseConfig(envPrefix, ymlFile string, object interface{}) error {
	v := newViper(envPrefix)
	v.SetConfigFile(ymlFile)
	err := v.ReadInConfig()
	if err != nil {
		return fmt.Errorf("read in config error: %s", err)
	}
	err = v.Unmarshal(object)
	if err != nil {
		return fmt.Errorf("unmarshal config to object error: %s", err)
	}
	return nil
}

func newViper(envPrefix string) *viper.Viper {
	v := viper.New()
	v.SetEnvPrefix(envPrefix)
	v.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	v.SetEnvKeyReplacer(replacer)
	return v
}