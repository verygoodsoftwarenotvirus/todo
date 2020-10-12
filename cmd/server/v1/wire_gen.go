// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package main

import (
	"context"
	"database/sql"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
	"gitlab.com/verygoodsoftwarenotvirus/newsman"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/search/bleve"
	"gitlab.com/verygoodsoftwarenotvirus/todo/server/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/server/v1/http"
	auth2 "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/frontend"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/items"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/oauth2clients"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/webhooks"
)

// Injectors from wire.go:

// BuildServer builds a server.
func BuildServer(ctx context.Context, cfg *config.ServerConfig, logger logging.Logger, database2 database.DataManager, db *sql.DB) (*server.Server, error) {
	authSettings := config.ProvideConfigAuthSettings(cfg)
	bcryptHashCost := auth.ProvideBcryptHashCost()
	authenticator := auth.ProvideBcryptAuthenticator(bcryptHashCost, logger)
	userDataManager := users.ProvideUserDataManager(database2)
	clientIDFetcher := httpserver.ProvideOAuth2ClientsServiceClientIDFetcher(logger)
	encoderDecoder := encoding.ProvideResponseEncoder(logger)
	unitCounterProvider := metrics.ProvideUnitCounterProvider()
	service, err := oauth2clients.ProvideOAuth2ClientsService(logger, database2, authenticator, clientIDFetcher, encoderDecoder, unitCounterProvider)
	if err != nil {
		return nil, err
	}
	oAuth2ClientValidator := auth2.ProvideOAuth2ClientValidator(service)
	databaseSettings := config.ProvideConfigDatabaseSettings(cfg)
	sessionManager := config.ProvideSessionManager(authSettings, databaseSettings, db)
	authService, err := auth2.ProvideAuthService(logger, authSettings, authenticator, userDataManager, oAuth2ClientValidator, sessionManager, encoderDecoder)
	if err != nil {
		return nil, err
	}
	frontendSettings := config.ProvideConfigFrontendSettings(cfg)
	frontendService := frontend.ProvideFrontendService(logger, frontendSettings)
	itemDataManager := items.ProvideItemDataManager(database2)
	itemIDFetcher := httpserver.ProvideItemsServiceItemIDFetcher(logger)
	sessionInfoFetcher := httpserver.ProvideItemsSessionInfoFetcher()
	websocketAuthFunc := auth2.ProvideWebsocketAuthFunc(authService)
	typeNameManipulationFunc := httpserver.ProvideNewsmanTypeNameManipulationFunc()
	newsmanNewsman := newsman.NewNewsman(websocketAuthFunc, typeNameManipulationFunc)
	reporter := ProvideReporter(newsmanNewsman)
	searchSettings := config.ProvideSearchSettings(cfg)
	indexManagerProvider := bleve.ProvideBleveIndexManagerProvider()
	searchIndex, err := items.ProvideItemsServiceSearchIndex(searchSettings, indexManagerProvider, logger)
	if err != nil {
		return nil, err
	}
	itemsService, err := items.ProvideItemsService(logger, itemDataManager, itemIDFetcher, sessionInfoFetcher, encoderDecoder, unitCounterProvider, reporter, searchIndex)
	if err != nil {
		return nil, err
	}
	itemDataServer := items.ProvideItemDataServer(itemsService)
	userIDFetcher := httpserver.ProvideUsersServiceUserIDFetcher(logger)
	usersService, err := users.ProvideUsersService(authSettings, logger, userDataManager, authenticator, userIDFetcher, encoderDecoder, unitCounterProvider, reporter)
	if err != nil {
		return nil, err
	}
	userDataServer := users.ProvideUserDataServer(usersService)
	oAuth2ClientDataServer := oauth2clients.ProvideOAuth2ClientDataServer(service)
	webhookDataManager := webhooks.ProvideWebhookDataManager(database2)
	webhooksUserIDFetcher := httpserver.ProvideWebhooksServiceUserIDFetcher()
	webhookIDFetcher := httpserver.ProvideWebhooksServiceWebhookIDFetcher(logger)
	webhooksService, err := webhooks.ProvideWebhooksService(logger, webhookDataManager, webhooksUserIDFetcher, webhookIDFetcher, encoderDecoder, unitCounterProvider, newsmanNewsman)
	if err != nil {
		return nil, err
	}
	webhookDataServer := webhooks.ProvideWebhookDataServer(webhooksService)
	httpserverServer, err := httpserver.ProvideServer(ctx, cfg, authService, frontendService, itemDataServer, userDataServer, oAuth2ClientDataServer, webhookDataServer, database2, logger, encoderDecoder, newsmanNewsman)
	if err != nil {
		return nil, err
	}
	serverServer, err := server.ProvideServer(cfg, httpserverServer)
	if err != nil {
		return nil, err
	}
	return serverServer, nil
}

// wire.go:

// ProvideReporter is an obligatory function that hopefully wire will eliminate for me one day.
func ProvideReporter(n *newsman.Newsman) newsman.Reporter {
	return n
}
