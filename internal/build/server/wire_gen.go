// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package server

import (
	"context"
	"database/sql"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/server"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/server/http"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/accounts"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/accountsubscriptionplans"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/admin"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/apiclients"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/frontend"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/frontend2"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/items"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/users"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/services/webhooks"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	config2 "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/passwords"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing/chi"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/search/bleve"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads/images"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/uploads/storage"
)

// Injectors from build.go:

// Build builds a server.
func Build(ctx context.Context, cfg *config.ServerConfig, logger logging.Logger, dbm database.DataManager, db *sql.DB, authenticator passwords.Authenticator) (*server.Server, error) {
	httpserverConfig := cfg.Server
	observabilityConfig := &cfg.Observability
	metricsConfig := observabilityConfig.Metrics
	config3 := &observabilityConfig.Metrics
	instrumentationHandler, err := metrics.ProvideMetricsInstrumentationHandlerForServer(config3, logger)
	if err != nil {
		return nil, err
	}
	authConfig := &cfg.Auth
	userDataManager := database.ProvideUserDataManager(dbm)
	authAuditManager := database.ProvideAuthAuditManager(dbm)
	apiClientDataManager := database.ProvideAPIClientDataManager(dbm)
	accountUserMembershipDataManager := database.ProvideAccountUserMembershipDataManager(dbm)
	cookieConfig := authConfig.Cookies
	configConfig := cfg.Database
	sessionManager, err := config2.ProvideSessionManager(cookieConfig, configConfig, db)
	if err != nil {
		return nil, err
	}
	encodingConfig := cfg.Encoding
	contentType := encoding.ProvideContentType(encodingConfig)
	serverEncoderDecoder := encoding.ProvideServerEncoderDecoder(logger, contentType)
	routeParamManager := chi.NewRouteParamManager()
	authService, err := auth.ProvideService(logger, authConfig, authenticator, userDataManager, authAuditManager, apiClientDataManager, accountUserMembershipDataManager, sessionManager, serverEncoderDecoder, routeParamManager)
	if err != nil {
		return nil, err
	}
	auditLogEntryDataManager := database.ProvideAuditLogEntryDataManager(dbm)
	auditLogEntryDataService := audit.ProvideService(logger, auditLogEntryDataManager, serverEncoderDecoder, routeParamManager)
	accountDataManager := database.ProvideAccountDataManager(dbm)
	unitCounterProvider, err := metrics.ProvideUnitCounterProvider(config3, logger)
	if err != nil {
		return nil, err
	}
	imageUploadProcessor := images.NewImageUploadProcessor(logger)
	uploadsConfig := &cfg.Uploads
	storageConfig := &uploadsConfig.Storage
	uploader, err := storage.NewUploadManager(ctx, logger, storageConfig, routeParamManager)
	if err != nil {
		return nil, err
	}
	uploadManager := uploads.ProvideUploadManager(uploader)
	userDataService := users.ProvideUsersService(authConfig, logger, userDataManager, accountDataManager, authenticator, serverEncoderDecoder, unitCounterProvider, imageUploadProcessor, uploadManager, routeParamManager)
	accountDataService := accounts.ProvideService(logger, accountDataManager, accountUserMembershipDataManager, serverEncoderDecoder, unitCounterProvider, routeParamManager)
	accountSubscriptionPlanDataManager := database.ProvidePlanDataManager(dbm)
	accountSubscriptionPlanDataService := accountsubscriptionplans.ProvideService(logger, accountSubscriptionPlanDataManager, serverEncoderDecoder, unitCounterProvider, routeParamManager)
	apiClientDataService := apiclients.ProvideAPIClientsService(logger, apiClientDataManager, userDataManager, authenticator, serverEncoderDecoder, unitCounterProvider, routeParamManager)
	itemDataManager := database.ProvideItemDataManager(dbm)
	searchConfig := cfg.Search
	indexManagerProvider := bleve.ProvideBleveIndexManagerProvider()
	itemDataService, err := items.ProvideService(logger, itemDataManager, serverEncoderDecoder, unitCounterProvider, searchConfig, indexManagerProvider, routeParamManager)
	if err != nil {
		return nil, err
	}
	webhookDataManager := database.ProvideWebhookDataManager(dbm)
	webhookDataService := webhooks.ProvideWebhooksService(logger, webhookDataManager, serverEncoderDecoder, unitCounterProvider, routeParamManager)
	adminUserDataManager := database.ProvideAdminUserDataManager(dbm)
	adminAuditManager := database.ProvideAdminAuditManager(dbm)
	adminService := admin.ProvideService(logger, authConfig, authenticator, adminUserDataManager, adminAuditManager, sessionManager, serverEncoderDecoder, routeParamManager)
	frontendConfig := cfg.Frontend
	frontendService := frontend.ProvideService(logger, frontendConfig)
	service := frontend2.ProvideService(logger)
	router := chi.NewRouter(logger)
	httpserverServer, err := httpserver.ProvideServer(ctx, httpserverConfig, metricsConfig, instrumentationHandler, authService, auditLogEntryDataService, userDataService, accountDataService, accountSubscriptionPlanDataService, apiClientDataService, itemDataService, webhookDataService, adminService, frontendService, service, dbm, logger, serverEncoderDecoder, router)
	if err != nil {
		return nil, err
	}
	serverServer, err := server.ProvideServer(cfg, httpserverServer)
	if err != nil {
		return nil, err
	}
	return serverServer, nil
}
