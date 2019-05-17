// +build wasm

package html

// Anchor represents a Anchor
type Anchor struct {
	Element
}

// NewAnchor builds a Anchor
func NewAnchor(href string) *Anchor {
	a := &Anchor{
		Element: *(NewElement("a")),
	}
	a.Element.JSValue().Set("href", href)

	return a
}
