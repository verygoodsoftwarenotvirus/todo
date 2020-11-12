package audit

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

const (
	serviceName = "audit_log_entries_service"
)

var (
	_ models.AuditLogDataServer = (*Service)(nil)
)

type (
	// Service handles audit log entries.
	Service struct {
		logger                 logging.Logger
		auditLog               models.AuditLogDataManager
		auditLogEntryIDFetcher EntryIDFetcher
		sessionInfoFetcher     SessionInfoFetcher
		encoderDecoder         encoding.EncoderDecoder
	}

	// SessionInfoFetcher is a function that fetches user IDs.
	SessionInfoFetcher func(*http.Request) (*models.SessionInfo, error)

	// EntryIDFetcher is a function that fetches entry IDs.
	EntryIDFetcher func(*http.Request) uint64
)

// ProvideAuditService builds a new Service.
func ProvideAuditService(
	logger logging.Logger,
	auditLog models.AuditLogDataManager,
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
