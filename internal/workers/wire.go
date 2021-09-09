package workers

import "github.com/google/wire"

var (
	Providers = wire.NewSet(
		ProvidePendingWriter,
	)
)
