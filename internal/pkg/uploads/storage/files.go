package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
)

// SaveFile saves a file to the blob.
func (u *Uploader) SaveFile(ctx context.Context, path string, content []byte) error {
	ctx, span := u.tracer.StartSpan(ctx)
	defer span.End()

	if err := u.bucket.WriteAll(ctx, path, content, nil); err != nil {
		return fmt.Errorf("writing file content: %w", err)
	}

	return nil
}

// ReadFile reads a file from the blob.
func (u *Uploader) ReadFile(ctx context.Context, path string) ([]byte, error) {
	ctx, span := u.tracer.StartSpan(ctx)
	defer span.End()

	r, err := u.bucket.NewReader(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("fetching file: %w", err)
	}

	defer func() {
		if closeErr := r.Close(); closeErr != nil {
			u.logger.Error(closeErr, "error closing file reader")
		}
	}()

	fileBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	return fileBytes, nil
}

// ServeFiles saves a file to the blob.
func (u *Uploader) ServeFiles(res http.ResponseWriter, req *http.Request) {
	ctx, span := u.tracer.StartSpan(req.Context())
	defer span.End()

	fileName := u.filenameFetcher(req)

	fileBytes, err := u.ReadFile(ctx, fileName)
	if err != nil {
		u.logger.Error(err, "trying to read uploaded file")
		res.WriteHeader(http.StatusNotFound)

		return
	}

	if attrs, err := u.bucket.Attributes(ctx, fileName); attrs != nil && err == nil {
		res.Header().Set(encoding.ContentTypeHeaderKey, attrs.ContentType)
	}

	if _, copyErr := io.Copy(res, bytes.NewReader(fileBytes)); copyErr != nil {
		u.logger.Error(copyErr, "copying file bytes to response")
	}
}
