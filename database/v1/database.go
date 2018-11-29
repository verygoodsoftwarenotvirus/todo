package database

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/sirupsen/logrus"
)

type Config struct {
	Debug            bool
	ConnectionString string
	Logger           *logrus.Logger
	SchemaDir        string
}

type Database interface {
	Migrate(schemaDir string) error

	models.ItemHandler
}
