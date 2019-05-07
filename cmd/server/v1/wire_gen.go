// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package main

import (
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/newsman"
	"gitlab.com/verygoodsoftwarenotvirus/todo/config/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/auth/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/metrics/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/server/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/server/v1/http"
	auth2 "gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/auth"
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
	clientIDFetcher := httpserver.ProvideOAuth2ServiceClientIDFetcher()
	encoderDecoder := encoding.ProvideJSONResponseEncoder()
	unitCounterProvider := metrics.ProvideUnitCounterProvider()
	service, err := oauth2clients.ProvideOAuth2ClientsService(logger, database2, authenticator, clientIDFetcher, encoderDecoder, unitCounterProvider)
	if err != nil {
		return nil, err
	}
	userIDFetcher := httpserver.ProvideAuthUserIDFetcher()
	authService := auth2.ProvideAuthService(logger, cfg, authenticator, userDataManager, service, userIDFetcher, encoderDecoder)
	itemsUserIDFetcher := httpserver.ProvideUserIDFetcher()
	itemIDFetcher := httpserver.ProvideItemIDFetcher()
	itemsService, err := items.ProvideItemsService(logger, database2, itemsUserIDFetcher, itemIDFetcher, encoderDecoder, unitCounterProvider)
	if err != nil {
		return nil, err
	}
	authSettings := config.ProvideConfigAuthSettings(cfg)
	usersUserIDFetcher := httpserver.ProvideUsernameFetcher()
	usersService, err := users.ProvideUsersService(authSettings, logger, database2, authenticator, usersUserIDFetcher, encoderDecoder, unitCounterProvider)
	if err != nil {
		return nil, err
	}
	webhooksUserIDFetcher := httpserver.ProvideWebhooksUserIDFetcher()
	webhookIDFetcher := httpserver.ProvideWebhookIDFetcher()
	webhooksService, err := webhooks.ProvideWebhooksService(logger, database2, webhooksUserIDFetcher, webhookIDFetcher, encoderDecoder, unitCounterProvider)
	if err != nil {
		return nil, err
	}
	websocketAuthFunc := auth2.ProvideWebsocketAuthFunc(authService)
	newsmanNewsman := newsman.NewNewsman(websocketAuthFunc)
	httpserverServer, err := httpserver.ProvideServer(cfg, authService, itemsService, usersService, service, webhooksService, database2, logger, encoderDecoder, newsmanNewsman)
	if err != nil {
		return nil, err
	}
	serverServer, err := server.ProvideServer(database2, logger, cfg, httpserverServer)
	if err != nil {
		return nil, err
	}
	return serverServer, nil
}
