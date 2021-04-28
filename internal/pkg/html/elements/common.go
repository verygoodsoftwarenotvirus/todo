package elements

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/html"
)

func NewLabel(label, labelFor string) *Element {
	attr := html.Attr{}

	if labelFor != "" {
		attr.Modify(html.WithValue("for", labelFor))
	}

	return &Element{
		name:     "label",
		attr:     attr,
		children: []interface{}{},
	}
}

func NewTitle(title string) *Element {
	return &Element{
		name: "title",
		attr: html.Attr{},
		children: []interface{}{
			title,
		},
	}
}

func NewHead(children ...html.HTML) *Element {
	return &Element{
		name:     "head",
		attr:     nil,
		children: htmlArrayToInterfaces(children),
	}
}
