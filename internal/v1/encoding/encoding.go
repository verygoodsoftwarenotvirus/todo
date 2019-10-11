package encoding

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"strings"

	"github.com/google/wire"
)

const (
	// ContentTypeHeader is the HTTP standard header name for content type
	ContentTypeHeader = "Content-type"
	// XMLContentType represents the XML content type
	XMLContentType = "application/xml"
	// JSONContentType represents the JSON content type
	JSONContentType = "application/json"
	// DefaultContentType is what the library defaults to
	DefaultContentType = JSONContentType
)

var (
	// Providers provides ResponseEncoders for dependency injection
	Providers = wire.NewSet(
		ProvideResponseEncoder,
	)
)

type (
	// EncoderDecoder is an interface that allows for multiple implementations of HTTP response formats
	EncoderDecoder interface {
		EncodeResponse(http.ResponseWriter, interface{}) error
		DecodeRequest(*http.Request, interface{}) error
	}

	// ServerEncoderDecoder is our concrete implementation of EncoderDecoder
	ServerEncoderDecoder struct{}

	encoder interface {
		Encode(v interface{}) error
	}

	decoder interface {
		Decode(v interface{}) error
	}
)

// EncodeResponse encodes responses
func (ed *ServerEncoderDecoder) EncodeResponse(res http.ResponseWriter, v interface{}) error {
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
	return e.Encode(v)
}

// DecodeRequest decodes responses
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

// ProvideResponseEncoder provides a jsonResponseEncoder
func ProvideResponseEncoder() EncoderDecoder {
	return &ServerEncoderDecoder{}
}
