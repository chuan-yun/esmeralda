package setting

import (
	"net/url"
	"strings"

	"chuanyun.io/esmeralda/util"
	"github.com/sirupsen/logrus"
)

type WebSettings struct {
	Port   int64
	Prefix string
	Schema string
}

func ValidateWebSettings() {

	u, err := url.Parse("http://localhost/" + Settings.Web.Prefix)
	if err != nil {
		logrus.Fatal(util.Message("Web prefix path error"))
	}

	Settings.Web.Prefix = u.Path

	if strings.HasSuffix(Settings.Web.Prefix, "/") {
		Settings.Web.Prefix = strings.TrimSuffix(Settings.Web.Prefix, "/")
	}
	if !strings.HasPrefix(Settings.Web.Prefix, "/") {
		Settings.Web.Prefix = "/" + Settings.Web.Prefix
	}

	if Settings.Web.Prefix == "/" {
		Settings.Web.Prefix = ""
	}
}
