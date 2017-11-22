package setting

import (
	"fmt"
	"os"

	"chuanyun.io/esmeralda/util"
	"github.com/sirupsen/logrus"
)

type LogSettings struct {
	Level string
	Path  string
}

func LogInitialize() {

	logrus.SetFormatter(&logrus.JSONFormatter{})

	logFile, err := os.OpenFile(Settings.Log.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		panic(util.Message(err.Error()))
	}
	logrus.SetOutput(logFile)

	level, err := logrus.ParseLevel(Settings.Log.Level)
	if err != nil {
		level = logrus.InfoLevel
		Settings.Log.Level = fmt.Sprintf("%s", logrus.InfoLevel)
	}
	logrus.SetLevel(level)

	logrus.WithFields(logrus.Fields{
		"log.formatter": "JSONFormatter",
		"log.path":      Settings.Log.Path,
		"log.level":     Settings.Log.Level,
	}).Info("logger init")
}
