package html

import (
	"fmt"
	"html"
	"strings"
)

var selfClosingTags = map[string]struct{}{
	"area":  {},
	"br":    {},
	"hr":    {},
	"image": {},
	"input": {},
	"img":   {},
	"link":  {},
	"meta":  {},
}

type HTML func() string

var (
	// EscapeString is an alias for convenience's sake.
	EscapeString = html.EscapeString
	// UnescapeString is an alias for convenience's sake.
	UnescapeString = html.UnescapeString
)

// UnsafeContent allows injection of JS or HTML from functions.
func UnsafeContent(str string) func() (string, bool) {
	return func() (string, bool) {
		return str, true
	}
}

// New is the base HTML func.
func New(tagName string, attrs ...interface{}) HTML {
	contents := []string{}
	attributes := ""

	for _, attr := range attrs {
		if attr != nil {
			switch a := attr.(type) {
			case Attr:
				attributes += getAttributes(a)
			case string:
				contents = append(contents, html.EscapeString(a))
			case []string:
				contents = append(contents, html.EscapeString(strings.Join(a, "")))
			case []HTML:
				contents = append(contents, RenderMultiple(a...))
			case HTML:
				contents = append(contents, a())
			case func() string:
				contents = append(contents, html.EscapeString(a()))
			case func() (string, bool):
				data, shouldNotEscape := a()
				if shouldNotEscape {
					contents = append(contents, data)
				} else {
					contents = append(contents, html.EscapeString(data))
				}
			case fmt.Stringer:
				contents = append(contents, html.EscapeString(a.String()))
			default:
				contents = append(contents, html.EscapeString(fmt.Sprintf("%v", a)))
			}
		}
	}

	return func() string {
		elc := html.EscapeString(tagName)
		if _, ok := selfClosingTags[elc]; ok {
			return "<" + elc + attributes + " />"
		}
		return "<" + elc + attributes + ">" + strings.Join(contents, "") + "</" + elc + ">"
	}
}

func RenderMultiple(in ...HTML) string {
	results := []string{}

	for _, v := range in {
		results = append(results, v())
	}

	return strings.Join(results, "")
}

func getAttributes(attributes Attr) string {
	results := []string{}
	for k, v := range attributes {
		results = append(results, fmt.Sprintf("%s='%s'", html.EscapeString(k), html.EscapeString(v)))
	}

	prefix := ""
	if len(results) > 0 {
		prefix = " "
	}

	return prefix + strings.Join(results, " ")
}
