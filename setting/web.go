package setting

type WebSettings struct {
	Port   int64
	Prefix string
}

func ValidateWebSettings() {
	if Settings.Web.Prefix == "/" {
		Settings.Web.Prefix = ""
	}
}
