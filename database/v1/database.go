package database

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

type (
	// Database describes anything that stores data for our services
	Database interface {
		Migrate(ctx context.Context) error
		IsReady(ctx context.Context) (ready bool)

		AdminUserExists(ctx context.Context) (bool, error)

		models.ItemHandler
		models.UserHandler
		models.OAuth2ClientHandler
	}
)
