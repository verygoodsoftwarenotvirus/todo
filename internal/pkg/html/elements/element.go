package elements

import "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/html"

type Element struct {
	name     string
	attr     html.Attr
	children []interface{}
}

func (e *Element) AddHTMLChild(x html.HTML) {
	e.children = append(e.children, x)
}

func (e *Element) ModifyAttributes(mods ...html.AttrModifier) {
	e.attr.Modify(mods...)
}

func (e *Element) HTML() html.HTML {
	return html.New(e.name, e.attr, e.children)
}
