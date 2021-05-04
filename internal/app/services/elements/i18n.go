package elements

import (
	"embed"
	"io/fs"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed translations/*.toml
var translationsDir embed.FS

var localizer *i18n.Localizer

func getLocalizer() {
	if localizer == nil {
		bundle := i18n.NewBundle(language.English)
		bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

		translationFolderContents, err := fs.ReadDir(translationsDir, "translations")
		if err != nil {
			panic(err)
		}

		for _, f := range translationFolderContents {
			translationFileBytes, err := fs.ReadFile(translationsDir, filepath.Join("translations", f.Name()))
			if err != nil {
				panic(err)
			}

			bundle.MustParseMessageFileBytes(translationFileBytes, f.Name())
		}

		localizer = i18n.NewLocalizer(bundle, "en")
	}
}

func getSimpleLocalizedString(messageID string) string {
	return localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID:      messageID,
		DefaultMessage: nil,
		TemplateData:   nil,
		Funcs:          nil,
	})
}

func init() {
	getLocalizer()
}
