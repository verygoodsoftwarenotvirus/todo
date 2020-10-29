package audit

import (
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/metrics"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

const (
	counterName        metrics.CounterName = "audit_log_entries"
	counterDescription string              = "the number of audit log entries managed by the audit service"
	serviceName        string              = "audit_log_entries_service"
)

var (
	_ models.AuditLogEntryDataServer = (*Service)(nil)
)

type (
	// Service handles to-do list items
	Service struct {
		logger                 logging.Logger
		auditLog               models.AuditLogEntryDataManager
		auditLogEntryIDFetcher EntryIDFetcher
		sessionInfoFetcher     SessionInfoFetcher
		itemCounter            metrics.UnitCounter
		encoderDecoder         encoding.EncoderDecoder
	}

	// SessionInfoFetcher is a function that fetches user IDs.
	SessionInfoFetcher func(*http.Request) (*models.SessionInfo, error)

	// EntryIDFetcher is a function that fetches item IDs.
	EntryIDFetcher func(*http.Request) uint64
)

// ProvideAuditService builds a new ItemsService.
func ProvideAuditService(
	logger logging.Logger,
	auditLog models.AuditLogEntryDataManager,
	auditLogEntryIDFetcher EntryIDFetcher,
	sessionInfoFetcher SessionInfoFetcher,
	counterProvider metrics.UnitCounterProvider,
	encoder encoding.EncoderDecoder,
) (*Service, error) {
	entryCounter, err := counterProvider(counterName, counterDescription)
	if err != nil {
		return nil, fmt.Errorf("error initializing counter: %w", err)
	}

	svc := &Service{
		logger:                 logger.WithName(serviceName),
		auditLog:               auditLog,
		auditLogEntryIDFetcher: auditLogEntryIDFetcher,
		sessionInfoFetcher:     sessionInfoFetcher,
		itemCounter:            entryCounter,
		encoderDecoder:         encoder,
	}

	return svc, nil
}
