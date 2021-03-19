package images

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"io/ioutil"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	headerContentType = "Content-Type"

	imagePNG  = "image/png"
	imageJPEG = "image/jpeg"
	imageGIF  = "image/gif"

	// parsedImageContextKey is what we use to attach parsed images to requests.
	parsedImageContextKey types.ContextKey = "parsed_image"
)

var (
	// ErrInvalidImageContentType is what we return to indicate the provided image was of the wrong type.
	ErrInvalidImageContentType = errors.New("invalid content type")
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
		BuildAvatarUploadMiddleware(next http.Handler, encoderDecoder encoding.ServerEncoderDecoder, filename string) http.Handler
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
	w.Header().Set("Content-Length", strconv.Itoa(i.Size))

	if _, err := w.Write(i.Data); err != nil {
		return fmt.Errorf("writing image to HTTP response: %w", err)
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
		return fmt.Errorf("invalid content type: %s", contentType)
	}
}

// Process extracts an image from an *http.Request.
func (p *uploadProcessor) Process(ctx context.Context, req *http.Request, filename string) (*Image, error) {
	_, span := p.tracer.StartSpan(ctx)
	defer span.End()

	logger := p.logger.WithRequest(req)

	file, info, err := req.FormFile(filename)
	if err != nil {
		logger.Error(err, "parsing file from request")
		return nil, fmt.Errorf("reading image from request: %w", err)
	}

	if contentTypeErr := validateContentType(info.Filename); contentTypeErr != nil {
		logger.Error(contentTypeErr, "validating the content type")
		return nil, contentTypeErr
	}

	bs, err := ioutil.ReadAll(file)
	if err != nil {
		logger.Error(err, "reading file from request")
		return nil, fmt.Errorf("reading attached image: %w", err)
	}

	if _, _, err = image.Decode(bytes.NewReader(bs)); err != nil {
		logger.Error(err, "decoding the image data")
		return nil, fmt.Errorf("decoding attached image: %w", err)
	}

	i := &Image{
		Filename:    info.Filename,
		ContentType: contentTypeFromFilename(filename),
		Data:        bs,
		Size:        len(bs),
	}

	return i, nil
}

// BuildAvatarUploadMiddleware ensures that an image is attached to the request.
func (p *uploadProcessor) BuildAvatarUploadMiddleware(next http.Handler, encoderDecoder encoding.ServerEncoderDecoder, filename string) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx, span := p.tracer.StartSpan(req.Context())
		defer span.End()

		logger := p.logger.WithRequest(req)
		logger.Debug("avatar upload middleware invoked")

		file, info, err := req.FormFile(filename)
		if err != nil {
			logger.Error(err, "parsing request form")
			encoderDecoder.EncodeInvalidInputResponse(ctx, res)
			return
		}

		logger.Debug("content type check passed")

		bs, err := ioutil.ReadAll(file)
		if err != nil {
			logger.Error(err, "reading attached file")
			encoderDecoder.EncodeInvalidInputResponse(ctx, res)
			return
		}

		logger.Debug("image upload read from request")

		if _, _, err = image.Decode(bytes.NewReader(bs)); err != nil {
			logger.Error(err, "decoding attached image")
			encoderDecoder.EncodeInvalidInputResponse(ctx, res)
			return
		}

		logger.Debug("image decoded successfully")

		i := &Image{
			Filename:    info.Filename,
			ContentType: contentTypeFromFilename(info.Filename),
			Data:        bs,
			Size:        len(bs),
		}

		req = req.WithContext(context.WithValue(ctx, parsedImageContextKey, i))

		logger.Debug("attached image to context")

		next.ServeHTTP(res, req)
	})
}
