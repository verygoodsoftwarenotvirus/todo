// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package main

import (
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/newsman"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/auth/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/metrics/v1"
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

func BuildServer(cfg *config.ServerConfig, logger logging.Logger, database2 database.Database) (*server.Server, error) {
	bcryptHashCost := auth.ProvideBcryptHashCost()
	authenticator := auth.ProvideBcrypt(bcryptHashCost, logger)
	userDataManager := users.ProvideUserDataManager(database2)
	clientIDFetcher := httpserver.ProvideOAuth2ServiceClientIDFetcher(logger)
	encoderDecoder := encoding.ProvideResponseEncoder()
	unitCounterProvider := metrics.ProvideUnitCounterProvider()
	service, err := oauth2clients.ProvideOAuth2ClientsService(logger, database2, authenticator, clientIDFetcher, encoderDecoder, unitCounterProvider)
	if err != nil {
		return nil, err
	}
	oAuth2ClientValidator := auth2.ProvideOAuth2ClientValidator(service)
	userIDFetcher := httpserver.ProvideAuthUserIDFetcher()
	authService := auth2.ProvideAuthService(logger, cfg, authenticator, userDataManager, oAuth2ClientValidator, userIDFetcher, encoderDecoder)
	frontendService := frontend.ProvideFrontendService(logger)
	itemDataManager := items.ProvideItemDataManager(database2)
	itemsUserIDFetcher := httpserver.ProvideUserIDFetcher()
	itemIDFetcher := httpserver.ProvideItemIDFetcher(logger)
	websocketAuthFunc := auth2.ProvideWebsocketAuthFunc(authService)
	typeNameManipulationFunc := httpserver.ProvideNewsmanTypeNameManipulationFunc(logger)
	newsmanNewsman := newsman.NewNewsman(websocketAuthFunc, typeNameManipulationFunc)
	itemsService, err := items.ProvideItemsService(logger, itemDataManager, itemsUserIDFetcher, itemIDFetcher, encoderDecoder, unitCounterProvider, newsmanNewsman)
	if err != nil {
		return nil, err
	}
	authSettings := config.ProvideConfigAuthSettings(cfg)
	usersUserIDFetcher := httpserver.ProvideUsernameFetcher(logger)
	usersService, err := users.ProvideUsersService(authSettings, logger, database2, authenticator, usersUserIDFetcher, encoderDecoder, unitCounterProvider, newsmanNewsman)
	if err != nil {
		return nil, err
	}
	webhookDataManager := webhooks.ProvideWebhookDataManager(database2)
	webhooksUserIDFetcher := httpserver.ProvideWebhooksUserIDFetcher()
	webhookIDFetcher := httpserver.ProvideWebhookIDFetcher(logger)
	webhooksService, err := webhooks.ProvideWebhooksService(logger, webhookDataManager, webhooksUserIDFetcher, webhookIDFetcher, encoderDecoder, unitCounterProvider, newsmanNewsman)
	if err != nil {
		return nil, err
	}
	httpserverServer, err := httpserver.ProvideServer(cfg, authService, frontendService, itemsService, usersService, service, webhooksService, database2, logger, encoderDecoder, newsmanNewsman)
	if err != nil {
		return nil, err
	}
	serverServer, err := server.ProvideServer(database2, logger, cfg, httpserverServer)
	if err != nil {
		return nil, err
	}
	return serverServer, nil
}
