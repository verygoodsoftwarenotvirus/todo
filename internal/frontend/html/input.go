// +build wasm

package html

type inputType string

var (
	// ButtonInputType represents the button HTML input type
	ButtonInputType inputType = "button"
	// CheckboxInputType represents the checkbox HTML input type
	CheckboxInputType inputType = "checkbox"
	// ColorInputType represents the color HTML input type
	ColorInputType inputType = "color"
	// DateInputType represents the date HTML input type
	DateInputType inputType = "date"
	// DatetimeLocalInputType represents the datetime-local HTML input type
	DatetimeLocalInputType inputType = "datetime-local"
	// EmailInputType represents the email HTML input type
	EmailInputType inputType = "email"
	// FileInputType represents the file HTML input type
	FileInputType inputType = "file"
	// HiddenInputType represents the hidden HTML input type
	HiddenInputType inputType = "hidden"
	// ImageInputType represents the image HTML input type
	ImageInputType inputType = "image"
	// MonthInputType represents the month HTML input type
	MonthInputType inputType = "month"
	// NumberInputType represents the number HTML input type
	NumberInputType inputType = "number"
	// PasswordInputType represents the password HTML input type
	PasswordInputType inputType = "password"
	// RadioInputType represents the radio HTML input type
	RadioInputType inputType = "radio"
	// RangeInputType represents the range HTML input type
	RangeInputType inputType = "range"
	// ResetInputType represents the reset HTML input type
	ResetInputType inputType = "reset"
	// SearchInputType represents the search HTML input type
	SearchInputType inputType = "search"
	// SubmitInputType represents the submit HTML input type
	SubmitInputType inputType = "submit"
	// TelInputType represents the tel HTML input type
	TelInputType inputType = "tel"
	// TextInputType represents the text HTML input type
	TextInputType inputType = "text"
	// TimeInputType represents the time HTML input type
	TimeInputType inputType = "time"
	// URLInputType represents the url HTML input type
	URLInputType inputType = "url"
	// WeekInputType represents the week HTML input type
	WeekInputType inputType = "week"
)

// Input represents an input element
type Input struct {
	Element
}

// NewInput builds a div
func NewInput(t inputType) *Input {
	i := &Input{Element: *(NewElement("input"))}
	i.setType(t)
	return i
}

// setType sets the input type
func (i *Input) setType(t inputType) {
	i.Element.JSValue().Set("type", string(t))
}

// SetName sets the input's name
func (i *Input) SetName(name string) {
	i.Element.JSValue().Set("name", name)
}

// Value fetches the input's value
func (i *Input) Value() string {
	return i.Element.JSValue().Get("value").String()
}

// SetValue sets the input's value
func (i *Input) SetValue(value string) {
	i.Element.JSValue().Set("value", value)
}
