
// +build wasm

package main

import (
	// "context"
	"fmt"
	"log"
	"strings"
	"time"
	"net/http"
	"encoding/json"
	"bytes"

	// client "gitlab.com/verygoodsoftwarenotvirus/todo/client/v1/http"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/frontend/html"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/frontend/router/fragment"

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

func buildLoginFunc(usernameInput,passwordInput,totpTokenInput *html.Input) func() {
	return func() {
		username := usernameInput.Value()
		password := passwordInput.Value()
		totpToken := totpTokenInput.Value()

		/////////////////////////

		loginBody, _ := json.Marshal(&models.UserLoginInput{
			Username:  username,
			Password:  password,
			TOTPToken: totpToken,
		})
		req, _ := http.NewRequest(http.MethodPost, "/users/login", bytes.NewReader(loginBody))
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatalf("error executing request: %v", err)
		}
		var cookie *http.Cookie
		cookies := res.Cookies()
		if len(cookies) > 0 {
			cookie = cookies[0]
		}

		/////////////////////////

		// cookie, err := apiClient.Login(context.Background(), username, password, totpToken)
		// if err != nil {
		// 	log.Println("error building login request: ", err)
		// }

		/////////////////////////

		if cookie == nil {
			log.Printf("invalid request")
		} else {
			log.Println("hacker voice: i'm in")
		}
	}
}

func buildLoginPage() *html.Div {
	container := html.NewDiv()

	formDiv := html.NewDiv()
	formDiv.SetStyle("margin-top: 15%; text-align: center;")

	usernameP, usernameInput := buildFormP("username", "username")
	passwordP, passwordInput := buildFormP("password", "password")
	tokenP, totpTokenInput := buildFormP("2FA Code", "totp_token")

	submit := html.NewInput(html.SubmitInputType)
	submit.SetValue("login")
	submit.OnClick(buildLoginFunc(usernameInput,passwordInput,totpTokenInput))

	registerLink := html.NewAnchor("#/register")
	registerLink.SetTextContent("register instead")
	registerLink.SetStyle("margin-left: 2rem;")

	formDiv.AppendChildren(
		usernameP,
		passwordP,
		tokenP,
		submit,
		registerLink,
	)

	container.AppendChild(formDiv)
	return container
}

func main() {
	logger := zerolog.NewZeroLogger()
	logger.Info("hi there")

	body := html.Body()
	appDiv := html.NewDiv()
	appDiv.SetID(appDivID)
	body.AppendChild(appDiv)

	// u := html.GetLocation().Href()
	r := router.NewClientSideRouter(logger, &appDiv.Element)
	r.AddRoute("/login", router.ViewRendererFunc(func() (*html.Div, error) {
		buildLoginPage()
		return appDiv, nil
	}))
	r.AddRoute("/register", router.ViewRendererFunc(func() (*html.Div, error) {
		buildLoginPage()
		return appDiv, nil
	}))
	r.Route()

	// apiClient, err := client.NewClient("", "", u, logger, nil, true)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// appDiv := buildLoginPage()

	logger.Info("awaiting activity")
	// suspend loop
	for {
		select {
		case <-time.NewTicker((1<<31 - 1) * time.Second).C:
			//
		}
	}
}
