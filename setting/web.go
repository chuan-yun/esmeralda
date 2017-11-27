package setting

import (
	"net/url"
	"strings"

	"chuanyun.io/esmeralda/util"
	"github.com/sirupsen/logrus"
)

type Scheme string

const (
	HTTP              Scheme = "http"
	DEFAULT_HTTP_ADDR string = "0.0.0.0"
)

type WebSettings struct {
	Port    int64
	Address string
	Prefix  string
	Schema  Scheme
}

func ValidateWebSettings() {

	u, err := url.Parse("http://" + DEFAULT_HTTP_ADDR + "/" + Settings.Web.Prefix)
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
