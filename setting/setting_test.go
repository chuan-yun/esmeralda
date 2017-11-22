package setting

import "testing"
import "github.com/sirupsen/logrus"

func TestReadConfigFile(t *testing.T) {
	configFilePath := "../esmeralda.toml"
	ReadConfigFile(configFilePath)
	// logrus.Info("tst ", Settings)
	LogInitialize()
	logrus.Info("tst2 ", Settings)
}
