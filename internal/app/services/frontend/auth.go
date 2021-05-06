package frontend

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const loginPrompt = `<form hx-post="/auth/submit_login" hx-ext="json-enc, ajax-header, event-header">
   <h1 class="h3 mb-3 fw-normal">Sign in</h1>
   <div class="form-floating"><input id="usernameInput" required type="text" placeholder="username" minlength=4 name="username" placeholder="username" class="form-control"><label for="usernameInput">username</label></div>
   <div class="form-floating"><input id="passwordInput" required type="password" minlength=8 name="password" placeholder="password" class="form-control"><label for="passwordInput">password</label></div>
   <div class="form-floating"><input id="totpTokenInput" required type="text" pattern="\d{6}" minlength=6 maxlength=6 name="totpToken" placeholder="123456" class="form-control"><label for="totpTokenInput">2FA Token</label></div>
   <button class="w-100 btn btn-lg btn-primary" type="submit">Sign in</button>
</form>`

func (s *Service) loginComponent(res http.ResponseWriter, _ *http.Request) {
	tmpl := parseTemplate("", loginPrompt, nil)

	if err := renderTemplateToResponse(tmpl, nil, res); err != nil {
		panic(err)
	}
}

func (s *Service) handleLoginSubmission(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	loginInput, ok := ctx.Value(types.UserLoginInputContextKey).(*types.UserLoginInput)
	if !ok || loginInput == nil {
		logger.Debug("no input found for login request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	logger.WithValue("login_input", loginInput).Info("we made it!")

	_, cookie, err := s.authService.AuthenticateUser(ctx, loginInput)
	if err != nil {
		renderStringToResponse(loginPrompt, res)
		return
	}

	req.AddCookie(cookie)
	http.Redirect(res, req, "/", http.StatusAccepted)
}
