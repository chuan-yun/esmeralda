package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {

	logrus.Info(viper.Get("elasticsearch"))
}
