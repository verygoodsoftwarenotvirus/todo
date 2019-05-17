// +build wasm

package html

// Form represents a form tag
type Form struct {
	Element

	action string
}

// NewForm builds a Button
func NewForm(action string) *Form {
	f := &Form{
		Element: *(NewElement("form")),
		action:  action,
	}

	f.Element.JSValue().Set("action", action)

	return f
}
