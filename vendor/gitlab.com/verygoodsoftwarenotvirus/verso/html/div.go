// +build wasm

package html

import (
// "syscall/js"
)

// Div represents a div
type Div struct {
	Element
}

// NewDiv builds a div
func NewDiv() *Div {
	return &Div{Element: *(NewElement("div"))}
}
