package audit

import (
	"database/sql"
	"errors"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	// LogEntryURIParamKey is a standard string that we'll use to refer to entry IDs with.
	LogEntryURIParamKey = "entryID"
)

// ListHandler is our list route.
func (s *service) ListHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	logger.Debug("ListHandler invoked")

	// ensure query filter.
	filter := types.ExtractQueryFilter(req)

	// determine user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, si.UserID, si.UserIsAdmin)
	logger = logger.WithValue("user_id", si.UserID)

	var (
		entries *types.AuditLogEntryList
		err     error
	)

	if entries, err = s.auditLog.GetAuditLogEntries(ctx, filter); errors.Is(err, sql.ErrNoRows) {
		// in the event no rows exist return an empty list.
		entries = &types.AuditLogEntryList{
			Entries: []types.AuditLogEntry{},
		}
	} else if err != nil {
		logger.Error(err, "error encountered fetching entries")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(res)
		return
	}

	// encode our response and peace.
	s.encoderDecoder.EncodeResponse(res, entries)
}

// ReadHandler returns a GET handler that returns an audit log entry.
func (s *service) ReadHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	logger.Debug("ReadHandler invoked")

	// determine user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(res, "unauthenticated", http.StatusUnauthorized)
		return
	}

	tracing.AttachSessionInfoToSpan(span, si.UserID, si.UserIsAdmin)
	logger = logger.WithValue("user_id", si.UserID)

	// determine audit log entry ID.
	entryID := s.auditLogEntryIDFetcher(req)
	tracing.AttachAuditLogEntryIDToSpan(span, entryID)
	logger = logger.WithValue("audit_log_entry_id", entryID)

	// fetch audit log entry from database.
	x, err := s.auditLog.GetAuditLogEntry(ctx, entryID)
	if errors.Is(err, sql.ErrNoRows) {
		s.encoderDecoder.EncodeNotFoundResponse(res)
		return
	} else if err != nil {
		logger.Error(err, "error fetching audit log entry from database")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(res)
		return
	}

	// encode our response and peace.
	s.encoderDecoder.EncodeResponse(res, x)
}
