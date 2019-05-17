// +build wasm

package html

// Image represents a Image
type Image struct {
	Element
}

// NewImage builds a Image
func NewImage(src string) *Image {
	i := &Image{
		Element: *(NewElement("img")),
	}
	i.SetSrc(src)
	return i
}

// SetSrc sets the `src` field
func (i *Image) SetSrc(src string) {
	i.Element.JSValue().Set("src", src)
}

// SetWidth sets the `width` field
func (i *Image) SetWidth(val uint) {
	i.Element.JSValue().Set("width", val)
}

// SetHeight sets the `height` field
func (i *Image) SetHeight(val uint) {
	i.Element.JSValue().Set("height", val)
}
