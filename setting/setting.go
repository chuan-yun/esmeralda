package config

import (
	"os"
	"path/filepath"

	"chuanyun.io/esmeralda/util"
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	PROD    string = "production"
	STAGING string = "staging"
	TEST    string = "testing"
	DEV     string = "development"
)

var Application struct {
	Env string
}

var Log struct {
	Level string
	Path  string
}

var Api struct {
	Port   int64
	Prefix string
}

var Exporter struct {
	Port   int64
	Prefix string
}

func logsettings() {

	// content, err := ioutil.ReadFile("esmeralda.log")
	// if err != nil {
	// 	return nil, err
	// }
	// cfg, err := Load(string(content))
	// if err != nil {
	// 	return nil, err
	// }
	// resolveFilepaths(filepath.Dir(filename), cfg)

	filepath.Base("/a/b.c")

	logrus.SetFormatter(&logrus.JSONFormatter{})

	file, err := os.OpenFile("esmeralda.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err == nil {
		logrus.SetOutput(file)
	} else {
		logrus.SetOutput(os.Stdout)
		logrus.Fatal("Failed to log to file, using default stderr")
	}
	defer file.Close()

	logrus.Debug("Hello World!")
}

func ReadConfigFile(in string) {
	in, err := filepath.Abs(filepath.Clean(in))
	if err != nil {
		panic(util.Message(err.Error()))
	}

	viper.SetConfigFile(in)
	viper.SetConfigType("toml")

	err = viper.ReadInConfig()
	if err != nil {
		panic(util.Message(err.Error()))
	}
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		logrus.WithFields(logrus.Fields{
			"filename": e.Name,
		}).Info("Config file changed:")
	})

	for true {

	}
}

// func Initialize(configFilePath string) {
// 	configFilePath, err := filepath.Abs(filepath.Clean(configFilePath))
// 	if err != nil {
// 		panic(util.Message(err.Error()))
// 	}

// 	viper.SetConfigFile(configFilePath)
// 	viper.SetConfigType("toml")

// 	err = viper.ReadInConfig()
// 	if err != nil {
// 		panic(util.Message(err.Error()))
// 	}
// 	Esmeralda.Config = viper.GetViper()

// 	fmt.Println(Esmeralda.Config.GetString("log.path"))

// 	logrus.SetFormatter(&logrus.JSONFormatter{})

// 	file, err := os.OpenFile(Esmeralda.Config.GetString("log.path"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
// 	if err == nil {
// 		logrus.SetOutput(file)
// 	} else {
// 		panic(util.Message(err.Error()))
// 	}

// 	logrus.WithFields(logrus.Fields{
// 		"setting": viper.AllSettings(),
// 	}).Info("setting")

// 	Esmeralda.Config.WatchConfig()
// 	Esmeralda.Config.OnConfigChange(func(e fsnotify.Event) {
// 		logrus.WithFields(logrus.Fields{
// 			// "filename": e.Name,
// 			"string": e,
// 			// "op":       e.Op,
// 		}).Info("Config file changed:")

// 		// err = viper.ReadInConfig()
// 		// if err != nil {
// 		// 	panic(util.Message(err.Error()))
// 		// }

// 		// logrus.WithFields(logrus.Fields{
// 		// 	"setting": viper.AllSettings(),
// 		// }).Info("setting")
// 	})

// 	for true {

// 	}
// }
