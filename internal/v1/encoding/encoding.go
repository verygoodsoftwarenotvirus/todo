package encoding

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/google/wire"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"
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
)

type (
	// EncoderDecoder is an interface that allows for multiple implementations of HTTP response formats.
	EncoderDecoder interface {
		EncodeResponse(res http.ResponseWriter, val interface{})
		EncodeError(res http.ResponseWriter, msg string, code int)
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

// EncodeError encodes errors to responses.
func (ed *ServerEncoderDecoder) EncodeError(res http.ResponseWriter, msg string, code int) {
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

	if http.StatusText(code) != "" {
		res.WriteHeader(code)
	}

	if err := e.Encode(&models.ErrorResponse{Message: msg, Code: code}); err != nil {
		ed.logger.Error(err, "encoding error response")
	}
}

// EncodeResponse encodes responses.
func (ed *ServerEncoderDecoder) EncodeResponse(res http.ResponseWriter, v interface{}) {
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

	if err := e.Encode(v); err != nil {
		ed.logger.Error(err, "encoding response")
	}
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
