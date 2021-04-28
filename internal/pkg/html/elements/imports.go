package elements

import "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/html"

// StylesheetImport builds a new stylesheet import.
func StylesheetImport(href, integrity string) html.HTML {
	attr := html.Attr{
		"href":        href,
		"rel":         "stylesheet",
		"crossorigin": "anonymous",
	}

	if integrity != "" {
		attr["integrity"] = integrity
	}

	return html.New("link", attr)
}

// JavascriptImport builds a new script import.
func JavascriptImport(src, integrity string, module bool) html.HTML {
	attr := html.Attr{
		"src": src,
	}

	if module {
		attr["type"] = "module"
	}

	if integrity != "" {
		attr["integrity"] = integrity
	}

	return html.New("script", attr)
}
