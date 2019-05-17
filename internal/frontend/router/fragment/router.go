// +build wasm

package router

import (
	"errors"
	"log"
	"strings"
	"syscall/js"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/frontend/html"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"
)

type (
	// ViewRenderer renders views
	ViewRenderer interface {
		Render() (*html.Div, error)
	}

	anonViewRenderer struct {
		renderFunc renderFunc
	}

	renderFunc func() (*html.Div, error)
)

func (v *anonViewRenderer) Render() (*html.Div, error) {
	return v.renderFunc()
}

// ViewRendererFunc takes an anonymous view rendering function and returns a ViewRendererr
func ViewRendererFunc(f func() (*html.Div, error)) ViewRenderer {
	return &anonViewRenderer{
		renderFunc: f,
	}
}

// ClientSideRouter manages view renderers
type ClientSideRouter struct {
	hostElement *html.Element
	routes      map[string]ViewRenderer
	logger      logging.Logger
}

// NewClientSideRouter instantiates a new ClientSideRouter
func NewClientSideRouter(logger logging.Logger, hostElement *html.Element) *ClientSideRouter {
	if hostElement == nil {
		panic("nil host element passed to client-side router")
	}
	return &ClientSideRouter{
		logger:      logger.WithName("frontend_router"),
		routes:      make(map[string]ViewRenderer),
		hostElement: hostElement,
	}
}

// RouteFunc returns a jsFunc that should be assigned to hashchange events
func (r *ClientSideRouter) RouteFunc(hostElement *html.Element) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		r.logger.Info(`

	RouteFunc called

		`)

		content, err := r.Route()
		if err != nil {
			log.Fatal(err)
		}
		hostElement.OrphanChildren()
		hostElement.AppendChild(content)
		return nil
	})
}

// Route is our main function which determines what page we're on, what that page should show, and reconciles the difference
func (r *ClientSideRouter) Route() (*html.Div, error) {
	var url = "/"
	fullHash := html.GetLocation().Hash()
	urlParts := strings.Split(fullHash, "#")
	if len(urlParts) > 1 {
		url = urlParts[1]
	}
	logger := r.logger.WithValue("url", url)
	logger.Debug("route called")

	route, ok := r.routes[url]
	if !ok {
		logger.Debug("route not found")
		return nil, errors.New("blah")
	}

	return route.Render()
}

// AddRoute adds a new route to the router
func (r *ClientSideRouter) AddRoute(path string, vr ViewRenderer) {
	r.routes[path] = vr
}
