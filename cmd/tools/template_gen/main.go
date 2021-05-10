package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

var defaultTemplateFuncMap = map[string]interface{}{}

func writeFile(path, out string) error {
	containingDir := filepath.Dir(path)

	if err := os.MkdirAll(containingDir, 0777); err != nil {
		return fmt.Errorf("error writing to filepath %q: %w", path, err)
	}

	if err := ioutil.WriteFile(path, []byte(out), 0644); err != nil {
		return fmt.Errorf("error writing to filepath %q: %w", path, err)
	}

	return nil
}

func main() {
	for path, cfg := range editorConfigs {
		if err := writeFile(path, buildBasicEditorTemplate(cfg)); err != nil {
			log.Fatal(err)
		}
	}

	for path, cfg := range tableConfigs {
		if err := writeFile(path, buildBasicTableTemplate(cfg)); err != nil {
			log.Fatal(err)
		}
	}

	for path, cfg := range creatorConfigs {
		if err := writeFile(path, buildBasicCreatorTemplate(cfg)); err != nil {
			log.Fatal(err)
		}
	}
}
