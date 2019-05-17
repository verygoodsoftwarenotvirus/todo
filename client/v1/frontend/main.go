
// +build wasm

package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/frontend/html"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/frontend/router/fragment"

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
	logger logging.Logger
}

func (a *frontendApp) setupRoutes() {
	r := router.NewClientSideRouter(a.logger, a.hostElement)
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
		container := html.NewDiv()
		container.SetTextContent("how'd you get here? :)")
		return container, nil
	}))
	_, err := r.Route()
	if err != nil {
		log.Fatal(err)
	}
}


func main() {
	logger := zerolog.NewZeroLogger()
	logger.SetLevel(logging.InfoLevel)
	logger.Info("hi there")

	body := html.Body()

	appDiv := html.NewDiv()
	appDiv.SetID(appDivID)

	body.AppendChild(appDiv)
	appElement := &appDiv.Element

	a := frontendApp{
		logger: logger,
		hostElement: appElement,
	}
	a.setupRoutes()

	logger.Info("awaiting activity")
	// suspend loop
	for {
		select {
		case <-time.NewTicker((1<<31 - 1) * time.Second).C:
			//
		}
	}
}
