package encoding

import (
	"encoding/json"
	"net/http"

	"github.com/google/wire"
)

var (
	// Providers provides ResponseEncoders for dependency injection
	Providers = wire.NewSet(
		ProvideJSONResponseEncoder,
	)
)

type ServerEncoder interface {
	EncodeResponse(http.ResponseWriter, interface{}) error
}

type ServerDecoder interface {
	DecodeResponse(*http.Request, interface{}) error
}

// EncoderDecoder is an interface that allows for multiple implementations of HTTP response formats
type EncoderDecoder interface {
	ServerDecoder
	ServerEncoder
}

// jsonResponseEncoder is a dummy struct that implements our EncoderDecoder interface
type jsonResponseEncoder struct{}

// EncodeResponse encodes responses for JSON types
func (r *jsonResponseEncoder) EncodeResponse(res http.ResponseWriter, v interface{}) error {
	res.Header().Set("Content-type", "application/json")
	return json.NewEncoder(res).Encode(v)
}

// DecodeResponse decodes responses from JSON types
func (r *jsonResponseEncoder) DecodeResponse(req *http.Request, v interface{}) error {
	return json.NewDecoder(req.Body).Decode(v)
}

// ProvideJSONResponseEncoder provides a jsonResponseEncoder
func ProvideJSONResponseEncoder() EncoderDecoder {
	return &jsonResponseEncoder{}
}
