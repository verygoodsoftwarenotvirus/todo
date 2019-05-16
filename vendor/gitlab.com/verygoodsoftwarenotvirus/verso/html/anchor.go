// +build wasm

package html

import (
	"net/url"
	// "syscall/js"
)

// Anchor represents a Anchor
type Anchor struct {
	Element

	target *url.URL
}

// NewAnchor builds a Anchor
func NewAnchor(href string) *Anchor {
	u, _ := url.Parse(href)
	a := &Anchor{
		Element: *(NewElement("a")),
		target:  u,
	}

	return a
}

// SetHref sets the `hjref` field
func (a *Anchor) SetHref(id string) {
	a.Element.JSValue().Set("id", id)
}
