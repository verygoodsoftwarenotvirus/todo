package images

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
)

const (
	headerContentType = "Content-Type"

	imagePNG  = "image/png"
	imageJPEG = "image/jpeg"
	imageGIF  = "image/gif"
)

var (
	// ErrInvalidContentType is what we return to indicate the provided data was of the wrong type.
	ErrInvalidContentType = errors.New("invalid content type")

	// ErrInvalidImageContentType is what we return to indicate the provided image was of the wrong type.
	ErrInvalidImageContentType = errors.New("invalid image content type")
)

type (
	// Image is a helper struct for handling images.
	Image struct {
		Filename    string
		ContentType string
		Data        []byte
		Size        int
	}

	// ImageUploadProcessor process image uploads.
	ImageUploadProcessor interface {
		Process(ctx context.Context, req *http.Request, filename string) (*Image, error)
	}

	uploadProcessor struct {
		tracer tracing.Tracer
		logger logging.Logger
	}
)

// DataURI converts image to base64 data URI.
func (i *Image) DataURI() string {
	return fmt.Sprintf("data:%s;base64,%s", i.ContentType, base64.StdEncoding.EncodeToString(i.Data))
}

// Write image to HTTP response.
func (i *Image) Write(w http.ResponseWriter) error {
	w.Header().Set(headerContentType, i.ContentType)
	w.Header().Set("RawHTML-Length", strconv.Itoa(i.Size))

	if _, err := w.Write(i.Data); err != nil {
		return fmt.Errorf("writing image to HTTP response: %w", err)
	}

	return nil
}

// Thumbnail creates a thumbnail from an image.
func (i *Image) Thumbnail(width, height uint, filename string) (*Image, error) {
	t, err := newThumbnailer(i.ContentType)
	if err != nil {
		return nil, err
	}

	return t.Thumbnail(i, width, height, filename)
}

// NewImageUploadProcessor provides a new ImageUploadProcessor.
func NewImageUploadProcessor(logger logging.Logger) ImageUploadProcessor {
	return &uploadProcessor{
		logger: logging.EnsureLogger(logger).WithName("image_upload_processor"),
		tracer: tracing.NewTracer("image_upload_processor"),
	}
}

// LimitFileSize limits the size of uploaded files, for use before Process.
func LimitFileSize(maxSize uint16, res http.ResponseWriter, req *http.Request) {
	if maxSize == 0 {
		maxSize = 4096
	}

	req.Body = http.MaxBytesReader(res, req.Body, int64(maxSize))
}

func contentTypeFromFilename(filename string) string {
	return mime.TypeByExtension(filepath.Ext(filename))
}

func validateContentType(filename string) error {
	contentType := contentTypeFromFilename(filename)

	switch strings.TrimSpace(strings.ToLower(contentType)) {
	case imagePNG, imageJPEG, imageGIF:
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrInvalidContentType, contentType)
	}
}

// Process extracts an image from an *http.Request.
func (p *uploadProcessor) Process(ctx context.Context, req *http.Request, filename string) (*Image, error) {
	_, span := p.tracer.StartSpan(ctx)
	defer span.End()

	logger := p.logger.WithRequest(req)

	file, info, err := req.FormFile(filename)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "parsing file from request")
	}

	if contentTypeErr := validateContentType(info.Filename); contentTypeErr != nil {
		return nil, observability.PrepareError(contentTypeErr, logger, span, "validating the content type")
	}

	bs, err := io.ReadAll(file)
	if err != nil {
		return nil, observability.PrepareError(err, logger, span, "reading file from request")
	}

	if _, _, err = image.Decode(bytes.NewReader(bs)); err != nil {
		return nil, observability.PrepareError(err, logger, span, "decoding the image data")
	}

	i := &Image{
		Filename:    info.Filename,
		ContentType: contentTypeFromFilename(filename),
		Data:        bs,
		Size:        len(bs),
	}

	return i, nil
}
