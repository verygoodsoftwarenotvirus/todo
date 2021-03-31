package frontend

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/spf13/afero"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
)

const (
	loginRoute        = "/auth/login"
	registrationRoute = "/auth/register"
)

var (
	// Here is where you should put route regexes that need to be ignored by the static file server.
	// For instance, if you allow someone to see an event in the frontend via a url that contains dynamic.
	// information, such as `/event/123`, you would want to put something like this below:
	// 		eventsFrontendPathRegex = regexp.MustCompile(`/event/\d+`)

	// usersAdminFrontendPathRegex matches URLs for specific user routes.
	usersAdminFrontendPathRegex = regexp.MustCompile(`/admin/users/\d+`)

	// webhooksAdminFrontendPathRegex matches URLs for specific webhook routes.
	webhooksAdminFrontendPathRegex = regexp.MustCompile(`/admin/webhooks/\d+`)

	// webhooksUserFrontendPathRegex matches URLs for specific webhook routes.
	webhooksUserFrontendPathRegex = regexp.MustCompile(`/user/webhooks/\d+`)

	// itemsFrontendPathRegex matches URLs against our frontend router's specification for specific item routes.
	itemsFrontendPathRegex = regexp.MustCompile(`/things/items/\d+`)

	// itemsAdminFrontendPathRegex matches URLs against our frontend router's specification for specific item routes.
	itemsAdminFrontendPathRegex = regexp.MustCompile(`/admin/things/items/\d+`)

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
		"/items":             {},
		"/items/new":         {},
		"/things/items":      {},
		"/things/items/new":  {},
		"/dashboard":         {},
		"/user/webhooks":     {},
		"/user/webhooks/nu":  {},
		"/user/webhooks/new": {},
		"/user/settings":     {},
	}

	redirections = map[string]string{
		"/login":    loginRoute,
		"/register": registrationRoute,
	}
)

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

		if _, routeValid := validRoutes[req.URL.Path]; routeValid {
			req.URL.Path = "/"
		}

		fs.ServeHTTP(res, req)
	}, nil
}
