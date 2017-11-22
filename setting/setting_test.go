package setting

import "testing"

func TestReadConfigFile(t *testing.T) {
	configFilePath := "../esmeralda.toml"
	ReadConfigFile(configFilePath)
	t.Log("tst ", Settings)
	LogInitialize()
	t.Log("sss ", Settings)
}
