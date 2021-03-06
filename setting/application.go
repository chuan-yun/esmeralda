package setting

type Environment string

const (
	PROD    Environment = "production"
	STAGING Environment = "staging"
	TEST    Environment = "testing"
	DEV     Environment = "development"
)

type ApplicationSettings struct {
	Env   Environment
	Debug bool
}
