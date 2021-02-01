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

	// ParsedImageContextKey is what we use to attach parsed images to requests.
	ParsedImageContextKey types.ContextKey = "parsed_image"
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
		BuildAvatarUploadMiddleware(next http.Handler, logger logging.Logger, encoderDecoder encoding.EncoderDecoder, filename string) http.Handler
	}

	uploadProcessor struct {
		tracer tracing.Tracer
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
func NewImageUploadProcessor() ImageUploadProcessor {
	return &uploadProcessor{
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

	file, info, err := req.FormFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading image from request: %w", err)
	}

	if contentTypeErr := validateContentType(info.Filename); contentTypeErr != nil {
		return nil, contentTypeErr
	}

	bs, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("reading attached image: %w", err)
	}

	if _, _, err = image.Decode(bytes.NewReader(bs)); err != nil {
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
func (p *uploadProcessor) BuildAvatarUploadMiddleware(next http.Handler, logger logging.Logger, encoderDecoder encoding.EncoderDecoder, filename string) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx, span := p.tracer.StartSpan(req.Context())
		defer span.End()

		file, info, err := req.FormFile(filename)
		if err != nil {
			encoderDecoder.EncodeInvalidInputResponse(ctx, res)
			return
		}

		contentType := info.Header.Get(headerContentType)

		switch strings.TrimSpace(strings.ToLower(contentType)) {
		case imagePNG, imageJPEG, imageGIF:
			// we good
		default:
			logger.WithValue("invalid_content_type", contentType).Error(ErrInvalidImageContentType, "")
			encoderDecoder.EncodeInvalidInputResponse(ctx, res)
			return
		}

		bs, err := ioutil.ReadAll(file)
		if err != nil {
			logger.Error(err, "reading attached file")
			encoderDecoder.EncodeInvalidInputResponse(ctx, res)
			return
		}

		if _, _, err = image.Decode(bytes.NewReader(bs)); err != nil {
			logger.Error(err, "decoding attached image")
			encoderDecoder.EncodeInvalidInputResponse(ctx, res)
			return
		}

		i := &Image{
			Filename:    info.Filename,
			ContentType: contentType,
			Data:        bs,
			Size:        len(bs),
		}

		req = req.WithContext(context.WithValue(ctx, ParsedImageContextKey, i))

		next.ServeHTTP(res, req)
	})
}
