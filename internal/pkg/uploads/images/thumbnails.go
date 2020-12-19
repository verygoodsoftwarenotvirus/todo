package images

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"strings"

	"github.com/nfnt/resize"
)

const (
	allSupportedColors = 2 << 7
)

type thumbnailer interface {
	Thumbnail(i *Image, width, height uint, filename string) (*Image, error)
}

// newThumbnailer provides a thumbnailer given a particular content type.
func newThumbnailer(contentType string, quality int) (thumbnailer, error) {
	switch strings.TrimSpace(strings.ToLower(contentType)) {
	case imagePNG:
		return &pngThumbnailer{}, nil
	case imageJPEG:
		return &jpegThumbnailer{quality: quality}, nil
	case imageGIF:
		return &gifThumbnailer{}, nil
	default:
		return nil, fmt.Errorf("invalid content type: %s", contentType)
	}
}

func preprocess(i *Image, width, height uint) (*bytes.Buffer, image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(i.Data))
	if err != nil {
		return nil, nil, fmt.Errorf("error decoding image: %w", err)
	}

	thumbnail := resize.Thumbnail(width, height, img, resize.Lanczos3)
	data := new(bytes.Buffer)

	return data, thumbnail, nil
}

type jpegThumbnailer struct {
	quality int
}

// Thumbnail creates a JPEG thumbnail.
func (t *jpegThumbnailer) Thumbnail(i *Image, width, height uint, filename string) (*Image, error) {
	data, thumbnail, err := preprocess(i, width, height)
	if err != nil {
		return nil, err
	}

	if err = jpeg.Encode(data, thumbnail, &jpeg.Options{Quality: t.quality}); err != nil {
		return nil, fmt.Errorf("error encoding JPEG: %w", err)
	}

	bs := data.Bytes()

	return &Image{
		Filename:    fmt.Sprintf("%s.jpg", filename),
		ContentType: imageJPEG,
		Data:        bs,
		Size:        len(bs),
	}, nil
}

type pngThumbnailer struct{}

// Thumbnail creates a GIF thumbnail.
func (t *pngThumbnailer) Thumbnail(i *Image, width, height uint, filename string) (*Image, error) {
	data, thumbnail, err := preprocess(i, width, height)
	if err != nil {
		return nil, err
	}

	if err = gif.Encode(data, thumbnail, &gif.Options{NumColors: allSupportedColors}); err != nil {
		return nil, fmt.Errorf("error encoding JPEG: %w", err)
	}

	bs := data.Bytes()

	return &Image{
		Filename:    fmt.Sprintf("%s.gif", filename),
		ContentType: imageGIF,
		Data:        bs,
		Size:        len(bs),
	}, nil
}

type gifThumbnailer struct{}

// Thumbnail creates a PNG thumbnail.
func (t *gifThumbnailer) Thumbnail(i *Image, width, height uint, filename string) (*Image, error) {
	data, thumbnail, err := preprocess(i, width, height)
	if err != nil {
		return nil, err
	}

	if err = png.Encode(data, thumbnail); err != nil {
		return nil, fmt.Errorf("error encoding PNG: %w", err)
	}

	bs := data.Bytes()

	return &Image{
		Filename:    fmt.Sprintf("%s.png", filename),
		ContentType: imagePNG,
		Data:        bs,
		Size:        len(bs),
	}, nil
}
