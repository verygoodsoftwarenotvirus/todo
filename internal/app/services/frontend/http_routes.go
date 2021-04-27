package frontend

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
)

const (
	loginRoute        = "/auth/login"
	registrationRoute = "/auth/register"
)

var (
	validRoutes = map[string]struct{}{
		// entry routes
		loginRoute:        {},
		registrationRoute: {},
		// admin routes
		"/admin":           {},
		"/admin/dashboard": {},
		"/admin/users":     {},
		"/admin/webhooks":  {},
		"/admin/audit_log": {},
		"/admin/settings":  {},
		// user routes
		"/items":                {},
		"/items/new":            {},
		"/things/items":         {},
		"/things/items/new":     {},
		"/dashboard":            {},
		"/account/webhooks":     {},
		"/account/webhooks/new": {},
		"/account/settings":     {},
		"/user/api_clients":     {},
		"/user/api_clients/new": {},
		"/user/settings":        {},
	}
)

func shouldRedirectToHome(path string) bool {
	for rt := range validRoutes {
		if strings.HasPrefix(path, rt) {
			return true
		}
	}

	return false
}

func (s *service) cacheFile(ctx context.Context, afs afero.Fs, filePath string) error {
	_, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue("file_path", filePath)

	f, err := afs.Create(filePath)
	if err != nil {
		return observability.PrepareError(err, logger, span, "creating static file in memory")
	}

	bs, err := os.ReadFile(filePath)
	if err != nil {
		return observability.PrepareError(err, logger, span, "reading static file from directory")
	}

	if _, err = f.Write(bs); err != nil {
		return observability.PrepareError(err, logger, span, "loading static file into memory")
	}

	if err = f.Close(); err != nil {
		return observability.PrepareError(err, logger, span, "closing file while setting up static dir")
	}

	return nil
}

func (s *service) buildStaticFileServer(ctx context.Context, fileDir string) (*afero.HttpFs, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue("directory", fileDir)

	var afs afero.Fs
	if s.config.CacheStaticFiles {
		afs = afero.NewMemMapFs()

		files, err := os.ReadDir(fileDir)
		if err != nil {
			return nil, observability.PrepareError(err, logger, span, "reading directory for frontend files")
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			path := file.Name()
			logger = logger.WithValue("path", path)

			if err = s.cacheFile(ctx, afs, filepath.Join(fileDir, path)); err != nil {
				return nil, observability.PrepareError(err, logger, span, "caching file while setting up static dir")
			}
		}

		logger.Debug("returning read-only static file server")
		afs = afero.NewReadOnlyFs(afs)
	} else {
		logger.Debug("returning standard static file server")
		afs = afero.NewOsFs()
	}

	return afero.NewHttpFs(afs), nil
}

// StaticDir builds a static directory handler.
func (s *service) StaticDir(ctx context.Context, staticFilesDirectory string) (http.HandlerFunc, error) {
	rootCtx, rootSpan := s.tracer.StartSpan(ctx)
	defer rootSpan.End()

	logger := s.logger.WithValue("directory", staticFilesDirectory)

	fileDir, err := filepath.Abs(staticFilesDirectory)
	if err != nil {
		return nil, observability.PrepareError(err, logger, rootSpan, "determining absolute path of static files directory")
	}

	logger = logger.WithValue("static_dir", fileDir)

	httpFs, err := s.buildStaticFileServer(rootCtx, fileDir)
	if err != nil {
		return nil, fmt.Errorf("establishing static server filesystem: %w", err)
	}

	logger.Debug("setting static file server")
	fs := http.StripPrefix("/", http.FileServer(httpFs.Dir(fileDir)))

	f := func(res http.ResponseWriter, req *http.Request) {
		rl := s.logger.WithRequest(req)

		if s.config.LogStaticFiles {
			rl.Debug("static file requested")
		}

		if shouldRedirectToHome(req.URL.Path) {
			rl.Debug("redirecting to home page")
			req.URL.Path = "/"
		}

		w := &responseWrapper{ResponseWriter: res}

		fs.ServeHTTP(w, req)

		if w.status == http.StatusNotFound {
			// we could redirect to a common 404 page here
			rl.Debug("resource was NOT found")
		}
	}

	return f, nil
}

type responseWrapper struct {
	http.ResponseWriter
	status int
	length int
}

func (w *responseWrapper) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseWrapper) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	n, err := w.ResponseWriter.Write(b)
	w.length += n
	return n, err
}
