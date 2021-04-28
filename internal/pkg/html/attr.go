package html

import "strings"

// Attr is an HTML element attribute
// <a href="#"> => Attr{"href": "#"}
type Attr map[string]string

type AttrModifier func(a Attr)

// Attribute constructs a new Attr.
func Attribute(modifiers ...AttrModifier) Attr {
	a := Attr{}

	for _, modifier := range modifiers {
		modifier(a)
	}

	return a
}

// Modify modifies an existing Attr.
func (a Attr) Modify(modifiers ...AttrModifier) Attr {
	x := Attr{}

	if a != nil {
		x = a
	}

	for _, modifier := range modifiers {
		modifier(x)
	}

	return x
}

// WithValue modifies an Attr to add a key/value pair.
func WithValue(key, value string) AttrModifier {
	return func(a Attr) {
		a[key] = value
	}
}

// WithValues modifies an Attr to add multiple key/value pairs.
func WithValues(values map[string]string) AttrModifier {
	return func(a Attr) {
		for k, v := range values {
			a[k] = v
		}
	}
}

// WithClass modifies an Attr to add a class.
func WithClass(class string) AttrModifier {
	return WithValue("class", class)
}

// WithClasses is an alias for WithClass.
func WithClasses(classes ...string) AttrModifier {
	return WithClass(strings.Join(classes, " "))
}
