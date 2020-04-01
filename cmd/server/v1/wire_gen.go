// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package main

import (
	"context"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/newsman"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/metrics"
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

func BuildServer(ctx context.Context, cfg *config.ServerConfig, logger logging.Logger, database2 database.Database) (*server.Server, error) {
	bcryptHashCost := auth.ProvideBcryptHashCost()
	authenticator := auth.ProvideBcryptAuthenticator(bcryptHashCost, logger)
	userDataManager := users.ProvideUserDataManager(database2)
	clientIDFetcher := httpserver.ProvideOAuth2ServiceClientIDFetcher(logger)
	encoderDecoder := encoding.ProvideResponseEncoder()
	unitCounterProvider := metrics.ProvideUnitCounterProvider()
	service, err := oauth2clients.ProvideOAuth2ClientsService(ctx, logger, database2, authenticator, clientIDFetcher, encoderDecoder, unitCounterProvider)
	if err != nil {
		return nil, err
	}
	oAuth2ClientValidator := auth2.ProvideOAuth2ClientValidator(service)
	userIDFetcher := httpserver.ProvideAuthUserIDFetcher()
	authService := auth2.ProvideAuthService(logger, cfg, authenticator, userDataManager, oAuth2ClientValidator, userIDFetcher, encoderDecoder)
	frontendSettings := config.ProvideConfigFrontendSettings(cfg)
	frontendService := frontend.ProvideFrontendService(logger, frontendSettings)
	itemDataManager := items.ProvideItemDataManager(database2)
	itemsUserIDFetcher := httpserver.ProvideItemServiceUserIDFetcher()
	itemIDFetcher := httpserver.ProvideItemIDFetcher(logger)
	websocketAuthFunc := auth2.ProvideWebsocketAuthFunc(authService)
	typeNameManipulationFunc := httpserver.ProvideNewsmanTypeNameManipulationFunc()
	newsmanNewsman := newsman.NewNewsman(websocketAuthFunc, typeNameManipulationFunc)
	reporter := ProvideReporter(newsmanNewsman)
	itemsService, err := items.ProvideItemsService(ctx, logger, itemDataManager, itemsUserIDFetcher, itemIDFetcher, encoderDecoder, unitCounterProvider, reporter)
	if err != nil {
		return nil, err
	}
	itemDataServer := items.ProvideItemDataServer(itemsService)
	authSettings := config.ProvideConfigAuthSettings(cfg)
	usersUserIDFetcher := httpserver.ProvideUsernameFetcher(logger)
	usersService, err := users.ProvideUsersService(ctx, authSettings, logger, database2, authenticator, usersUserIDFetcher, encoderDecoder, unitCounterProvider, reporter)
	if err != nil {
		return nil, err
	}
	userDataServer := users.ProvideUserDataServer(usersService)
	oAuth2ClientDataServer := oauth2clients.ProvideOAuth2ClientDataServer(service)
	webhookDataManager := webhooks.ProvideWebhookDataManager(database2)
	webhooksUserIDFetcher := httpserver.ProvideWebhooksUserIDFetcher()
	webhookIDFetcher := httpserver.ProvideWebhookIDFetcher(logger)
	webhooksService, err := webhooks.ProvideWebhooksService(ctx, logger, webhookDataManager, webhooksUserIDFetcher, webhookIDFetcher, encoderDecoder, unitCounterProvider, newsmanNewsman)
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

// ProvideReporter is an obligatory function that hopefully wire will eliminate for me one day
func ProvideReporter(n *newsman.Newsman) newsman.Reporter {
	return n
}
