package audit

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

const (
	serviceName = "audit_log_entries_service"
)

var (
	_ types.AuditLogDataService = (*Service)(nil)
)

type (
	// Service handles audit log entries.
	Service struct {
		logger                 logging.Logger
		auditLog               types.AuditLogDataManager
		auditLogEntryIDFetcher EntryIDFetcher
		sessionInfoFetcher     SessionInfoFetcher
		encoderDecoder         encoding.EncoderDecoder
	}

	// SessionInfoFetcher is a function that fetches user IDs.
	SessionInfoFetcher func(*http.Request) (*types.SessionInfo, error)

	// EntryIDFetcher is a function that fetches entry IDs.
	EntryIDFetcher func(*http.Request) uint64
)

// ProvideService builds a new Service.
func ProvideService(
	logger logging.Logger,
	auditLog types.AuditLogDataManager,
	auditLogEntryIDFetcher EntryIDFetcher,
	sessionInfoFetcher SessionInfoFetcher,
	encoder encoding.EncoderDecoder,
) *Service {
	svc := &Service{
		logger:                 logger.WithName(serviceName),
		auditLog:               auditLog,
		auditLogEntryIDFetcher: auditLogEntryIDFetcher,
		sessionInfoFetcher:     sessionInfoFetcher,
		encoderDecoder:         encoder,
	}

	return svc
}
