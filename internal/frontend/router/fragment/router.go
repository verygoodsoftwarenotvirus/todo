// +build wasm

package router

import (
	"errors"
	"fmt"
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
	hostElement          *html.Element
	logger               logging.Logger
	routes               map[string]ViewRenderer
	unauthenticatedRoute string
}

// NewClientSideRouter instantiates a new ClientSideRouter
func NewClientSideRouter(logger logging.Logger, hostElement *html.Element) *ClientSideRouter {
	if hostElement == nil {
		panic("nil host element passed to client-side router")
	}
	r := &ClientSideRouter{
		logger:      logger.WithName("frontend_router"),
		routes:      make(map[string]ViewRenderer),
		hostElement: hostElement,
	}

	window := html.GetWindow()
	window.AddEventListener("hashchange", r.routeFunc(hostElement))

	return r
}

// RouteFunc returns a jsFunc that should be assigned to hashchange events
func (r *ClientSideRouter) routeFunc(hostElement *html.Element) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return r.Route()
	})
}

// Route is our main function which determines what page we're on, what that page should show, and reconciles the difference
func (r *ClientSideRouter) Route() error {
	var (
		content  *html.Div
		err      error
		url      = "/"
		loggedIn = html.GetDocument().Cookie() == ""
	)

	urlParts := strings.Split(html.GetLocation().Hash(), "#")
	if len(urlParts) > 1 {
		url = urlParts[1]
	}

	if content == nil {
		route, ok := r.routes[url]
		if !ok {
			if loggedIn && r.unauthenticatedRoute != "" && url != r.unauthenticatedRoute {
				html.GetWindow().Location().Replace(fmt.Sprintf("/#%s", r.unauthenticatedRoute))
				return nil
			}

			r.logger.WithValue("url", url).Debug("route not found!")
			return errors.New("route not found")
		}
		content, err = route.Render()
	}

	if err != nil {
		return err
	}

	r.hostElement.OrphanChildren()
	r.hostElement.AppendChild(content)

	return nil
}

// AddRoute adds a new route to the router
func (r *ClientSideRouter) AddRoute(path string, vr ViewRenderer) {
	r.routes[path] = vr
}

// SetUnauthenticatedRoute sets the default route to use in the event that a user isn't logged in
// this MUST be called AFTER you initialize the route if you want it to not return an error
func (r *ClientSideRouter) SetUnauthenticatedRoute(path string) error {
	if _, ok := r.routes[path]; !ok {
		return fmt.Errorf("invalid default route: %q", path)
	}
	r.unauthenticatedRoute = path
	return nil
}
