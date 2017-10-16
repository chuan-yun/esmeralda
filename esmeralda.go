package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	version = flag.Bool("version", false, "output version information and exit")
	help    = flag.Bool("help", false, "output help information and exit")
	config  = flag.String("config", "/etc/chuanyun/esmeralda.toml", "config file path")

	GitTag    = "2000.01.01.release"
	BuildTime = "2000-01-01T00:00:00+0800"
)

func PrintVersionInfo() {
	fmt.Println("esmeralda")
	fmt.Println("version: " + GitTag + ", build: " + BuildTime)
	fmt.Println("Copyright (c) 2017, chuanyun.io. All rights reserved.")
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()

	if *version {
		PrintVersionInfo()
		os.Exit(0)
	}

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	fmt.Println(*config)

	dir := filepath.Dir(*config)
	fmt.Print("Dir=")
	fmt.Println(dir)

	dir, err := filepath.Abs(filepath.Clean(filepath.Dir(*config)))
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

func Config() {
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
		logrus.WithFields(logrus.Fields{
			"filename": e.Name,
		}).Info("Config file changed:")
	})

	logrus.WithFields(logrus.Fields{
		"settings": viper.AllSettings(),
	}).Info("all user settings")
}

beijingchewu009