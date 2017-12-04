package setting

import (
	"path/filepath"

	"github.com/sirupsen/logrus"

	"chuanyun.io/esmeralda/util"
	"github.com/spf13/viper"
)

type Validator interface {
	Validate() string
}

var Settings struct {
	Application    ApplicationSettings
	ConfigFilePath string
	Elasticsearch  ElasticsearchSettings
	Log            LogSettings
	Web            WebSettings
}

func Initialize(configFilePath string) {
	ReadConfigFile(configFilePath)
	LogInitialize()

	logrus.WithFields(logrus.Fields{
		"settings": Settings,
	}).Info("Initialize settings completed")

	ValidateWebSettings()
}

func ReadConfigFile(configFilePath string) {
	configFilePath, err := filepath.Abs(filepath.Clean(configFilePath))
	if err != nil {
		panic(util.Message(err.Error()))
	}

	viper.SetConfigFile(configFilePath)
	viper.SetConfigType("toml")
	err = viper.ReadInConfig()
	if err != nil {
		panic(util.Message(err.Error()))
	}

	err = viper.Unmarshal(&Settings)
	if err != nil {
		panic(util.Message(err.Error()))
	}

	Settings.ConfigFilePath = configFilePath
}
