package encoding

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/google/wire"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
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
		ProvideResponseEncoder,
	)

	_ EncoderDecoder = (*ServerEncoderDecoder)(nil)
)

type (
	// EncoderDecoder is an interface that allows for multiple implementations of HTTP response formats.
	EncoderDecoder interface {
		EncodeResponse(res http.ResponseWriter, val interface{})
		EncodeResponseWithStatus(res http.ResponseWriter, val interface{}, statusCode int)
		EncodeErrorResponse(res http.ResponseWriter, msg string, statusCode int)
		EncodeNoInputResponse(res http.ResponseWriter)
		EncodeNotFoundResponse(res http.ResponseWriter)
		EncodeUnspecifiedInternalServerErrorResponse(res http.ResponseWriter)
		EncodeUnauthorizedResponse(res http.ResponseWriter)
		DecodeRequest(req *http.Request, dest interface{}) error
	}

	// ServerEncoderDecoder is our concrete implementation of EncoderDecoder.
	ServerEncoderDecoder struct {
		logger logging.Logger
	}

	encoder interface {
		Encode(v interface{}) error
	}

	decoder interface {
		Decode(v interface{}) error
	}
)

// EncodeErrorResponse encodes errors to responses.
func (ed *ServerEncoderDecoder) EncodeErrorResponse(res http.ResponseWriter, msg string, statusCode int) {
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

	if err := e.Encode(&models.ErrorResponse{Message: msg, Code: statusCode}); err != nil {
		ed.logger.Error(err, "encoding error response")
	}
}

// EncodeNoInputResponse encodes a generic 400 error to a response.
func (ed *ServerEncoderDecoder) EncodeNoInputResponse(res http.ResponseWriter) {
	ed.EncodeErrorResponse(res, "no input attached to request", http.StatusBadRequest)
}

// EncodeNotFoundResponse encodes a generic 404 error to a response.
func (ed *ServerEncoderDecoder) EncodeNotFoundResponse(res http.ResponseWriter) {
	ed.EncodeErrorResponse(res, "resource not found", http.StatusNotFound)
}

// EncodeUnspecifiedInternalServerErrorResponse encodes a generic 500 error to a response.
func (ed *ServerEncoderDecoder) EncodeUnspecifiedInternalServerErrorResponse(res http.ResponseWriter) {
	ed.EncodeErrorResponse(res, "something has gone awry", http.StatusInternalServerError)
}

// EncodeUnauthorizedResponse encodes a generic 401 error to a response.
func (ed *ServerEncoderDecoder) EncodeUnauthorizedResponse(res http.ResponseWriter) {
	ed.EncodeErrorResponse(res, "invalid credentials provided", http.StatusUnauthorized)
}

// EncodeResponse encodes responses.
func (ed *ServerEncoderDecoder) encodeResponse(res http.ResponseWriter, v interface{}, statusCode int) {
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
func (ed *ServerEncoderDecoder) EncodeResponse(res http.ResponseWriter, v interface{}) {
	ed.encodeResponse(res, v, http.StatusOK)
}

// EncodeResponseWithStatus encodes responses and writes the provided status to the response.
func (ed *ServerEncoderDecoder) EncodeResponseWithStatus(res http.ResponseWriter, v interface{}, statusCode int) {
	ed.encodeResponse(res, v, statusCode)
}

// DecodeRequest decodes responses.
func (ed *ServerEncoderDecoder) DecodeRequest(req *http.Request, v interface{}) error {
	var ct = strings.ToLower(req.Header.Get(ContentTypeHeader))
	if ct == "" {
		ct = DefaultContentType
	}

	var d decoder
	switch ct {
	case XMLContentType:
		d = xml.NewDecoder(req.Body)
	default:
		d = json.NewDecoder(req.Body)
	}

	return d.Decode(v)
}

// ProvideResponseEncoder provides a jsonResponseEncoder.
func ProvideResponseEncoder(logger logging.Logger) EncoderDecoder {
	return &ServerEncoderDecoder{
		logger: logger.WithName("response_encoder"),
	}
}
