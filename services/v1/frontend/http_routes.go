package frontend

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"regexp"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

// Routes returns a map of route to HandlerFunc for the parent router to set
// this keeps routing logic in the frontend service and not in the server itself.
func (s *Service) Routes() map[string]http.HandlerFunc {
	return map[string]http.HandlerFunc{
		// "/login":    s.LoginPage,
		// "/register": s.RegistrationPage,
	}
}

func (s *Service) buildStaticFileServer(fileDir string) (*afero.HttpFs, error) {
	var afs afero.Fs
	if s.config.CacheStaticFiles {
		afs = afero.NewMemMapFs()
		files, err := ioutil.ReadDir(fileDir)
		if err != nil {
			return nil, errors.Wrap(err, "reading directory for frontend files")
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			fp := filepath.Join(fileDir, file.Name())
			f, err := afs.Create(fp)
			if err != nil {
				return nil, errors.Wrap(err, "creating static file in memory")
			}

			bs, err := ioutil.ReadFile(fp)
			if err != nil {
				return nil, errors.Wrap(err, "reading static file from directory")
			}

			if _, err = f.Write(bs); err != nil {
				return nil, errors.Wrap(err, "loading static file into memory")
			}

			if err = f.Close(); err != nil {
				s.logger.Error(err, "closing file while setting up static dir")
			}
		}
		afs = afero.NewReadOnlyFs(afs)
	} else {
		afs = afero.NewOsFs()
	}

	return afero.NewHttpFs(afs), nil
}

var (
	// Here is where you should put route regexes that need to be ignored by the static file server.
	// For instance, if you allow someone to see an event in the frontend via a URL that contains dynamic
	// information, such as `/event/123`, you would want to put something like this below:
	// 		eventsFrontendPathRegex = regexp.MustCompile(`/event/\d+`)

	// itemsFrontendPathRegex matches URLs against our frontend router's specification for specific item routes
	itemsFrontendPathRegex = regexp.MustCompile(`/items/\d+`)
)

// StaticDir builds a static directory handler
func (s *Service) StaticDir(staticFilesDirectory string) (http.HandlerFunc, error) {
	fileDir, err := filepath.Abs(staticFilesDirectory)
	if err != nil {
		return nil, errors.Wrap(err, "determining absolute path of static files directory")
	}

	httpFs, err := s.buildStaticFileServer(fileDir)
	if err != nil {
		return nil, errors.Wrap(err, "establishing static server filesystem")
	}

	s.logger.WithValue("static_dir", fileDir).Debug("setting static file server")
	fs := http.StripPrefix("/", http.FileServer(httpFs.Dir(fileDir)))

	return func(res http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		// list your frontend history routes here
		case "/register",
			"/login",
			"/items",
			"/items/new",
			"/password/new":
			s.logger.Debug(fmt.Sprintf("rerouting %q", req.URL.Path))
			req.URL.Path = "/"
		}
		if itemsFrontendPathRegex.MatchString(req.URL.Path) {
			req.URL.Path = "/"
		}

		fs.ServeHTTP(res, req)
	}, nil
}
