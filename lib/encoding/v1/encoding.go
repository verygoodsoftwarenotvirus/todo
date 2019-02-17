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

// ResponseEncoder is an interface that allows for multiple implementations of HTTP response formats
type ResponseEncoder interface {
	EncodeResponse(http.ResponseWriter, interface{}) error
}

// jsonResponseEncoder is a dummy struct that implements our ResponseEncoder interface
type jsonResponseEncoder struct{}

// EncodeResponse encodes responses for JSON types
func (r *jsonResponseEncoder) EncodeResponse(w http.ResponseWriter, v interface{}) error {
	w.Header().Set("Content-type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

// ProvideJSONResponseEncoder provides a jsonResponseEncoder
func ProvideJSONResponseEncoder() ResponseEncoder {
	return &jsonResponseEncoder{}
}
