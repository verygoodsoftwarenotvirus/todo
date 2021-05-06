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

func (s *Service) loginView(res http.ResponseWriter, _ *http.Request) {
	tmpl := renderTemplateIntoDashboard(loginPrompt, nil)

	pageData := dashboardPageData{
		Title:       "Login",
		ContentData: nil,
	}

	if err := renderTemplateToResponse(tmpl, pageData, res); err != nil {
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

	_, cookie, err := s.authService.AuthenticateUser(ctx, loginInput)
	if err != nil {
		renderStringToResponse(loginPrompt, res)
		return
	}

	req.AddCookie(cookie)

	http.Redirect(res, req, "/", http.StatusAccepted)
}

const registrationPrompt = `<form hx-post="/auth/submit_registration" hx-ext="json-enc, ajax-header, event-header">
   <h1 class="h3 mb-3 fw-normal">Register</h1>
   <div class="form-floating"><input id="usernameInput" required type="text" placeholder="username" minlength=4 name="username" placeholder="username" class="form-control"><label for="usernameInput">username</label></div>
   <div class="form-floating"><input id="passwordInput" required type="password" minlength=8 name="password" placeholder="password" class="form-control"><label for="passwordInput">password</label></div>
   <button class="w-100 btn btn-lg btn-primary" type="submit">Register</button>
</form>`

func (s *Service) registrationComponent(res http.ResponseWriter, _ *http.Request) {
	tmpl := parseTemplate("", registrationPrompt, nil)

	if err := renderTemplateToResponse(tmpl, nil, res); err != nil {
		panic(err)
	}
}

func (s *Service) registrationView(res http.ResponseWriter, _ *http.Request) {
	tmpl := renderTemplateIntoDashboard(registrationPrompt, nil)

	pageData := dashboardPageData{
		Title:       "Register",
		ContentData: nil,
	}

	if err := renderTemplateToResponse(tmpl, pageData, res); err != nil {
		panic(err)
	}
}

const successfulRegistrationResponse = `<div>
   <img src={{ .TwoFactorQRCode }}/>
</div>`

func (s *Service) handleRegistrationSubmission(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	registrationInput, ok := ctx.Value(types.UserRegistrationInputContextKey).(*types.UserRegistrationInput)
	if !ok || registrationInput == nil {
		logger.Debug("no input found for registration request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	ucr, err := s.usersService.RegisterUser(ctx, registrationInput)
	if err != nil {
		renderStringToResponse(registrationPrompt, res)
		return
	}

	tmpl := parseTemplate("", successfulRegistrationResponse, nil)
	if err = renderTemplateToResponse(tmpl, ucr, res); err != nil {
		panic(err)
	}
}
