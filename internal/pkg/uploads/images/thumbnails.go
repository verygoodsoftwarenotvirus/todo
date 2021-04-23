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
	allSupportedColors = 2 << 7 // 256
)

type thumbnailer interface {
	Thumbnail(i *Image, width, height uint, filename string) (*Image, error)
}

// newThumbnailer provides a thumbnailer given a particular content type.
func newThumbnailer(contentType string) (thumbnailer, error) {
	switch strings.TrimSpace(strings.ToLower(contentType)) {
	case imagePNG:
		return &pngThumbnailer{}, nil
	case imageJPEG:
		return &jpegThumbnailer{}, nil
	case imageGIF:
		return &gifThumbnailer{}, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrInvalidImageContentType, contentType)
	}
}

func preprocess(i *Image, width, height uint) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(i.Data))
	if err != nil {
		return nil, fmt.Errorf("decoding image: %w", err)
	}

	thumbnail := resize.Thumbnail(width, height, img, resize.Lanczos3)

	return thumbnail, nil
}

type jpegThumbnailer struct{}

// Thumbnail creates a JPEG thumbnail.
func (t *jpegThumbnailer) Thumbnail(img *Image, width, height uint, filename string) (*Image, error) {
	thumbnail, err := preprocess(img, width, height)
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	if err = jpeg.Encode(&b, thumbnail, &jpeg.Options{Quality: jpeg.DefaultQuality}); err != nil {
		return nil, fmt.Errorf("encoding JPEG: %w", err)
	}

	bs := b.Bytes()

	i := &Image{
		Filename:    fmt.Sprintf("%s.jpg", filename),
		ContentType: imageJPEG,
		Data:        bs,
		Size:        len(bs),
	}

	return i, nil
}

type pngThumbnailer struct{}

// Thumbnail creates a PNG thumbnail.
func (t *pngThumbnailer) Thumbnail(img *Image, width, height uint, filename string) (*Image, error) {
	thumbnail, err := preprocess(img, width, height)
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	if err = gif.Encode(&b, thumbnail, &gif.Options{NumColors: allSupportedColors}); err != nil {
		return nil, fmt.Errorf("encoding JPEG: %w", err)
	}

	bs := b.Bytes()

	i := &Image{
		Filename:    fmt.Sprintf("%s.gif", filename),
		ContentType: imageGIF,
		Data:        bs,
		Size:        len(bs),
	}

	return i, nil
}

type gifThumbnailer struct{}

// Thumbnail creates a GIF thumbnail.
func (t *gifThumbnailer) Thumbnail(img *Image, width, height uint, filename string) (*Image, error) {
	thumbnail, err := preprocess(img, width, height)
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	if err = png.Encode(&b, thumbnail); err != nil {
		return nil, fmt.Errorf("encoding PNG: %w", err)
	}

	bs := b.Bytes()

	i := &Image{
		Filename:    fmt.Sprintf("%s.png", filename),
		ContentType: imagePNG,
		Data:        bs,
		Size:        len(bs),
	}

	return i, nil
}
