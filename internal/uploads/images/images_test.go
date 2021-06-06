package images

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	testutils "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

func buildPNGBytes(t *testing.T) *bytes.Buffer {
	t.Helper()

	b := new(bytes.Buffer)
	exampleImage := testutils.BuildArbitraryImage(256)
	require.NoError(t, png.Encode(b, exampleImage))

	expected := b.Bytes()
	return bytes.NewBuffer(expected)
}

func buildJPEGBytes(t *testing.T) *bytes.Buffer {
	t.Helper()

	b := new(bytes.Buffer)
	exampleImage := testutils.BuildArbitraryImage(256)
	require.NoError(t, jpeg.Encode(b, exampleImage, &jpeg.Options{Quality: jpeg.DefaultQuality}))

	expected := b.Bytes()
	return bytes.NewBuffer(expected)
}

func buildGIFBytes(t *testing.T) *bytes.Buffer {
	t.Helper()

	b := new(bytes.Buffer)
	exampleImage := testutils.BuildArbitraryImage(256)
	require.NoError(t, gif.Encode(b, exampleImage, &gif.Options{NumColors: 256}))

	expected := b.Bytes()
	return bytes.NewBuffer(expected)
}

func TestImage_DataURI(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		i := &Image{
			Filename:    t.Name(),
			ContentType: "things/stuff",
			Data:        []byte(t.Name()),
			Size:        12345,
		}

		expected := "data:things/stuff;base64,VGVzdEltYWdlX0RhdGFVUkkvc3RhbmRhcmQ="
		actual := i.DataURI()

		assert.Equal(t, expected, actual)
	})
}

func TestImage_Write(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		i := &Image{
			Filename:    t.Name(),
			ContentType: "things/stuff",
			Data:        []byte(t.Name()),
			Size:        12345,
		}

		res := httptest.NewRecorder()
		assert.NoError(t, i.Write(res))
	})

	T.Run("with write error", func(t *testing.T) {
		t.Parallel()

		i := &Image{
			Filename:    t.Name(),
			ContentType: "things/stuff",
			Data:        []byte(t.Name()),
			Size:        12345,
		}

		res := &testutils.MockHTTPResponseWriter{}
		res.On("Header").Return(http.Header{})
		res.On("Write", mock.IsType([]byte(nil))).Return(0, errors.New("blah"))

		assert.Error(t, i.Write(res))
	})
}

func TestImage_Thumbnail(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		imgBytes := buildPNGBytes(t).Bytes()

		i := &Image{
			Filename:    t.Name(),
			ContentType: imagePNG,
			Data:        imgBytes,
			Size:        len(imgBytes),
		}

		tempFile, err := os.CreateTemp("", "")
		require.NoError(t, err)

		actual, err := i.Thumbnail(123, 123, tempFile.Name())
		assert.NoError(t, err)
		assert.NotNil(t, actual)

		require.NoError(t, os.Remove(tempFile.Name()))
	})

	T.Run("with invalid content type", func(t *testing.T) {
		t.Parallel()

		i := &Image{
			ContentType: t.Name(),
		}

		actual, err := i.Thumbnail(123, 123, t.Name())
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestLimitFileSize(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		imgBytes := buildPNGBytes(t)
		req := newAvatarUploadRequest(t, "avatar.png", imgBytes)
		res := httptest.NewRecorder()

		LimitFileSize(0, res, req)
	})
}

func Test_uploadProcessor_Process(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		p := NewImageUploadProcessor(nil)
		expectedFieldName := "avatar"

		imgBytes := buildPNGBytes(t)
		expected := imgBytes.Bytes()

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

		imgBytes := buildPNGBytes(t)

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
