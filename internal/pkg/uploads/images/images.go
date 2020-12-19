package images

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"strings"

	"io/ioutil"
	"net/http"
	"strconv"
)

const (
	headerContentType = "Content-Type"

	imagePNG  = "image/png"
	imageJPEG = "image/jpeg"
	imageGIF  = "image/gif"
)

// Image is a helper struct for handling images.
type Image struct {
	Filename    string
	ContentType string
	Data        []byte
	Size        int
}

// DataURI converts image to base64 data URI.
func (i *Image) DataURI() string {
	return fmt.Sprintf("data:%s;base64,%s", i.ContentType, base64.StdEncoding.EncodeToString(i.Data))
}

// Write image to HTTP response.
func (i *Image) Write(w http.ResponseWriter) error {
	w.Header().Set(headerContentType, i.ContentType)
	w.Header().Set("Content-Length", strconv.Itoa(i.Size))

	if _, err := w.Write(i.Data); err != nil {
		return fmt.Errorf("error writing image to HTTP response: %w", err)
	}

	return nil
}

// Thumbnail creates a thumbnail from an image.
func (i *Image) Thumbnail(width, height uint, quality int, filename string) (*Image, error) {
	t, err := newThumbnailer(i.ContentType, quality)
	if err != nil {
		return nil, err
	}

	return t.Thumbnail(i, width, height, filename)
}

// ImageUploadProcessor process image uploads.
type ImageUploadProcessor interface {
	Process(r *http.Request, filename string) (*Image, error)
}

type imageUploadProcessor struct {
}

// NewImageUploadProcessor provides a new ImageUploadProcessor.
func NewImageUploadProcessor() ImageUploadProcessor {
	return &imageUploadProcessor{}
}

// LimitFileSize limits the size of uploaded files, for use before Process.
func (p *imageUploadProcessor) LimitFileSize(maxSize int64, res http.ResponseWriter, req *http.Request) {
	req.Body = http.MaxBytesReader(res, req.Body, maxSize)
}

// Process extracts an image from an *http.Request.
func (p *imageUploadProcessor) Process(req *http.Request, filename string) (*Image, error) {
	file, info, err := req.FormFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading image from request: %w", err)
	}

	contentType := info.Header.Get(headerContentType)

	switch strings.TrimSpace(strings.ToLower(contentType)) {
	case imagePNG, imageJPEG, imageGIF:
		// we good
	default:
		return nil, fmt.Errorf("invalid content type: %s", contentType)
	}

	bs, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading attached image: %w", err)
	}

	if _, _, err = image.Decode(bytes.NewReader(bs)); err != nil {
		return nil, fmt.Errorf("error decoding attached image: %w", err)
	}

	i := &Image{
		Filename:    info.Filename,
		ContentType: contentType,
		Data:        bs,
		Size:        len(bs),
	}

	return i, nil
}
