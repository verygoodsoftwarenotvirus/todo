package database

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/sirupsen/logrus"
)

type Config struct {
	Debug            bool
	ConnectionString string
	Logger           *logrus.Logger
	SchemaDir        string

	Extractor       ClientIDExtractor
	UserIDExtractor UserIDExtractor
	SecretGenerator SecretGenerator
}

type Database interface {
	Migrate(schemaDir string) error

	models.ItemHandler
	models.UserHandler
	models.OauthClientHandler
}

type SecretGenerator interface {
	GenerateSecret(length uint) string
}

type ClientIDExtractor interface {
	ExtractClientID(req *http.Request) string
}

type UserIDExtractor interface {
	ExtractUserID(req *http.Request) (string, error)
}

type Scannable interface {
	Scan(dest ...interface{}) error
}
