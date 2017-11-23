package setting

const (
	PROD    string = "production"
	STAGING string = "staging"
	TEST    string = "testing"
	DEV     string = "development"
)

type ApplicationSettings struct {
	Env      string
	Username string
	Password string
	Bulk     int64
}
