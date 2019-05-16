// +build wasm

package html

import (
	"syscall/js"
)

// https://developer.mozilla.org/en-US/docs/Web/API/Element/classList

// ClassList is an element's classList attribute
type ClassList struct {
	this js.Value
}

// NewClassList is a constructor for a ClassList
func NewClassList(jsClassList js.Value) *ClassList {
	return &ClassList{
		this: jsClassList,
	}
}

func stringsToInterfaces(input ...string) (out []interface{}) {
	for _, s := range input {
		out = append(out, s)
	}
	return
}

func (c *ClassList) String() string {
	if c == nil {
		return "somehow classlist is nil"
	}
	return c.this.String()
}

// Add adds the specified class values.
func (c *ClassList) Add(classes ...string) {
	c.this.Call("add", stringsToInterfaces(classes...))
}

// Remove removes the specified class values
func (c *ClassList) Remove(classes ...string) {
	c.this.Call("remove", stringsToInterfaces(classes...))
}

// Item returns the class value by index in the collection
func (c *ClassList) Item(index uint) string {
	return c.this.Call("item", index).String()
}

// Toggle changes the presence of a class in the class list either adding it if it's not there or removing it if it is.
func (c *ClassList) Toggle(class string) {
	c.this.Call("toggle", class)
}

// ToggleWithForce is an adaptation around the browser Toggle function, which allows an optional second parameter to be passed.
// When force is set to true, the class will be added. When it is set to false, it will be removed.
func (c *ClassList) ToggleWithForce(class string, force bool) {
	c.this.Call("toggle", class, force)
}

// Contains checks if the specified class value exists in the classList
func (c *ClassList) Contains(class string) bool {
	return c.this.Call("contains", class).Bool()
}

// Replace replaces an existing class with a new class
func (c *ClassList) Replace(oldClass, newClass string) {
	c.this.Call("replace", oldClass, newClass)
}
