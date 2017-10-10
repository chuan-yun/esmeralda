package main

import (
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	log()
	config()

	logrus.Info("main")

	logrus.Info(viper.Get("elasticsearch"))
}

func log() {
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

	viper.SetConfigType("toml")           // or viper.SetConfigType("YAML")
	viper.SetConfigName("esmeralda")      // name of config file (without extension)
	viper.AddConfigPath("/etc/chuanyun/") // path to look for the config file in
	viper.AddConfigPath(".")              // optionally look for config in the working directory
	err := viper.ReadInConfig()           // Find and read the config file
	if err != nil {                       // Handle errors reading the config file
		logrus.Panic(err)
	}
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
	})
	logrus.Info(viper.AllSettings())
}
