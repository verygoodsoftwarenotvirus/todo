package images

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_newThumbnailer(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		for _, ct := range []string{imagePNG, imageJPEG, imageGIF} {
			x, err := newThumbnailer(ct)
			assert.NoError(t, err)
			assert.NotNil(t, x)
		}
	})

	T.Run("invalid content type", func(t *testing.T) {
		t.Parallel()

		x, err := newThumbnailer(t.Name())
		assert.Error(t, err)
		assert.Nil(t, x)
	})
}

func Test_preprocess(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		imgBytes := buildPNGBytes(t).Bytes()
		i := &Image{
			Filename:    "whatever.png",
			ContentType: imagePNG,
			Data:        imgBytes,
			Size:        len(imgBytes),
		}

		img, err := preprocess(i, 128, 128)
		assert.NoError(t, err)
		assert.NotNil(t, img)
	})

	T.Run("with invalid content", func(t *testing.T) {
		t.Parallel()

		i := &Image{
			Filename:    "whatever.png",
			ContentType: imagePNG,
			Data:        []byte(t.Name()),
			Size:        1024,
		}

		img, err := preprocess(i, 128, 128)
		assert.Error(t, err)
		assert.Nil(t, img)
	})
}

func Test_jpegThumbnailer_Thumbnail(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		imgBytes := buildJPEGBytes(t).Bytes()
		i := &Image{
			Filename:    "whatever.png",
			ContentType: imagePNG,
			Data:        imgBytes,
			Size:        len(imgBytes),
		}

		tempFile, err := os.CreateTemp("", "")
		require.NoError(t, err)

		actual, err := (&jpegThumbnailer{}).Thumbnail(i, 128, 128, tempFile.Name())
		assert.NoError(t, err)
		assert.NotNil(t, actual)

		require.NoError(t, os.Remove(tempFile.Name()))
	})

	T.Run("with invalid content", func(t *testing.T) {
		t.Parallel()

		i := &Image{
			Filename:    "whatever.png",
			ContentType: imagePNG,
			Data:        []byte(t.Name()),
			Size:        1024,
		}

		tempFile, err := os.CreateTemp("", "")
		require.NoError(t, err)

		actual, err := (&jpegThumbnailer{}).Thumbnail(i, 128, 128, tempFile.Name())
		assert.Error(t, err)
		assert.Nil(t, actual)

		require.NoError(t, os.Remove(tempFile.Name()))
	})
}

func Test_pngThumbnailer_Thumbnail(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		imgBytes := buildPNGBytes(t).Bytes()
		i := &Image{
			Filename:    "whatever.png",
			ContentType: imagePNG,
			Data:        imgBytes,
			Size:        len(imgBytes),
		}

		tempFile, err := os.CreateTemp("", "")
		require.NoError(t, err)

		actual, err := (&pngThumbnailer{}).Thumbnail(i, 128, 128, tempFile.Name())
		assert.NoError(t, err)
		assert.NotNil(t, actual)

		require.NoError(t, os.Remove(tempFile.Name()))
	})

	T.Run("with invalid content", func(t *testing.T) {
		t.Parallel()

		i := &Image{
			Filename:    "whatever.png",
			ContentType: imagePNG,
			Data:        []byte(t.Name()),
			Size:        1024,
		}

		tempFile, err := os.CreateTemp("", "")
		require.NoError(t, err)

		actual, err := (&pngThumbnailer{}).Thumbnail(i, 128, 128, tempFile.Name())
		assert.Error(t, err)
		assert.Nil(t, actual)

		require.NoError(t, os.Remove(tempFile.Name()))
	})
}

func Test_gifThumbnailer_Thumbnail(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		imgBytes := buildGIFBytes(t).Bytes()
		i := &Image{
			Filename:    "whatever.png",
			ContentType: imagePNG,
			Data:        imgBytes,
			Size:        len(imgBytes),
		}

		tempFile, err := os.CreateTemp("", "")
		require.NoError(t, err)

		actual, err := (&gifThumbnailer{}).Thumbnail(i, 128, 128, tempFile.Name())
		assert.NoError(t, err)
		assert.NotNil(t, actual)

		require.NoError(t, os.Remove(tempFile.Name()))
	})

	T.Run("with invalid content", func(t *testing.T) {
		t.Parallel()

		i := &Image{
			Filename:    "whatever.png",
			ContentType: imagePNG,
			Data:        []byte(t.Name()),
			Size:        1024,
		}

		tempFile, err := os.CreateTemp("", "")
		require.NoError(t, err)

		actual, err := (&gifThumbnailer{}).Thumbnail(i, 128, 128, tempFile.Name())
		assert.Error(t, err)
		assert.Nil(t, actual)

		require.NoError(t, os.Remove(tempFile.Name()))
	})
}
