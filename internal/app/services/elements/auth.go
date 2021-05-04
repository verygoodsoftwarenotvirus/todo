package elements

import (
	"net/http"
)

const loginPrompt = `<form hx-post="/login" hx-ext="json-enc, ajax-header, event-header">
   <h1 class="h3 mb-3 fw-normal">Sign in</h1>
   <div class="form-floating"><input required placeholder="username" minlength=4 type="text" class="form-control" id="usernameInput" name="username"><label for="usernameInput"></label></div>
   <div class="form-floating"><input required type="password" minlength=8 class="form-control" id="passwordInput" name="password" placeholder="password"><label for="passwordInput"></label></div>
   <div class="form-floating"><input required id="totpTokenInput" name="totpToken" pattern="\d{6}" placeholder="123456" type="numeric" class="form-control"><label for="totpTokenInput"></label></div>
   <button class="w-100 btn btn-lg btn-primary" type="submit">Sign in</button>
</form>`

func loginComponent(res http.ResponseWriter, req *http.Request) {
	renderHTMLTemplateToResponse(loginPrompt)(res, req)
}
