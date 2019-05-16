package html

import (
	"net/url"
	"syscall/js"
)

var (
	globalLocation Location
)

// Location is a stand-in for a browser's `location` object
type Location interface {
	Href() string
	Protocol() string
	Host() string
	Hostname() string
	Port() uint16
	Pathname() string
	Search() url.Values
	Hash() string
	Username() string
	Password() string
	Origin() string
}

type location struct {
	jsLocation js.Value
}

func (w *location) Href() string {
	return w.jsLocation.Get("href").String()
}

func (w *location) Protocol() string {
	return w.jsLocation.Get("protocol").String()
}

func (w *location) Host() string {
	return w.jsLocation.Get("host").String()
}

func (w *location) Hostname() string {
	return w.jsLocation.Get("hostname").String()
}

func (w *location) Port() uint16 {
	return 0
}

func (w *location) Pathname() string {
	return w.jsLocation.Get("pathname").String()
}

func (w *location) Search() url.Values {
	query := w.jsLocation.Get("search").String()
	u, err := url.ParseQuery(query)
	if err != nil {
		return nil
	}
	return u
}

func (w *location) Hash() string {
	return w.jsLocation.Get("hash").String()
}

func (w *location) Username() string {
	return w.jsLocation.Get("username").String()
}

func (w *location) Password() string {
	return w.jsLocation.Get("password").String()
}

func (w *location) Origin() string {
	return w.jsLocation.Get("origin").String()
}

// GetLocation returns the js.Value for the `location` object in a browser
func GetLocation() Location {
	if globalLocation != nil {
		return globalLocation
	}

	jsw := js.Global().Get("location")
	globalLocation = &location{
		jsLocation: jsw,
	}

	return globalLocation
}

// // Listen on hash change:
// location.addEventListener('hashchange', router);
// // Listen on page load:
// location.addEventListener('load', router);
