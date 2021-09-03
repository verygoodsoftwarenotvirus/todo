// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package server

import (
	"context"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authentication"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/capitalism/stripe"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	config2 "gitlab.com/verygoodsoftwarenotvirus/todo/internal/database/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/events"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/metrics"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing/chi"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/search/bleve"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/server"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/accounts"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/admin"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/apiclients"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/audit"
	authentication2 "gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/authentication"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/frontend"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/items"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/users"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/webhooks"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/storage"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/uploads"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/uploads/images"
)

// Injectors from build.go:

// Build builds a server.
func Build(ctx context.Context, cfg *config.InstanceConfig, logger logging.Logger) (*server.HTTPServer, error) {
	serverConfig := cfg.Server
	observabilityConfig := &cfg.Observability
	metricsConfig := &observabilityConfig.Metrics
	instrumentationHandler, err := metrics.ProvideMetricsInstrumentationHandlerForServer(metricsConfig, logger)
	if err != nil {
		return nil, err
	}
	servicesConfigurations := &cfg.Services
	authenticationConfig := &servicesConfigurations.Auth
	authenticator := authentication.ProvideArgon2Authenticator(logger)
	configConfig := &cfg.Database
	db, err := config2.ProvideDatabaseConnection(logger, configConfig)
	if err != nil {
		return nil, err
	}
	dataManager, err := config.ProvideDatabaseClient(ctx, logger, db, cfg)
	if err != nil {
		return nil, err
	}
	userDataManager := database.ProvideUserDataManager(dataManager)
	authAuditManager := database.ProvideAuthAuditManager(dataManager)
	apiClientDataManager := database.ProvideAPIClientDataManager(dataManager)
	accountUserMembershipDataManager := database.ProvideAccountUserMembershipDataManager(dataManager)
	cookieConfig := authenticationConfig.Cookies
	config3 := cfg.Database
	sessionManager, err := config2.ProvideSessionManager(cookieConfig, config3, db)
	if err != nil {
		return nil, err
	}
	encodingConfig := cfg.Encoding
	contentType := encoding.ProvideContentType(encodingConfig)
	serverEncoderDecoder := encoding.ProvideServerEncoderDecoder(logger, contentType)
	routeParamManager := chi.NewRouteParamManager()
	authService, err := authentication2.ProvideService(logger, authenticationConfig, authenticator, userDataManager, authAuditManager, apiClientDataManager, accountUserMembershipDataManager, sessionManager, serverEncoderDecoder, routeParamManager)
	if err != nil {
		return nil, err
	}
	auditLogEntryDataManager := database.ProvideAuditLogEntryDataManager(dataManager)
	auditLogEntryDataService := audit.ProvideService(logger, auditLogEntryDataManager, serverEncoderDecoder, routeParamManager)
	accountDataManager := database.ProvideAccountDataManager(dataManager)
	unitCounterProvider, err := metrics.ProvideUnitCounterProvider(metricsConfig, logger)
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
	userDataService := users.ProvideUsersService(authenticationConfig, logger, userDataManager, accountDataManager, authenticator, serverEncoderDecoder, unitCounterProvider, imageUploadProcessor, uploadManager, routeParamManager)
	accountDataService := accounts.ProvideService(logger, accountDataManager, accountUserMembershipDataManager, serverEncoderDecoder, unitCounterProvider, routeParamManager)
	apiclientsConfig := apiclients.ProvideConfig(authenticationConfig)
	apiClientDataService := apiclients.ProvideAPIClientsService(logger, apiClientDataManager, userDataManager, authenticator, serverEncoderDecoder, unitCounterProvider, routeParamManager, apiclientsConfig)
	itemsConfig := servicesConfigurations.Items
	itemDataManager := database.ProvideItemDataManager(dataManager)
	indexManagerProvider := bleve.ProvideBleveIndexManagerProvider()
	producerConfig := cfg.Events
	eventQueueAddress := producerConfig.Address
	producerProvider := events.NewProducerProvider(logger, eventQueueAddress)
	itemDataService, err := items.ProvideService(logger, itemsConfig, itemDataManager, serverEncoderDecoder, unitCounterProvider, indexManagerProvider, routeParamManager, producerProvider)
	if err != nil {
		return nil, err
	}
	webhookDataManager := database.ProvideWebhookDataManager(dataManager)
	webhookDataService := webhooks.ProvideWebhooksService(logger, webhookDataManager, serverEncoderDecoder, unitCounterProvider, routeParamManager)
	adminUserDataManager := database.ProvideAdminUserDataManager(dataManager)
	adminAuditManager := database.ProvideAdminAuditManager(dataManager)
	adminService := admin.ProvideService(logger, authenticationConfig, authenticator, adminUserDataManager, adminAuditManager, sessionManager, serverEncoderDecoder, routeParamManager)
	frontendConfig := &servicesConfigurations.Frontend
	frontendAuthService := frontend.ProvideAuthService(authService)
	usersService := frontend.ProvideUsersService(userDataService)
	capitalismConfig := &cfg.Capitalism
	stripeConfig := capitalismConfig.Stripe
	paymentManager := stripe.ProvideStripePaymentManager(logger, stripeConfig)
	service := frontend.ProvideService(frontendConfig, logger, frontendAuthService, usersService, dataManager, routeParamManager, paymentManager)
	router := chi.NewRouter(logger)
	httpServer, err := server.ProvideHTTPServer(ctx, serverConfig, instrumentationHandler, authService, auditLogEntryDataService, userDataService, accountDataService, apiClientDataService, itemDataService, webhookDataService, adminService, service, logger, serverEncoderDecoder, router)
	if err != nil {
		return nil, err
	}
	return httpServer, nil
}
