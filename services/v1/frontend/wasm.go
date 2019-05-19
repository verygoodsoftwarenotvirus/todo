package frontend

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fsnotify/fsnotify"
)

const (
	rootPkg = "gitlab.com/verygoodsoftwarenotvirus/todo"
)

func (s *Service) buildWASM(clientPkg string) error {
	s.logger.Debug("rebuilding WASM")
	cmd := exec.Command(
		"go", "build", "-o", "/frontend/website.wasm",
		clientPkg,
	)
	cmd.Env = []string{
		"GOOS=js",
		"GOARCH=wasm",
		fmt.Sprintf("HOME=%s", os.Getenv("HOME")),
		fmt.Sprintf("GOPATH=%s", os.Getenv("GOPATH")),
	}
	cmd.Stderr, cmd.Stdout = os.Stderr, os.Stdout

	return cmd.Run()
}

// DebugWASMPackage watches a directory for changes and recompiles it to WASM
func (s *Service) DebugWASMPackage(clientPkg string) error {
	logger := s.logger.WithValue("directory", clientPkg)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	dir := strings.Replace(clientPkg, fmt.Sprintf("%s/", rootPkg), "./", 1)
	logger.WithValue("dir", dir).Debug("Watching directory for changes")
	if err := watcher.Add(dir); err != nil {
		return err
	}

	if err := watcher.Add("/go/src/gitlab.com/verygoodsoftwarenotvirus/todo/internal/frontend"); err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-watcher.Events:
				if err := s.buildWASM(clientPkg); err != nil {
					s.logger.Error(err, "building WASM")
				}
			}
		}
	}()
	return nil
}
