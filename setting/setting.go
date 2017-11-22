package setting

import (
	"path/filepath"

	"chuanyun.io/esmeralda/util"
	"github.com/spf13/viper"
)

var v *viper.Viper

func Initialize(configFilePath string) {
	ReadConfigFile(configFilePath)
	LogSettingInitialize()
}

func ReadConfigFile(configFilePath string) {
	configFilePath, err := filepath.Abs(filepath.Clean(configFilePath))
	if err != nil {
		panic(util.Message(err.Error()))
	}

	v.SetConfigFile(configFilePath)
	v.SetConfigType("toml")

	err = v.ReadInConfig()
	if err != nil {
		panic(util.Message(err.Error()))
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
