// +build wasm

package html

// <ul>
// 	<li><a href="#">Home</a></li>
// 	<li><a href="#/page1">Page 1</a></li>
// 	<li><a href="#/page2">Page 2</a></li>
// </ul>

import (
// "syscall/js"
)

// NewUnorderedList builds an UnorderedList
func NewUnorderedList(items []string) *UnorderedList {
	b := &UnorderedList{
		Element: *(NewElement("ul")),
	}

	for _, item := range items {
		b.AddToList(item)
	}

	return b
}

// UnorderedList represents a <ul> element
type UnorderedList struct {
	Element
}

// AddToList adds an item to the unordered list
func (ul *UnorderedList) AddToList(item string) {
	li := NewElement("li")
	li.SetInnerHTML(item)
	ul.Element.AppendChild(li)
}
