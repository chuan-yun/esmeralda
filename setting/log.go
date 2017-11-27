package setting

import (
	"fmt"
	"os"
	"path/filepath"

	"chuanyun.io/esmeralda/util"
	"github.com/sirupsen/logrus"
)

type LogSettings struct {
	Level string
	Path  string
}

func LogInitialize() {

	logPath, err := filepath.Abs(Settings.Log.Path)
	if err != nil {
		panic(util.Message(err.Error()))
	}

	Settings.Log.Path = logPath

	logFile, err := os.OpenFile(Settings.Log.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		panic(util.Message(err.Error()))
	}
	logrus.SetOutput(logFile)
	handler := func() {
		logFile.Close()
	}
	logrus.RegisterExitHandler(handler)

	level, err := logrus.ParseLevel(Settings.Log.Level)
	if err != nil {
		level = logrus.InfoLevel
		Settings.Log.Level = fmt.Sprintf("%s", logrus.InfoLevel)
	}
	logrus.SetLevel(level)

	logrus.SetFormatter(&logrus.JSONFormatter{})

	logrus.Info("@@@@@@@@@@@@@@@@@@@@")

	logrus.WithFields(logrus.Fields{
		"log.formatter": "JSONFormatter",
		"log.path":      Settings.Log.Path,
		"log.level":     Settings.Log.Level,
	}).Info("logger init")
}
