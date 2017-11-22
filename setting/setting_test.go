package setting

import "testing"
import "github.com/spf13/viper"

func TestReadConfigFile(t *testing.T) {
	configFilePath := "../esmeralda.toml"
	ReadConfigFile(configFilePath)
	t.Log(viper.AllSettings())
}
