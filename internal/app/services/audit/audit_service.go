package audit

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/routeparams"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

const (
	serviceName = "audit_log_entries_service"
)

var (
	_ types.AuditLogDataService = (*service)(nil)
)

type (
	// service handles audit log entries.
	service struct {
		logger                 logging.Logger
		auditLog               types.AuditLogDataManager
		auditLogEntryIDFetcher func(*http.Request) uint64
		sessionInfoFetcher     func(*http.Request) (*types.SessionInfo, error)
		encoderDecoder         encoding.EncoderDecoder
		tracer                 tracing.Tracer
	}
)

// ProvideService builds a new service.
func ProvideService(
	logger logging.Logger,
	auditLog types.AuditLogDataManager,
	encoder encoding.EncoderDecoder,
) types.AuditLogDataService {
	svc := &service{
		logger:                 logger.WithName(serviceName),
		auditLog:               auditLog,
		auditLogEntryIDFetcher: routeparams.BuildRouteParamIDFetcher(logger, LogEntryURIParamKey, "audit log entry"),
		sessionInfoFetcher:     routeparams.SessionInfoFetcherFromRequestContext,
		encoderDecoder:         encoder,
		tracer:                 tracing.NewTracer(serviceName),
	}

	return svc
}
