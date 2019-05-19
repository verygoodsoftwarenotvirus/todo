// +build wasm

package main

import (
	"fmt"
	"strings"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/frontend/html"
	router "gitlab.com/verygoodsoftwarenotvirus/todo/internal/frontend/router/fragment"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1/zerolog"
)

const (
	appDivID = "app"
)

func buildFormP(title, formName string) (*html.Element, *html.Input) {
	p := html.NewElement("p")
	p.SetTextContent(fmt.Sprintf("%s: ", title))

	var input *html.Input
	if strings.ToLower(formName) == "password" {
		input = html.NewInput(html.PasswordInputType)
	} else {
		input = html.NewInput(html.TextInputType)
	}

	input.SetID(formName)
	input.SetName(formName)
	p.AppendChild(input)

	return p, input
}

type frontendApp struct {
	hostElement *html.Element
	logger      logging.Logger
}

func (a *frontendApp) setupRoutes() {
	r := router.NewClientSideRouter(a.logger, a.hostElement)

	// r.AddRoute("/", router.ViewRendererFunc(func() (*html.Div, error) {
	// 	// REPLACEME
	// 	return a.buildLoginPage(), nil
	// }))
	r.AddRoute("/register", router.ViewRendererFunc(func() (*html.Div, error) {
		return a.buildRegistrationPage(), nil
	}))
	r.AddRoute("/login", router.ViewRendererFunc(func() (*html.Div, error) {
		return a.buildLoginPage(), nil
	}))
	r.AddRoute("/accept_token", router.ViewRendererFunc(func() (*html.Div, error) {
		return a.buildLoginPage(), nil
	}))
	r.AddRoute("/items", router.ViewRendererFunc(func() (*html.Div, error) {
		return a.buildItemsPage(), nil
	}))
	r.AddRoute("/items/new", router.ViewRendererFunc(func() (*html.Div, error) {
		return a.buildItemCreationPage(), nil
	}))

	r.SetUnauthenticatedRoute("/register")
	// r.Set404Handler(router.ViewRendererFunc(func() (*html.Div, error) {
	// 	return a.buildItemCreationPage(), nil
	// }))

	if err := r.Route(); err != nil {
		a.logger.Fatal(err)
	}
}

func main() {
	logger := zerolog.NewZeroLogger()
	// logger.SetLevel(logging.InfoLevel)

	body := html.Body()
	appDiv := html.NewDiv()
	appDiv.SetID(appDivID)

	body.AppendChild(appDiv)
	appElement := &appDiv.Element

	a := frontendApp{
		logger:      logger,
		hostElement: appElement,
	}
	a.setupRoutes()

	// suspend loop
	for {
		select {
		case <-time.NewTicker(time.Second / 60).C:
			//
		}
	}
}
