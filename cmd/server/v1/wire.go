//+build wireinject

package main

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/sqlite"
	"gitlab.com/verygoodsoftwarenotvirus/todo/server/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/items"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"

	"github.com/google/wire"
	"github.com/sirupsen/logrus"
)

// BuildServer builds a server
func BuildServer(
	connectionDetails database.ConnectionDetails,
	SchemaDirectory database.SchemaDirectory,
	CertPair server.CertPair,
	CookieName users.CookieName,
	CookieSecret []byte,
	Debug bool,
) (*server.Server, error) {

	wire.Build(
		server.ProvideUserIDFetcher,
		server.ProvideUsernameFetcher,
		auth.NewBcrypt,
		logrus.New,
		provideJaeger,
		sqlite.ProvideSqlite,
		users.ProvideUsersService,
		items.ProvideItemsService,
		server.ProvideServer,
	)
	return nil, nil
}
