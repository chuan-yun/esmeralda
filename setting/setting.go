package setting

import (
	"path/filepath"

	"chuanyun.io/esmeralda/util"
	"github.com/spf13/viper"
)

var Settings struct {
	Log           LogSettings
	Web           WebSettings
	Application   ApplicationSettings
	Elasticsearch ElasticsearchSettings
}

func Initialize(configFilePath string) {
	ReadConfigFile(configFilePath)
	LogInitialize()
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
}
