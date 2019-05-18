// +build wasm

package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	// client "gitlab.com/verygoodsoftwarenotvirus/todo/client/v1/http"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/frontend/html"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

func (a *frontendApp) buildLoginFunc(usernameInput, passwordInput, totpTokenInput *html.Input) func() {
	return func() {
		username := usernameInput.Value()
		password := passwordInput.Value()
		totpToken := totpTokenInput.Value()

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

		if res.StatusCode == http.StatusNoContent {
			html.GetWindow().Location().Replace("/#/items")
		}
	}
}

func (a *frontendApp) buildLoginPage() *html.Div {
	container := html.NewDiv()

	formDiv := html.NewDiv()
	formDiv.SetStyle("margin-top: 3rem; text-align: center;")

	usernameP, usernameInput := buildFormP("username", "username")
	passwordP, passwordInput := buildFormP("password", "password")
	tokenP, totpTokenInput := buildFormP("2FA Code", "totp_token")

	submit := html.NewInput(html.SubmitInputType)
	submit.SetValue("login")
	submit.OnClick(a.buildLoginFunc(usernameInput, passwordInput, totpTokenInput))

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
