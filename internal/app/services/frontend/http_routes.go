package frontend

import (
	"context"
	"fmt"
	"io/ioutil"
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

	f, err := afs.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating static file in memory: %w", err)
	}

	bs, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("reading static file from directory: %w", err)
	}

	if _, err = f.Write(bs); err != nil {
		return fmt.Errorf("loading static file into memory: %w", err)
	}

	if err = f.Close(); err != nil {
		s.logger.Error(err, "closing file while setting up static dir")
	}

	return nil
}

func (s *service) buildStaticFileServer(ctx context.Context, fileDir string) (*afero.HttpFs, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	var afs afero.Fs
	if s.config.CacheStaticFiles {
		afs = afero.NewMemMapFs()

		files, err := os.ReadDir(fileDir)
		if err != nil {
			return nil, fmt.Errorf("reading directory for frontend files: %w", err)
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			path := file.Name()
			logger := s.logger.WithValue("path", path)

			if err = s.cacheFile(ctx, afs, filepath.Join(fileDir, path)); err != nil {
				return nil, observability.PrepareError(err, logger, span, "caching file while setting up static dir")
			}
		}

		s.logger.Debug("returning read-only static file server")
		afs = afero.NewReadOnlyFs(afs)
	} else {
		s.logger.Debug("returning standard static file server")
		afs = afero.NewOsFs()
	}

	return afero.NewHttpFs(afs), nil
}

// StaticDir builds a static directory handler.
func (s *service) StaticDir(_ctx context.Context, staticFilesDirectory string) (http.HandlerFunc, error) {
	rootCtx, rootSpan := s.tracer.StartSpan(_ctx)
	defer rootSpan.End()

	fileDir, err := filepath.Abs(staticFilesDirectory)
	if err != nil {
		return nil, fmt.Errorf("determining absolute path of static files directory: %w", err)
	}

	logger := s.logger.WithValue("static_dir", fileDir)

	httpFs, err := s.buildStaticFileServer(rootCtx, fileDir)
	if err != nil {
		return nil, fmt.Errorf("establishing static server filesystem: %w", err)
	}

	logger.Debug("setting static file server")
	fs := http.StripPrefix("/", http.FileServer(httpFs.Dir(fileDir)))

	return func(res http.ResponseWriter, req *http.Request) {
		rl := s.logger.WithRequest(req)

		if s.config.LogStaticFiles {
			rl.Debug("static file requested")
		}

		if shouldRedirectToHome(req.URL.Path) {
			req.URL.Path = "/"
		}

		w := &reseponseWrapper{ResponseWriter: res}

		fs.ServeHTTP(w, req)

		if w.status == http.StatusNotFound {
			// we could redirect to a common 404 page here
			logger.Debug("resource was NOT found")
		}
	}, nil
}

type reseponseWrapper struct {
	http.ResponseWriter
	status int
	length int
}

func (w *reseponseWrapper) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *reseponseWrapper) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	n, err := w.ResponseWriter.Write(b)
	w.length += n
	return n, err
}
