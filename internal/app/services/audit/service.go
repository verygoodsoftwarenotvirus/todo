package audit

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
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
		logger                 logging.Logger
		auditLog               types.AuditLogEntryDataManager
		auditLogEntryIDFetcher func(*http.Request) uint64
		sessionInfoFetcher     func(*http.Request) (*types.SessionInfo, error)
		encoderDecoder         encoding.HTTPResponseEncoder
		tracer                 tracing.Tracer
	}
)

// ProvideService builds a new service.
func ProvideService(
	logger logging.Logger,
	auditLog types.AuditLogEntryDataManager,
	encoder encoding.HTTPResponseEncoder,
	routeParamManager routing.RouteParamManager,
) types.AuditLogEntryDataService {
	svc := &service{
		logger:                 logging.EnsureLogger(logger).WithName(serviceName),
		auditLog:               auditLog,
		auditLogEntryIDFetcher: routeParamManager.BuildRouteParamIDFetcher(logger, LogEntryURIParamKey, "audit log entry"),
		sessionInfoFetcher:     routeParamManager.SessionInfoFetcherFromRequestContext,
		encoderDecoder:         encoder,
		tracer:                 tracing.NewTracer(serviceName),
	}

	return svc
}
