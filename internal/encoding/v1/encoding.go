package encoding

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"strings"

	"github.com/google/wire"
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
		DecodeResponse(*http.Request, interface{}) error
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
	ct := strings.ToLower(res.Header().Get("Content-type"))

	var e encoder
	switch ct {
	case "application/xml":
		e = xml.NewEncoder(res)
	case "application/json":
		e = json.NewEncoder(res)
	default:
		ct = "application/json"
		e = json.NewEncoder(res)
	}

	res.Header().Set("Content-type", ct)
	return e.Encode(v)
}

// DecodeResponse decodes responses
func (ed *ServerEncoderDecoder) DecodeResponse(req *http.Request, v interface{}) error {
	ct := strings.ToLower(req.Header.Get("Content-type"))

	var d decoder
	switch ct {
	case "application/xml":
		d = xml.NewDecoder(req.Body)
	case "application/json":
		d = json.NewDecoder(req.Body)
	default:
		ct = "application/json"
		d = json.NewDecoder(req.Body)
	}

	return d.Decode(v)
}

// ProvideResponseEncoder provides a jsonResponseEncoder
func ProvideResponseEncoder() EncoderDecoder {
	return &ServerEncoderDecoder{}
}
