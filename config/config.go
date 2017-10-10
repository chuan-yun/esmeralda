package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var Esmeralda struct {
	Config *viper.Viper
}

func log() {

	content, err := ioutil.ReadFile("esmeralda.log")
	if err != nil {
		return nil, err
	}
	cfg, err := Load(string(content))
	if err != nil {
		return nil, err
	}
	resolveFilepaths(filepath.Dir(filename), cfg)

	logrus.SetFormatter(&logrus.JSONFormatter{})

	file, err := os.OpenFile("esmeralda.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err == nil {
		logrus.SetOutput(file)
	} else {
		logrus.SetOutput(os.Stdout)
		logrus.Fatal("Failed to log to file, using default stderr")
	}

	logrus.Debug("Hello World!")
}

func config() {
	viper.SetEnvPrefix("esmeralda")

	viper.SetConfigType("toml")
	viper.SetConfigName("esmeralda")
	viper.AddConfigPath("/etc/chuanyun/")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		logrus.Panic(err)
	}
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
	})
	logrus.Info(viper.AllSettings())
}
