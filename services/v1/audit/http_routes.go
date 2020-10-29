package audit

import (
	"database/sql"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/tracing"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

const (
	// URIParamKey is a standard string that we'll use to refer to entry IDs with.
	URIParamKey = "entryID"
)

// ListHandler is our list route.
func (s *Service) ListHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := tracing.StartSpan(req.Context(), "ListHandler")
	defer span.End()

	logger := s.logger.WithRequest(req)

	// ensure query filter.
	filter := models.ExtractQueryFilter(req)

	// determine user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(res, "unauthenticated", http.StatusUnauthorized)
		return
	}
	tracing.AttachSessionInfoToSpan(span, *si)
	logger = logger.WithValue("user_id", si.UserID)

	var (
		entries *models.AuditLogEntryList
		err     error
	)

	if entries, err = s.auditLog.GetAuditLogEntries(ctx, filter); err == sql.ErrNoRows {
		// in the event no rows exist return an empty list.
		entries = &models.AuditLogEntryList{
			Entries: []models.AuditLogEntry{},
		}
	} else if err != nil {
		logger.Error(err, "error encountered fetching entries")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(res)
		return
	}

	// encode our response and peace.
	s.encoderDecoder.EncodeResponse(res, entries)
}

// ReadHandler returns a GET handler that returns an item.
func (s *Service) ReadHandler(res http.ResponseWriter, req *http.Request) {
	ctx, span := tracing.StartSpan(req.Context(), "ReadHandler")
	defer span.End()

	logger := s.logger.WithRequest(req)

	// determine user ID.
	si, sessionInfoRetrievalErr := s.sessionInfoFetcher(req)
	if sessionInfoRetrievalErr != nil {
		s.encoderDecoder.EncodeErrorResponse(res, "unauthenticated", http.StatusUnauthorized)
		return
	}
	tracing.AttachSessionInfoToSpan(span, *si)
	logger = logger.WithValue("user_id", si.UserID)

	// determine item ID.
	entryID := s.auditLogEntryIDFetcher(req)
	tracing.AttachAuditLogEntryIDToSpan(span, entryID)
	logger = logger.WithValue("audit_log_entry_id", entryID)

	// fetch item from database.
	x, err := s.auditLog.GetAuditLogEntry(ctx, entryID)
	if err == sql.ErrNoRows {
		s.encoderDecoder.EncodeNotFoundResponse(res)
		return
	} else if err != nil {
		logger.Error(err, "error fetching item from database")
		s.encoderDecoder.EncodeUnspecifiedInternalServerErrorResponse(res)
		return
	}

	// encode our response and peace.
	s.encoderDecoder.EncodeResponse(res, x)
}
