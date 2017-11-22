package setting

const (
	PROD    string = "production"
	STAGING string = "staging"
	TEST    string = "testing"
	DEV     string = "development"
)

var Application struct {
	Env string
}
