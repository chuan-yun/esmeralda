package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {

	// type Configuration struct {
	// 	Clusters    []Cluster `json:"clusters"`
	// 	MinReplicas int       `json:"min_replicas"`
	// 	MaxReplicas int       `json:"max_replicas"`
	// }

	// logrus.Info(viper.Get("elasticsearch"))

	config()
	log()

	fmt.Println(os.Args[0])

	dir := filepath.Dir(os.Args[0])
	fmt.Print("Dir=")
	fmt.Println(dir)

	dir, err := filepath.Abs(filepath.Clean(filepath.Dir(os.Args[0])))
	if err != nil {
		logrus.Fatal(err)
	}
	fmt.Print("Abs=")
	fmt.Println(dir)

	dir, err = os.Getwd()
	fmt.Print("Wd=")
	fmt.Println(dir)
}

func log() {

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

func config() {
	viper.SetEnvPrefix("esmeralda")
	viper.AutomaticEnv()

	viper.SetConfigType("toml")
	viper.SetConfigName("esmeralda")
	viper.AddConfigPath("/etc/chuanyun/")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Panic("error occurred during config initialization")
	}
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
		logrus.WithFields(logrus.Fields{
			"settings": viper.AllSettings(),
		}).Info("all user settings")
	})

	logrus.WithFields(logrus.Fields{
		"settings": viper.AllSettings(),
	}).Info("all user settings")
}
