package frontend

import (
	"html/template"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const loginPrompt = `<div class="container">
	<div class="row">
		<div class="col-3"></div>
		<div class="col-6">
			<h1 class="h3 mb-3 text-center fw-normal">Log in</h1>
			<form hx-post="/auth/submit_login" hx-target="#content" hx-ext="json-enc, ajax-header, event-header">
				<div class="form-floating"><input id="usernameInput" required type="text" placeholder="username" minlength=4 name="username" placeholder="username" class="form-control"><label for="usernameInput">username</label></div>
				<div class="form-floating"><input id="passwordInput" required type="password" minlength=8 name="password" placeholder="password" class="form-control"><label for="passwordInput">password</label></div>
				<div class="form-floating"><input id="totpTokenInput" required type="text" pattern="\d{6}" minlength=6 maxlength=6 name="totpToken" placeholder="123456" class="form-control"><label for="totpTokenInput">2FA Token</label></div>
				<input type="hidden" name="redirectTo" value="{{ .RedirectTo }}" />
				<hr />
				<button class="w-100 btn btn-lg btn-primary" type="submit">Log in</button>
			</form>
			<p class="text-center"><sub><a hx-target="#content" hx-push-url="/register" hx-get="/components/registration_prompt">Register instead</a></sub></p>
		</div>
		<div class="col-3"></div>
	</div>
</div>`

type loginPromptData struct {
	RedirectTo string
}

func (s *Service) buildLoginView(includeBaseTemplate bool) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		_, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		logger := s.logger.WithRequest(req)

		var data *loginPromptData
		if redirectTo := pluckRedirectURL(req); redirectTo != "" {
			data = &loginPromptData{RedirectTo: redirectTo}
		}

		if includeBaseTemplate {
			tmpl := s.renderTemplateIntoDashboard(loginPrompt, nil)

			pageData := dashboardPageData{
				LoggedIn:    false,
				Title:       "Login",
				ContentData: data,
			}

			if err := s.renderTemplateToResponse(tmpl, pageData, res); err != nil {
				observability.AcknowledgeError(err, logger, span, "rendering item viewer into dashboard")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			tmpl := s.parseTemplate("", loginPrompt, nil)

			if err := s.renderTemplateToResponse(tmpl, data, res); err != nil {
				observability.AcknowledgeError(err, logger, span, "rendering item viewer into dashboard")
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}
}

const (
	// usernameFormKey is the string we look for in request forms for username information.
	usernameFormKey = "username"
	// passwordFormKey is the string we look for in request forms for passwords information.
	passwordFormKey = "password"
	// totpTokenFormKey is the string we look for in request forms for TOTP token information.
	totpTokenFormKey = "totpToken"
	// userIDFormKey is the string we look for in request forms for user IDs.
	userIDFormKey = "userID"
)

// parseLoginInputFromForm checks a request for a login form, and returns the parsed login data if relevant.
func parseFormEncodedLoginRequest(req *http.Request) (loginData *types.UserLoginInput, redirectTo string) {
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, ""
	}

	form, err := url.ParseQuery(string(bodyBytes))
	if err != nil {
		return nil, ""
	}

	loginData = &types.UserLoginInput{
		Username:  form.Get(usernameFormKey),
		Password:  form.Get(passwordFormKey),
		TOTPToken: form.Get(totpTokenFormKey),
	}

	if loginData.Username != "" && loginData.Password != "" && loginData.TOTPToken != "" {
		return loginData, form.Get(redirectToQueryKey)
	}

	return nil, ""
}

func (s *Service) handleLoginSubmission(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	loginInput, redirectTo := parseFormEncodedLoginRequest(req)
	if loginInput == nil {
		logger.Debug("no input found for login request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	if redirectTo == "" {
		redirectTo = "/"
	}

	if !s.useFakeData {
		_, cookie, err := s.authService.AuthenticateUser(ctx, loginInput)
		if err != nil {
			renderStringToResponse(loginPrompt, res)
			return
		}

		http.SetCookie(res, cookie)
		res.Header().Set("HX-Redirect", redirectTo)
	}
}

const registrationPrompt = `<div class="container">
	<div class="row">
		<div class="col-3"></div>
		<div class="col-6">
			<h1 class="h3 mb-3 text-center fw-normal">Register</h1>
			<form hx-post="/auth/submit_registration" hx-ext="json-enc, ajax-header, event-header">
				<div class="form-floating"><input id="usernameInput" required type="text" placeholder="username" minlength=4 name="username" placeholder="username" class="form-control"><label for="usernameInput">username</label></div>
				<div class="form-floating"><input id="passwordInput" required type="password" minlength=8 name="password" placeholder="password" class="form-control"><label for="passwordInput">password</label></div>
				<hr />
				<button class="w-100 btn btn-lg btn-primary" type="submit">Register</button>
			</form>
			<p class="text-center"><sub><a hx-target="#content" hx-push-url="/login" hx-get="/components/login_prompt">Login instead</a></sub></p>
		</div>
		<div class="col-3"></div>
	</div>
</div>`

func (s *Service) registrationComponent(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	tmpl := s.parseTemplate("", registrationPrompt, nil)

	if err := s.renderTemplateToResponse(tmpl, nil, res); err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering item viewer into dashboard")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Service) registrationView(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	tmpl := s.renderTemplateIntoDashboard(registrationPrompt, nil)

	pageData := dashboardPageData{
		LoggedIn:    false,
		Title:       "Register",
		ContentData: nil,
	}

	if err := s.renderTemplateToResponse(tmpl, pageData, res); err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering item viewer into dashboard")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}

const successfulRegistrationResponse = `<div class="container">
	<div class="row">
		<div class="col-3"></div>
		<div class="col-6">
			<h1 class="h3 mb-3 text-center fw-normal">Verify 2FA</h1>
			<form hx-post="/auth/verify_two_factor_secret" hx-ext="json-enc, ajax-header, event-header">
				<img src={{ .TwoFactorQRCode }}>
				<div class="form-floating"><input id="totpTokenInput" required type="text" pattern="\d{6}" minlength=6 maxlength=6 name="totpToken" placeholder="123456" class="form-control"><label for="totpTokenInput">2FA Token</label></div>
				<input type="hidden" name="userID" value="{{ .UserID }}" />
				<hr />	
				<button class="w-100 btn btn-lg btn-primary" type="submit">Verify</button>
			</form>
		</div>
		<div class="col-3"></div>
	</div>
</div>`

type totpVerificationPrompt struct {
	TwoFactorQRCode template.URL
	UserID          uint64
}

// parseFormEncodedRegistrationRequest checks a request for a registration form, and returns the parsed login data if relevant.
func parseFormEncodedRegistrationRequest(req *http.Request) *types.UserRegistrationInput {
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return nil
	}

	form, err := url.ParseQuery(string(bodyBytes))
	if err != nil {
		return nil
	}

	input := &types.UserRegistrationInput{
		Username: form.Get(usernameFormKey),
		Password: form.Get(passwordFormKey),
	}

	if input.Username != "" && input.Password != "" {
		return input
	}

	return nil
}

func (s *Service) handleRegistrationSubmission(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	registrationInput := parseFormEncodedRegistrationRequest(req)
	if registrationInput == nil {
		logger.Debug("no input found for registration request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	var ucr *types.UserCreationResponse
	if !s.useFakeData {
		var err error
		ucr, err = s.usersService.RegisterUser(ctx, registrationInput)
		if err != nil {
			// return erroneous markup here
			renderStringToResponse(registrationPrompt, res)
			return
		}
	} else {
		ucr = &types.UserCreationResponse{TwoFactorQRCode: ""}
	}

	tmpl := s.parseTemplate("", successfulRegistrationResponse, nil)
	tmplData := &totpVerificationPrompt{
		/* #nosec G203 */
		TwoFactorQRCode: template.URL(ucr.TwoFactorQRCode),
		UserID:          ucr.CreatedUserID,
	}

	if err := s.renderTemplateToResponse(tmpl, tmplData, res); err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering item viewer into dashboard")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// parseFormEncodedTOTPSecretVerificationRequest checks a request for a registration form, and returns the parsed input.
func parseFormEncodedTOTPSecretVerificationRequest(req *http.Request) *types.TOTPSecretVerificationInput {
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return nil
	}

	form, err := url.ParseQuery(string(bodyBytes))
	if err != nil {
		return nil
	}

	userID, err := strconv.ParseUint(form.Get(userIDFormKey), 10, 64)
	if err != nil {
		return nil
	}

	input := &types.TOTPSecretVerificationInput{
		UserID:    userID,
		TOTPToken: form.Get(totpTokenFormKey),
	}

	if input.TOTPToken != "" && input.UserID != 0 {
		return input
	}

	return nil
}

func (s *Service) handleTOTPVerificationSubmission(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	verificationInput := parseFormEncodedTOTPSecretVerificationRequest(req)
	if verificationInput == nil {
		logger.Debug("no input found for registration request")
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := s.usersService.VerifyUserTwoFactorSecret(ctx, verificationInput); err != nil {
		observability.AcknowledgeError(err, logger, span, "rendering item viewer into dashboard")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(res, req, "/login", http.StatusAccepted)
}
