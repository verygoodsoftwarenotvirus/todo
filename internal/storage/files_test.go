package storage

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	testutils "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gocloud.dev/blob/memblob"
)

func TestUploader_ReadFile(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleFilename := "hello_world.txt"

		b := memblob.OpenBucket(&memblob.Options{})
		require.NoError(t, b.WriteAll(ctx, exampleFilename, []byte(t.Name()), nil))

		u := &Uploader{
			bucket: b,
			logger: logging.NewNoopLogger(),
			tracer: tracing.NewTracer(t.Name()),
			filenameFetcher: func(*http.Request) string {
				return t.Name()
			},
		}

		x, err := u.ReadFile(ctx, exampleFilename)
		assert.NoError(t, err)
		assert.NotNil(t, x)
	})

	T.Run("with invalid file", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleFilename := "hello_world.txt"

		u := &Uploader{
			bucket: memblob.OpenBucket(&memblob.Options{}),
			logger: logging.NewNoopLogger(),
			tracer: tracing.NewTracer(t.Name()),
			filenameFetcher: func(*http.Request) string {
				return t.Name()
			},
		}

		x, err := u.ReadFile(ctx, exampleFilename)
		assert.Error(t, err)
		assert.Nil(t, x)
	})
}

func TestUploader_SaveFile(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		tempFile, err := ioutil.TempFile("", "")
		require.NoError(t, err)

		ctx := context.Background()
		u := &Uploader{
			bucket: memblob.OpenBucket(&memblob.Options{}),
			logger: logging.NewNoopLogger(),
			tracer: tracing.NewTracer(t.Name()),
			filenameFetcher: func(*http.Request) string {
				return t.Name()
			},
		}

		assert.NoError(t, u.SaveFile(ctx, tempFile.Name(), []byte(t.Name())))
	})
}

func TestUploader_ServeFiles(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleFilename := "hello_world.txt"

		b := memblob.OpenBucket(&memblob.Options{})
		require.NoError(t, b.WriteAll(ctx, exampleFilename, []byte(t.Name()), nil))

		u := &Uploader{
			bucket: b,
			logger: logging.NewNoopLogger(),
			tracer: tracing.NewTracer(t.Name()),
			filenameFetcher: func(*http.Request) string {
				return exampleFilename
			},
		}

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/things", nil)

		u.ServeFiles(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
	})

	T.Run("with nonexistent file", func(t *testing.T) {
		t.Parallel()

		exampleFilename := "hello_world.txt"

		u := &Uploader{
			bucket: memblob.OpenBucket(&memblob.Options{}),
			logger: logging.NewNoopLogger(),
			tracer: tracing.NewTracer(t.Name()),
			filenameFetcher: func(*http.Request) string {
				return exampleFilename
			},
		}

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/things", nil)

		u.ServeFiles(res, req)

		assert.Equal(t, http.StatusNotFound, res.Code)
	})

	T.Run("with error writing file content", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleFilename := "hello_world.txt"

		b := memblob.OpenBucket(&memblob.Options{})
		require.NoError(t, b.WriteAll(ctx, exampleFilename, []byte(t.Name()), nil))

		u := &Uploader{
			bucket: b,
			logger: logging.NewNoopLogger(),
			tracer: tracing.NewTracer(t.Name()),
			filenameFetcher: func(*http.Request) string {
				return exampleFilename
			},
		}

		res := &testutils.MockHTTPResponseWriter{}
		res.On("Write", mock.IsType([]byte(nil))).Return(0, errors.New("blah"))
		res.On("Header").Return(http.Header{})
		req := httptest.NewRequest(http.MethodGet, "/things", nil)

		u.ServeFiles(res, req)

		mock.AssertExpectationsForObjects(t, res)
	})
}
