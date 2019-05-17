// +build wasm

package html

// Button represents a Button
type Button struct {
	Element
}

// NewButton builds a Button
func NewButton(name string) *Button {
	b := &Button{
		Element: *(NewElement("button")),
	}
	b.SetTextContent(name)
	return b
}
