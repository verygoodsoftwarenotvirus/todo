package frontend

import (
	// import embed for the side effect.
	"embed"
	"io/fs"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed translations/*.toml
var translationsDir embed.FS

func provideLocalizer() (*i18n.Localizer, error) {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	translationFolderContents, folderReadErr := fs.ReadDir(translationsDir, "translations")
	if folderReadErr != nil {
		return nil, folderReadErr
	}

	for _, f := range translationFolderContents {
		translationFileBytes, fileReadErr := fs.ReadFile(translationsDir, filepath.Join("translations", f.Name()))
		if fileReadErr != nil {
			return nil, fileReadErr
		}

		bundle.MustParseMessageFileBytes(translationFileBytes, f.Name())
	}

	return i18n.NewLocalizer(bundle, "en"), nil
}

func (s *Service) getSimpleLocalizedString(messageID string) string {
	return s.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID:      messageID,
		DefaultMessage: nil,
		TemplateData:   nil,
		Funcs:          nil,
	})
}
