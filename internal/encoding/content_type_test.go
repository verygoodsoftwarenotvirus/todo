package encoding

import (
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"

	"github.com/stretchr/testify/assert"
)

func Test_clientEncoder_ContentType(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		e := ProvideClientEncoder(logging.NewNoopLogger(), ContentTypeJSON)

		assert.NotEmpty(t, e.ContentType())
	})
}

func Test_buildContentType(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		//
	})
}

func Test_contentTypeToString(T *testing.T) {
	T.Parallel()

	T.Run("with JSON", func(t *testing.T) {
		t.Parallel()

		assert.NotEmpty(t, contentTypeToString(ContentTypeJSON))
	})

	T.Run("with XML", func(t *testing.T) {
		t.Parallel()

		assert.NotEmpty(t, contentTypeToString(ContentTypeXML))
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, contentTypeToString(nil))
	})
}

func Test_contentTypeFromString(T *testing.T) {
	T.Parallel()

	T.Run("with JSON", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, ContentTypeJSON, contentTypeFromString(contentTypeJSON))
	})

	T.Run("with XML", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, ContentTypeXML, contentTypeFromString(contentTypeXML))
	})
}
