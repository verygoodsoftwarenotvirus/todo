// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package main

import (
	"context"
	"database/sql"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/server"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/server/http"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/admin"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/frontend"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/items"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/oauth2clients"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/users"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/webhooks"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/password"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/search/bleve"
)

// Injectors from wire.go:

// BuildServer builds a server.
func BuildServer(ctx context.Context, cfg *config.ServerConfig, logger logging.Logger, dbm database.DataManager, db *sql.DB, authenticator password.Authenticator) (*server.Server, error) {
	serverSettings := cfg.Server
	frontendSettings := cfg.Frontend
	instrumentationHandler := frontend.ProvideMetricsInstrumentationHandlerForServer(cfg, logger)
	authSettings := cfg.Auth
	userDataManager := database.ProvideUserDataManager(dbm)
	authAuditManager := database.ProvideAuthAuditManager(dbm)
	oAuth2ClientDataManager := database.ProvideOAuth2ClientDataManager(dbm)
	oAuth2ClientAuditManager := database.ProvideOAuth2ClientAuditManager(dbm)
	clientIDFetcher := oauth2clients.ProvideOAuth2ClientsServiceClientIDFetcher(logger)
	encoderDecoder := encoding.ProvideResponseEncoder(logger)
	unitCounterProvider := metrics.ProvideUnitCounterProvider()
	oAuth2ClientDataService, err := oauth2clients.ProvideOAuth2ClientsService(logger, oAuth2ClientDataManager, userDataManager, oAuth2ClientAuditManager, authenticator, clientIDFetcher, encoderDecoder, unitCounterProvider)
	if err != nil {
		return nil, err
	}
	databaseSettings := cfg.Database
	sessionManager, err := config.ProvideSessionManager(authSettings, databaseSettings, db)
	if err != nil {
		return nil, err
	}
	sessionInfoFetcher := auth.ProvideAuthServiceSessionInfoFetcher()
	authService, err := auth.ProvideService(logger, authSettings, authenticator, userDataManager, authAuditManager, oAuth2ClientDataService, sessionManager, encoderDecoder, sessionInfoFetcher)
	if err != nil {
		return nil, err
	}
	frontendService := frontend.ProvideService(logger, frontendSettings)
	auditLogDataManager := database.ProvideAuditLogEntryDataManager(dbm)
	entryIDFetcher := audit.ProvideAuditServiceItemIDFetcher(logger)
	auditSessionInfoFetcher := audit.ProvideAuditServiceSessionInfoFetcher()
	auditLogDataService := audit.ProvideService(logger, auditLogDataManager, entryIDFetcher, auditSessionInfoFetcher, encoderDecoder)
	itemDataManager := database.ProvideItemDataManager(dbm)
	itemAuditManager := database.ProvideItemAuditManager(dbm)
	itemIDFetcher := items.ProvideItemsServiceItemIDFetcher(logger)
	itemsSessionInfoFetcher := items.ProvideItemsServiceSessionInfoFetcher()
	searchSettings := cfg.Search
	indexManagerProvider := bleve.ProvideBleveIndexManagerProvider()
	searchIndex, err := items.ProvideItemsServiceSearchIndex(searchSettings, indexManagerProvider, logger)
	if err != nil {
		return nil, err
	}
	itemDataService, err := items.ProvideService(logger, itemDataManager, itemAuditManager, itemIDFetcher, itemsSessionInfoFetcher, encoderDecoder, unitCounterProvider, searchIndex)
	if err != nil {
		return nil, err
	}
	userAuditManager := database.ProvideUserAuditManager(dbm)
	userIDFetcher := users.ProvideUsersServiceUserIDFetcher(logger)
	usersSessionInfoFetcher := users.ProvideUsersServiceSessionInfoFetcher()
	userDataService, err := users.ProvideUsersService(authSettings, logger, userDataManager, userAuditManager, authenticator, userIDFetcher, usersSessionInfoFetcher, encoderDecoder, unitCounterProvider)
	if err != nil {
		return nil, err
	}
	webhookDataManager := database.ProvideWebhookDataManager(dbm)
	webhookAuditManager := database.ProvideWebhookAuditManager(dbm)
	webhooksSessionInfoFetcher := webhooks.ProvideWebhooksServiceSessionInfoFetcher()
	webhookIDFetcher := webhooks.ProvideWebhooksServiceWebhookIDFetcher(logger)
	webhookDataService, err := webhooks.ProvideWebhooksService(logger, webhookDataManager, webhookAuditManager, webhooksSessionInfoFetcher, webhookIDFetcher, encoderDecoder, unitCounterProvider)
	if err != nil {
		return nil, err
	}
	adminUserDataManager := database.ProvideAdminUserDataManager(dbm)
	adminAuditManager := database.ProvideAdminAuditManager(dbm)
	adminSessionInfoFetcher := admin.ProvideAdminServiceSessionInfoFetcher()
	adminUserIDFetcher := admin.ProvideAdminServiceUserIDFetcher(logger)
	service, err := admin.ProvideService(logger, authSettings, authenticator, adminUserDataManager, adminAuditManager, sessionManager, encoderDecoder, adminSessionInfoFetcher, adminUserIDFetcher)
	if err != nil {
		return nil, err
	}
	adminService := admin.ProvideAdminService(service)
	httpserverServer, err := httpserver.ProvideServer(serverSettings, frontendSettings, instrumentationHandler, authService, frontendService, auditLogDataService, itemDataService, userDataService, oAuth2ClientDataService, webhookDataService, adminService, dbm, logger, encoderDecoder)
	if err != nil {
		return nil, err
	}
	serverServer, err := server.ProvideServer(cfg, httpserverServer)
	if err != nil {
		return nil, err
	}
	return serverServer, nil
}
