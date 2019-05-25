package frontend

import (
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
)

// Routes returns a map of route to handlerfunc for the parent router to set
// this keeps routing logic in the frontend service and not in the server itself.
func (s *Service) Routes() map[string]http.HandlerFunc {
	return map[string]http.HandlerFunc{
		// "/login":    s.LoginPage,
		// "/register": s.RegistrationPage,
	}
}

var (
	itemsFrontendPathRegex = regexp.MustCompile(`/items/\d+`)
)

func init() {

}

// StaticDir establishes a static directory handler
func (s *Service) StaticDir(staticFilesDirectory string) (http.HandlerFunc, error) {
	fileDir, err := filepath.Abs(staticFilesDirectory)
	if err != nil {
		return nil, err
	}
	s.logger.WithValue("static_dir", fileDir).Debug("setting static file server")
	fs := http.StripPrefix("/", http.FileServer(http.Dir(fileDir)))

	return func(res http.ResponseWriter, req *http.Request) {
		logger := s.logger.WithRequest(req)
		logger.Debug("static file requested")

		switch req.URL.Path {
		// list your frontend routes here
		case "/register",
			"/login",
			"/items",
			"/items/new":
			s.logger.Debug(fmt.Sprintf("rerouting %q", req.URL.Path))
			req.URL.Path = "/"
		}
		if itemsFrontendPathRegex.MatchString(req.URL.Path) {
			req.URL.Path = "/"
		}

		fs.ServeHTTP(res, req)
		logger.WithValue("content_type", res.Header().Get("Content-type")).Debug("serving static file")
	}, nil
}
