package setting

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Log struct {
	Level string
	Path  string
}

func LogSettingInitialize() {

	v.GetString("log.path")

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
