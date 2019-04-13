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

// ServerEncoderDecoder is an interface that allows for multiple implementations of HTTP response formats
// RENAMEME
type ServerEncoderDecoder interface {
	DecodeResponse(*http.Request, interface{}) error
	EncodeResponse(http.ResponseWriter, interface{}) error
}

// jsonResponseEncoder is a dummy struct that implements our ServerEncoderDecoder interface
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
func ProvideJSONResponseEncoder() ServerEncoderDecoder {
	return &jsonResponseEncoder{}
}
