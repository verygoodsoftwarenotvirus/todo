package audit

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/services/authentication"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

const (
	serviceName = "audit_log_entries_service"
)

var (
	_ types.AuditLogEntryDataService = (*service)(nil)
)

type (
	// service handles audit log entries.
	service struct {
		logger                    logging.Logger
		auditLog                  types.AuditLogEntryDataManager
		auditLogEntryIDFetcher    func(*http.Request) uint64
		sessionContextDataFetcher func(*http.Request) (*types.SessionContextData, error)
		encoderDecoder            encoding.ServerEncoderDecoder
		tracer                    tracing.Tracer
	}
)

// ProvideService builds a new service.
func ProvideService(
	logger logging.Logger,
	auditLog types.AuditLogEntryDataManager,
	encoder encoding.ServerEncoderDecoder,
	routeParamManager routing.RouteParamManager,
) types.AuditLogEntryDataService {
	return &service{
		logger:                    logging.EnsureLogger(logger).WithName(serviceName),
		auditLog:                  auditLog,
		auditLogEntryIDFetcher:    routeParamManager.BuildRouteParamIDFetcher(logger, LogEntryURIParamKey, "audit log entry"),
		sessionContextDataFetcher: authentication.FetchContextFromRequest,
		encoderDecoder:            encoder,
		tracer:                    tracing.NewTracer(serviceName),
	}
}
