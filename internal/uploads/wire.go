package uploads

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/storage"

	"github.com/google/wire"
)

var (
	// Providers represents this package's offering to the dependency manager.
	Providers = wire.NewSet(
		ProvideUploadManager,
		wire.FieldsOf(new(*Config),
			"Storage",
		),
	)
)

// ProvideUploadManager transforms a *storage.Uploader into an UploadManager.
func ProvideUploadManager(u *storage.Uploader) UploadManager {
	return u
}
