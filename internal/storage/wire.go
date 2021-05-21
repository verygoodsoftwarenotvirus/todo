package storage

import "github.com/google/wire"

var (
	// Providers is what we provide to the dependency injection framework.
	Providers = wire.NewSet(
		NewUploadManager,
	)
)
