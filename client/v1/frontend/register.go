// build +wasm

package main

import (
	"bytes"
	"encoding/json"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/frontend/html"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/pkg/errors"
)

func (a *frontendApp) buildRegistrationPage() *html.Div {
	container := html.NewDiv()

	formDiv := html.NewDiv()
	formDiv.SetStyle("margin-top: 15%; text-align: center;")

	usernameP, usernameInput := buildFormP("username", "username")
	passwordP, passwordInput := buildFormP("password", "password")

	registerLink := html.NewAnchor("/#/login")
	registerLink.SetTextContent("login instead")
	registerLink.SetStyle("margin-left: 2rem;")

	submit := html.NewInput(html.SubmitInputType)
	submit.SetValue("register")
	submit.OnClick(func() {
		username := usernameInput.Value()
		password := passwordInput.Value()

		/////////////////////////

		loginBody, _ := json.Marshal(&models.UserLoginInput{
			Username: username,
			Password: password,
		})
		req, _ := http.NewRequest(http.MethodPost, "/users/", bytes.NewReader(loginBody))
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			a.logger.Fatal(errors.Wrap(err, "executing request"))
		}
		var ucr models.UserCreationResponse
		json.NewDecoder(res.Body).Decode(&ucr)

		a.logger.Info(ucr.TwoFactorQRCode)

		formDiv.RemoveChildren(
			usernameP,
			passwordP,
			submit,
			registerLink,
		)

		img := html.NewImage(ucr.TwoFactorQRCode)
		img.SetStyle("height: 40%")

		disclaimer := html.NewElement("p")
		disclaimer.SetTextContent(`I will click this button only after I've saved my QR code`)

		b := html.NewButton("I saved it")
		b.OnClick(func() {
			html.GetWindow().Location().Replace("/#/login")
		})

		formDiv.AppendChildren(img, disclaimer, b)

	})

	formDiv.AppendChildren(
		usernameP,
		passwordP,
		submit,
		registerLink,
	)

	container.AppendChild(formDiv)
	return container
}
