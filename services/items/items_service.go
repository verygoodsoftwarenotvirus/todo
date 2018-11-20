package items

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/database"

	"github.com/sirupsen/logrus"
)

type ItemsService struct {
	logger *logrus.Logger
	db     database.Database
}

type ItemsServiceConfig struct {
	Logger *logrus.Logger
	DB     database.Database
}

func NewItemsService(cfg ItemsServiceConfig) *ItemsService {
	return &ItemsService{
		logger: cfg.Logger,
		db:     cfg.DB,
	}
}
