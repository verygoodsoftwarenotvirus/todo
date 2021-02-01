package encoding

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/google/wire"
)

const (
	// ContentTypeHeader is the HTTP standard header name for content type.
	ContentTypeHeader = "Content-type"
	// XMLContentType represents the XML content type.
	XMLContentType = "application/xml"
	// JSONContentType represents the JSON content type.
	JSONContentType = "application/json"
	// DefaultContentType is what the library defaults to.
	DefaultContentType = JSONContentType
)

var (
	// Providers provides ResponseEncoders for dependency injection.
	Providers = wire.NewSet(
		ProvideEncoderDecoder,
	)

	_ EncoderDecoder = (*serverEncoderDecoder)(nil)
)

type (
	// EncoderDecoder is an interface that allows for multiple implementations of HTTP response formats.
	EncoderDecoder interface {
		EncodeResponse(ctx context.Context, res http.ResponseWriter, val interface{})
		EncodeResponseWithStatus(ctx context.Context, res http.ResponseWriter, val interface{}, statusCode int)
		EncodeErrorResponse(ctx context.Context, res http.ResponseWriter, msg string, statusCode int)
		EncodeInvalidInputResponse(ctx context.Context, res http.ResponseWriter)
		EncodeNotFoundResponse(ctx context.Context, res http.ResponseWriter)
		EncodeUnspecifiedInternalServerErrorResponse(ctx context.Context, res http.ResponseWriter)
		EncodeUnauthorizedResponse(ctx context.Context, res http.ResponseWriter)
		EncodeInvalidPermissionsResponse(ctx context.Context, res http.ResponseWriter)
		DecodeRequest(ctx context.Context, req *http.Request, dest interface{}) error
	}

	// serverEncoderDecoder is our concrete implementation of EncoderDecoder.
	serverEncoderDecoder struct {
		logger logging.Logger
		tracer tracing.Tracer
	}

	encoder interface {
		Encode(v interface{}) error
	}

	decoder interface {
		Decode(v interface{}) error
	}
)

// EncodeErrorResponse encodes errors to responses.
func (ed *serverEncoderDecoder) EncodeErrorResponse(ctx context.Context, res http.ResponseWriter, msg string, statusCode int) {
	_, span := ed.tracer.StartSpan(ctx)
	defer span.End()

	var ct = strings.ToLower(res.Header().Get(ContentTypeHeader))
	if ct == "" {
		ct = DefaultContentType
	}

	var e encoder

	switch ct {
	case XMLContentType:
		e = xml.NewEncoder(res)
	default:
		e = json.NewEncoder(res)
	}

	res.Header().Set(ContentTypeHeader, ct)
	res.WriteHeader(statusCode)

	if err := e.Encode(&types.ErrorResponse{Message: msg, Code: statusCode}); err != nil {
		ed.logger.Error(err, "encoding error response")
	}
}

// EncodeInvalidInputResponse encodes a generic 400 error to a response.
func (ed *serverEncoderDecoder) EncodeInvalidInputResponse(ctx context.Context, res http.ResponseWriter) {
	ed.tracer.StartSpan(ctx)

	ed.EncodeErrorResponse(ctx, res, "invalid input attached to request", http.StatusBadRequest)
}

// EncodeNotFoundResponse encodes a generic 404 error to a response.
func (ed *serverEncoderDecoder) EncodeNotFoundResponse(ctx context.Context, res http.ResponseWriter) {
	ctx, span := ed.tracer.StartSpan(ctx)
	defer span.End()

	ed.EncodeErrorResponse(ctx, res, "resource not found", http.StatusNotFound)
}

// EncodeUnspecifiedInternalServerErrorResponse encodes a generic 500 error to a response.
func (ed *serverEncoderDecoder) EncodeUnspecifiedInternalServerErrorResponse(ctx context.Context, res http.ResponseWriter) {
	ctx, span := ed.tracer.StartSpan(ctx)
	defer span.End()

	ed.EncodeErrorResponse(ctx, res, "something has gone awry", http.StatusInternalServerError)
}

// EncodeUnauthorizedResponse encodes a generic 401 error to a response.
func (ed *serverEncoderDecoder) EncodeUnauthorizedResponse(ctx context.Context, res http.ResponseWriter) {
	ctx, span := ed.tracer.StartSpan(ctx)
	defer span.End()

	ed.EncodeErrorResponse(ctx, res, "invalid credentials provided", http.StatusUnauthorized)
}

// EncodeInvalidPermissionsResponse encodes a generic 403 error to a response.
func (ed *serverEncoderDecoder) EncodeInvalidPermissionsResponse(ctx context.Context, res http.ResponseWriter) {
	ctx, span := ed.tracer.StartSpan(ctx)
	defer span.End()

	ed.EncodeErrorResponse(ctx, res, "invalid permissions", http.StatusForbidden)
}

// EncodeResponse encodes responses.
func (ed *serverEncoderDecoder) encodeResponse(ctx context.Context, res http.ResponseWriter, v interface{}, statusCode int) {
	_, span := ed.tracer.StartSpan(ctx)
	defer span.End()

	var ct = strings.ToLower(res.Header().Get(ContentTypeHeader))
	if ct == "" {
		ct = DefaultContentType
	}

	var e encoder

	switch ct {
	case XMLContentType:
		e = xml.NewEncoder(res)
	default:
		e = json.NewEncoder(res)
	}

	res.Header().Set(ContentTypeHeader, ct)
	res.WriteHeader(statusCode)

	if err := e.Encode(v); err != nil {
		ed.logger.Error(err, "encoding response")
	}
}

// EncodeResponse encodes successful responses.
func (ed *serverEncoderDecoder) EncodeResponse(ctx context.Context, res http.ResponseWriter, v interface{}) {
	ctx, span := ed.tracer.StartSpan(ctx)
	defer span.End()

	ed.encodeResponse(ctx, res, v, http.StatusOK)
}

// EncodeResponseWithStatus encodes responses and writes the provided status to the response.
func (ed *serverEncoderDecoder) EncodeResponseWithStatus(ctx context.Context, res http.ResponseWriter, v interface{}, statusCode int) {
	ctx, span := ed.tracer.StartSpan(ctx)
	defer span.End()

	ed.encodeResponse(ctx, res, v, statusCode)
}

// DecodeRequest decodes responses.
func (ed *serverEncoderDecoder) DecodeRequest(ctx context.Context, req *http.Request, v interface{}) error {
	_, span := ed.tracer.StartSpan(ctx)
	defer span.End()

	var ct = strings.ToLower(req.Header.Get(ContentTypeHeader))
	if ct == "" {
		ct = DefaultContentType
	}

	var d decoder

	switch ct {
	case XMLContentType:
		d = xml.NewDecoder(req.Body)
	default:
		dec := json.NewDecoder(req.Body)
		// this could be cool, but it would also break a lot of how my client works
		// dec.DisallowUnknownFields()
		d = dec
	}

	return d.Decode(v)
}

const name = "response_encoder"

// ProvideEncoderDecoder provides an EncoderDecoder.
func ProvideEncoderDecoder(logger logging.Logger) EncoderDecoder {
	return &serverEncoderDecoder{
		logger: logger.WithName(name),
		tracer: tracing.NewTracer(name),
	}
}
