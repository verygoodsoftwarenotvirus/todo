package encoding

import (
	"github.com/google/wire"
)

var (
	// Providers provides ResponseEncoders for dependency injection.
	Providers = wire.NewSet(
		ProvideServerEncoderDecoder,
	)
)
