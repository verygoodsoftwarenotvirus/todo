package frontend

import (
	"context"
	// import embed for the side effect.
	_ "embed"
	"html/template"
	"net/http"
	"strconv"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

//go:embed templates/partials/auth/login.gotpl
var loginPrompt string

type loginPromptData struct {
	RedirectTo string
}

func (s *Service) buildLoginView(includeBaseTemplate bool) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		ctx, span := s.tracer.StartSpan(req.Context())
		defer span.End()

		contentData := &loginPromptData{
			RedirectTo: pluckRedirectURL(req),
		}

		if includeBaseTemplate {
			tmpl := s.renderTemplateIntoBaseTemplate(loginPrompt, nil)

			data := pageData{
				IsLoggedIn:  false,
				Title:       "Login",
				ContentData: contentData,
			}

			s.renderTemplateToResponse(ctx, tmpl, data, res)
		} else {
			tmpl := s.parseTemplate(ctx, "", loginPrompt, nil)

			s.renderTemplateToResponse(ctx, tmpl, contentData, res)
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
func (s *Service) parseFormEncodedLoginRequest(ctx context.Context, req *http.Request) (loginData *types.UserLoginInput, redirectTo string) {
	form, err := s.extractFormFromRequest(ctx, req)
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

	loginInput, redirectTo := s.parseFormEncodedLoginRequest(ctx, req)
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
			s.renderStringToResponse(loginPrompt, res)
			return
		}

		http.SetCookie(res, cookie)
		htmxRedirectTo(res, redirectTo)
	}
}

func (s *Service) handleLogoutSubmission(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)

	sessionCtxData, err := s.sessionContextDataFetcher(req)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "no session context data attached to request")
		http.Redirect(res, req, "/login", unauthorizedRedirectResponseCode)
		return
	}

	if !s.useFakeData {
		if err = s.authService.LogoutUser(ctx, sessionCtxData, req, res); err != nil {
			observability.AcknowledgeError(err, logger, span, "logging out user")
			return
		}
		htmxRedirectTo(res, "/")
	}
}

//go:embed templates/partials/auth/register.gotpl
var registrationPrompt string

func (s *Service) registrationComponent(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	tmpl := s.parseTemplate(ctx, "", registrationPrompt, nil)

	s.renderTemplateToResponse(ctx, tmpl, nil, res)
}

func (s *Service) registrationView(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	tmpl := s.renderTemplateIntoBaseTemplate(registrationPrompt, nil)
	data := pageData{
		IsLoggedIn:  false,
		Title:       "Register",
		ContentData: nil,
	}

	s.renderTemplateToResponse(ctx, tmpl, data, res)
}

//go:embed templates/partials/auth/registration_success.gotpl
var successfulRegistrationResponse string

type totpVerificationPrompt struct {
	TwoFactorQRCode template.URL
	UserID          uint64
}

// parseFormEncodedRegistrationRequest checks a request for a registration form, and returns the parsed login data if relevant.
func (s *Service) parseFormEncodedRegistrationRequest(ctx context.Context, req *http.Request) *types.UserRegistrationInput {
	form, err := s.extractFormFromRequest(ctx, req)
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

	registrationInput := s.parseFormEncodedRegistrationRequest(ctx, req)
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
			s.renderStringToResponse(registrationPrompt, res)
			return
		}
	} else {
		ucr = &types.UserCreationResponse{TwoFactorQRCode: ""}
	}

	tmpl := s.parseTemplate(ctx, "", successfulRegistrationResponse, nil)
	tmplData := &totpVerificationPrompt{
		/* #nosec G203 */
		TwoFactorQRCode: template.URL(ucr.TwoFactorQRCode),
		UserID:          ucr.CreatedUserID,
	}

	s.renderTemplateToResponse(ctx, tmpl, tmplData, res)
}

// parseFormEncodedTOTPSecretVerificationRequest checks a request for a registration form, and returns the parsed input.
func (s *Service) parseFormEncodedTOTPSecretVerificationRequest(ctx context.Context, req *http.Request) *types.TOTPSecretVerificationInput {
	form, err := s.extractFormFromRequest(ctx, req)
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

	verificationInput := s.parseFormEncodedTOTPSecretVerificationRequest(ctx, req)
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

	htmxRedirectTo(res, "/login")
	res.WriteHeader(http.StatusAccepted)
}
