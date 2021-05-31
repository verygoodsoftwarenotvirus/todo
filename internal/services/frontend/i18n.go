package frontend

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed translations/*.toml
var translationsDir embed.FS

func provideLocalizer() *i18n.Localizer {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	translationFolderContents, folderReadErr := fs.ReadDir(translationsDir, "translations")
	if folderReadErr != nil {
		panic(fmt.Errorf("error reading translations folder: %w", folderReadErr))
	}

	for _, f := range translationFolderContents {
		translationFilename := filepath.Join("translations", f.Name())
		translationFileBytes, fileReadErr := fs.ReadFile(translationsDir, translationFilename)
		if fileReadErr != nil {
			panic(fmt.Errorf("error reading translation file %q: %w", translationFilename, fileReadErr))
		}

		bundle.MustParseMessageFileBytes(translationFileBytes, f.Name())
	}

	return i18n.NewLocalizer(bundle, "en")
}

func (s *service) getSimpleLocalizedString(messageID string) string {
	return s.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID:      messageID,
		DefaultMessage: nil,
		TemplateData:   nil,
		Funcs:          nil,
	})
}
