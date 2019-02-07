// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package main

import (
	"github.com/sirupsen/logrus"
	"gitlab.com/verygoodsoftwarenotvirus/todo/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/sqlite"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1/zerolog"
	"gitlab.com/verygoodsoftwarenotvirus/todo/server/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/items"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/oauth2clients"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"
	"gopkg.in/oauth2.v3/manage"
)

// Injectors from wire.go:

func BuildServer(connectionDetails database.ConnectionDetails, SchemaDirectory database.SchemaDirectory, CertPair server.CertPair, CookieName users.CookieName, CookieSecret []byte, Debug bool) (*server.Server, error) {
	logger := logrus.New()
	enticator := auth.NewBcrypt(logger)
	zerologLogger := zerolog.ProvideZerologger()
	loggingLogger := zerolog.ProvideLogger(zerologLogger)
	tracer, err := sqlite.ProvideSqliteTracer()
	if err != nil {
		return nil, err
	}
	databaseDatabase, err := sqlite.ProvideSqlite(Debug, logger, loggingLogger, tracer, connectionDetails)
	if err != nil {
		return nil, err
	}
	userIDFetcher := server.ProvideUserIDFetcher()
	itemIDFetcher := server.ProvideItemIDFetcher()
	serviceTracer, err := items.ProvideItemsServiceTracer()
	if err != nil {
		return nil, err
	}
	service := items.ProvideItemsService(logger, databaseDatabase, userIDFetcher, itemIDFetcher, serviceTracer)
	usernameFetcher := server.ProvideUsernameFetcher()
	usersTracer, err := users.ProvideUserServiceTracer()
	if err != nil {
		return nil, err
	}
	usersService := users.ProvideUsersService(CookieName, logger, databaseDatabase, enticator, usernameFetcher, usersTracer)
	clientStore := server.ProvideClientStore()
	manager := manage.NewDefaultManager()
	tokenStore, err := server.ProvideTokenStore(manager)
	if err != nil {
		return nil, err
	}
	oauth2clientsTracer, err := oauth2clients.ProvideOAuth2ClientsServiceTracer()
	if err != nil {
		return nil, err
	}
	oauth2clientsService := oauth2clients.ProvideOAuth2ClientsService(databaseDatabase, enticator, logger, clientStore, tokenStore, oauth2clientsTracer)
	serverTracer, err := server.ProvideServerTracer()
	if err != nil {
		return nil, err
	}
	serverServer := server.ProvideOAuth2Server(manager, tokenStore, clientStore)
	server2, err := server.ProvideServer(Debug, CertPair, CookieSecret, enticator, SchemaDirectory, service, usersService, oauth2clientsService, databaseDatabase, loggingLogger, serverTracer, serverServer, tokenStore, clientStore)
	if err != nil {
		return nil, err
	}
	return server2, nil
}
