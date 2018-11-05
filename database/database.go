package database

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/models"
)

type Database interface {
	models.ItemHandler
}
