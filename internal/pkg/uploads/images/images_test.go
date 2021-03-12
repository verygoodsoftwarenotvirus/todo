package images

import (
	"bytes"
	"context"
	"fmt"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"
)

func newAvatarUploadRequest(t *testing.T, filename string, avatar io.Reader) *http.Request {
	t.Helper()

	ctx := context.Background()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("avatar", fmt.Sprintf("avatar.%s", filepath.Ext(filename)))
	require.NoError(t, err)

	_, err = io.Copy(part, avatar)
	require.NoError(t, err)

	require.NoError(t, writer.Close())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://tests.verygoodsoftwarenotvirus.ru", body)
	require.NoError(t, err)

	req.Header.Set(headerContentType, writer.FormDataContentType())

	return req
}

func Test_uploadProcessor_Process(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		p := NewImageUploadProcessor(nil)
		expectedFieldName := "avatar"

		b := new(bytes.Buffer)
		exampleImage := testutil.BuildArbitraryImage(256)
		require.NoError(t, png.Encode(b, exampleImage))

		expected := b.Bytes()
		imgBytes := bytes.NewBuffer(expected)

		req := newAvatarUploadRequest(t, "avatar.png", imgBytes)

		actual, err := p.Process(ctx, req, expectedFieldName)
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		assert.Equal(t, expected, actual.Data)
	})

	T.Run("with missing form file", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		p := NewImageUploadProcessor(nil)
		expectedFieldName := "avatar"

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://tests.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)

		actual, err := p.Process(ctx, req, expectedFieldName)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with invalid content type", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		p := NewImageUploadProcessor(nil)
		expectedFieldName := "avatar"

		b := new(bytes.Buffer)
		exampleImage := testutil.BuildArbitraryImage(256)
		require.NoError(t, png.Encode(b, exampleImage))

		expected := b.Bytes()
		imgBytes := bytes.NewBuffer(expected)

		req := newAvatarUploadRequest(t, "avatar.pizza", imgBytes)

		actual, err := p.Process(ctx, req, expectedFieldName)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	T.Run("with error decoding image", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		p := NewImageUploadProcessor(nil)
		expectedFieldName := "avatar"

		req := newAvatarUploadRequest(t, "avatar.png", bytes.NewBufferString(""))

		actual, err := p.Process(ctx, req, expectedFieldName)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}
