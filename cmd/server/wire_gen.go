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
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/plans"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/users"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/webhooks"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	config2 "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/password"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/search/bleve"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads/images"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads/storage"
)

// Injectors from wire.go:

// BuildServer builds a server.
func BuildServer(ctx context.Context, cfg *config.ServerConfig, logger logging.Logger, dbm database.DataManager, db *sql.DB, authenticator password.Authenticator) (*server.Server, error) {
	httpserverConfig := cfg.Server
	frontendConfig := cfg.Frontend
	observabilityConfig := &cfg.Observability
	metricsConfig := &observabilityConfig.Metrics
	instrumentationHandler := metrics.ProvideMetricsInstrumentationHandlerForServer(metricsConfig, logger)
	authConfig := &cfg.Auth
	userDataManager := database.ProvideUserDataManager(dbm)
	authAuditManager := database.ProvideAuthAuditManager(dbm)
	oAuth2ClientDataManager := database.ProvideOAuth2ClientDataManager(dbm)
	oAuth2ClientAuditManager := database.ProvideOAuth2ClientAuditManager(dbm)
	encoderDecoder := encoding.ProvideEncoderDecoder(logger)
	unitCounterProvider := metrics.ProvideUnitCounterProvider()
	oAuth2ClientDataService, err := oauth2clients.ProvideOAuth2ClientsService(logger, oAuth2ClientDataManager, userDataManager, oAuth2ClientAuditManager, authenticator, encoderDecoder, unitCounterProvider)
	if err != nil {
		return nil, err
	}
	cookieConfig := authConfig.Cookies
	configConfig := cfg.Database
	sessionManager, err := config2.ProvideSessionManager(cookieConfig, configConfig, db)
	if err != nil {
		return nil, err
	}
	authService, err := auth.ProvideService(logger, authConfig, authenticator, userDataManager, authAuditManager, oAuth2ClientDataService, sessionManager, encoderDecoder)
	if err != nil {
		return nil, err
	}
	frontendService := frontend.ProvideService(logger, frontendConfig)
	auditLogEntryDataManager := database.ProvideAuditLogEntryDataManager(dbm)
	auditLogEntryDataService := audit.ProvideService(logger, auditLogEntryDataManager, encoderDecoder)
	itemDataManager := database.ProvideItemDataManager(dbm)
	itemAuditManager := database.ProvideItemAuditManager(dbm)
	searchConfig := cfg.Search
	indexManagerProvider := bleve.ProvideBleveIndexManagerProvider()
	itemDataService, err := items.ProvideService(logger, itemDataManager, itemAuditManager, encoderDecoder, unitCounterProvider, searchConfig, indexManagerProvider)
	if err != nil {
		return nil, err
	}
	accountDataManager := database.ProvideAccountDataManager(dbm)
	userAuditManager := database.ProvideUserAuditManager(dbm)
	imageUploadProcessor := images.NewImageUploadProcessor()
	uploadsConfig := &cfg.Uploads
	storageConfig := &uploadsConfig.Storage
	uploader, err := storage.NewUploadManager(ctx, logger, storageConfig)
	if err != nil {
		return nil, err
	}
	uploadManager := uploads.ProvideUploadManager(uploader)
	userDataService, err := users.ProvideUsersService(authConfig, logger, userDataManager, accountDataManager, userAuditManager, authenticator, encoderDecoder, unitCounterProvider, imageUploadProcessor, uploadManager)
	if err != nil {
		return nil, err
	}
	accountSubscriptionPlanDataManager := database.ProvidePlanDataManager(dbm)
	accountSubscriptionPlanAuditManager := database.ProvidePlanAuditManager(dbm)
	accountSubscriptionPlanDataService, err := plans.ProvideService(logger, accountSubscriptionPlanDataManager, accountSubscriptionPlanAuditManager, encoderDecoder, unitCounterProvider)
	if err != nil {
		return nil, err
	}
	webhookDataManager := database.ProvideWebhookDataManager(dbm)
	webhookAuditManager := database.ProvideWebhookAuditManager(dbm)
	webhookDataService, err := webhooks.ProvideWebhooksService(logger, webhookDataManager, webhookAuditManager, encoderDecoder, unitCounterProvider)
	if err != nil {
		return nil, err
	}
	adminUserDataManager := database.ProvideAdminUserDataManager(dbm)
	adminAuditManager := database.ProvideAdminAuditManager(dbm)
	adminService, err := admin.ProvideService(logger, authConfig, authenticator, adminUserDataManager, adminAuditManager, sessionManager, encoderDecoder)
	if err != nil {
		return nil, err
	}
	httpserverServer, err := httpserver.ProvideServer(httpserverConfig, frontendConfig, instrumentationHandler, authService, frontendService, auditLogEntryDataService, itemDataService, userDataService, accountSubscriptionPlanDataService, oAuth2ClientDataService, webhookDataService, adminService, dbm, logger, encoderDecoder)
	if err != nil {
		return nil, err
	}
	serverServer, err := server.ProvideServer(cfg, httpserverServer)
	if err != nil {
		return nil, err
	}
	return serverServer, nil
}
