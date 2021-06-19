package frontend

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/mock"
	testutils "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_buildLoginView(T *testing.T) {
	T.Parallel()

	T.Run("with base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/whatever", nil)

		s.service.buildLoginView(true)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
	})

	T.Run("without base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/whatever", nil)

		s.service.buildLoginView(false)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
	})
}

func buildFormFromLoginRequest(input *types.UserLoginInput) url.Values {
	form := url.Values{}

	form.Set(usernameFormKey, input.Username)
	form.Set(passwordFormKey, input.Password)
	form.Set(totpTokenFormKey, input.TOTPToken)

	return form
}

func TestService_parseFormEncodedLoginRequest(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleUser := fakes.BuildFakeUser()
		expected := fakes.BuildFakeUserLoginInputFromUser(exampleUser)
		expectedRedirectTo := "/somewheres"

		form := buildFormFromLoginRequest(expected)
		form.Set(redirectToQueryKey, expectedRedirectTo)
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))

		actual, actualRedirectTo := s.service.parseFormEncodedLoginRequest(s.ctx, req)

		assert.Equal(t, expected, actual)
		assert.Equal(t, expectedRedirectTo, actualRedirectTo)
	})

	T.Run("with invalid request body", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		badBody := &testutils.MockReadCloser{}
		badBody.On("Read", mock.IsType([]byte{})).Return(0, errors.New("blah"))

		req := httptest.NewRequest(http.MethodPost, "/", badBody)

		actual, actualRedirectTo := s.service.parseFormEncodedLoginRequest(s.ctx, req)

		assert.Nil(t, actual)
		assert.Empty(t, actualRedirectTo)
	})

	T.Run("with invalid form", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		req := httptest.NewRequest(http.MethodPost, "/", nil)

		actual, actualRedirectTo := s.service.parseFormEncodedLoginRequest(s.ctx, req)

		assert.Nil(t, actual)
		assert.Empty(t, actualRedirectTo)
	})
}

func TestService_handleLoginSubmission(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleUser := fakes.BuildFakeUser()
		expectedCookie := &http.Cookie{
			Name:  "testing",
			Value: t.Name(),
		}
		expected := fakes.BuildFakeUserLoginInputFromUser(exampleUser)

		mockAuthService := &mocktypes.AuthService{}
		mockAuthService.On(
			"AuthenticateUser",
			testutils.ContextMatcher,
			expected,
		).Return((*types.User)(nil), expectedCookie, nil)
		s.service.authService = mockAuthService

		res := httptest.NewRecorder()
		form := buildFormFromLoginRequest(expected)
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))

		s.service.handleLoginSubmission(res, req)

		mock.AssertExpectationsForObjects(t, mockAuthService)
		assert.Equal(t, http.StatusOK, res.Code)
		assert.NotEmpty(t, res.Header().Get("Set-Cookie"))
		assert.NotEmpty(t, res.Header().Get(htmxRedirectionHeader))
	})

	T.Run("with invalid request content", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/", nil)

		s.service.handleLoginSubmission(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
		assert.Empty(t, res.Header().Get("Set-Cookie"))
	})

	T.Run("with error authenticating user", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleUser := fakes.BuildFakeUser()
		expected := fakes.BuildFakeUserLoginInputFromUser(exampleUser)

		mockAuthService := &mocktypes.AuthService{}
		mockAuthService.On(
			"AuthenticateUser",
			testutils.ContextMatcher,
			expected,
		).Return((*types.User)(nil), (*http.Cookie)(nil), errors.New("blah"))
		s.service.authService = mockAuthService

		res := httptest.NewRecorder()
		form := buildFormFromLoginRequest(expected)
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))

		s.service.handleLoginSubmission(res, req)

		mock.AssertExpectationsForObjects(t, mockAuthService)
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Empty(t, res.Header().Get("Set-Cookie"))
	})
}

func TestService_handleLogoutSubmission(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		s.service.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/whatever", nil)

		mockAuthService := &mocktypes.AuthService{}
		mockAuthService.On(
			"LogoutUser",
			testutils.ContextMatcher,
			exampleSessionContextData,
			req,
			res,
		).Return(nil)
		s.service.authService = mockAuthService

		s.service.handleLogoutSubmission(res, req)

		mock.AssertExpectationsForObjects(t, mockAuthService)

		assert.Equal(t, http.StatusOK, res.Code)
		assert.NotEmpty(t, res.Header().Get(htmxRedirectionHeader))
	})

	T.Run("with error fetching session context data", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)
		s.service.sessionContextDataFetcher = testutils.BrokenSessionContextDataFetcher

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/whatever", nil)

		s.service.handleLogoutSubmission(res, req)

		assert.Equal(t, unauthorizedRedirectResponseCode, res.Code)
	})

	T.Run("with error logging user out", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		s.service.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/whatever", nil)

		mockAuthService := &mocktypes.AuthService{}
		mockAuthService.On(
			"LogoutUser",
			testutils.ContextMatcher,
			exampleSessionContextData,
			req,
			res,
		).Return(errors.New("blah"))
		s.service.authService = mockAuthService

		s.service.handleLogoutSubmission(res, req)

		mock.AssertExpectationsForObjects(t, mockAuthService)

		assert.Equal(t, http.StatusOK, res.Code)
	})
}

func TestService_registrationComponent(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/whatever", nil)

		s.service.registrationComponent(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
	})
}

func TestService_registrationView(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/whatever", nil)

		s.service.registrationView(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
	})
}

func buildFormFromRegistrationRequest(input *types.UserRegistrationInput) url.Values {
	form := url.Values{}

	form.Set(usernameFormKey, input.Username)
	form.Set(passwordFormKey, input.Password)

	return form
}

func TestService_parseFormEncodedRegistrationRequest(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		expected := fakes.BuildFakeUserRegistrationInput()

		form := buildFormFromRegistrationRequest(expected)
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))

		actual := s.service.parseFormEncodedRegistrationRequest(s.ctx, req)

		assert.Equal(t, expected, actual)
	})

	T.Run("with invalid request body", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		badBody := &testutils.MockReadCloser{}
		badBody.On("Read", mock.IsType([]byte{})).Return(0, errors.New("blah"))

		req := httptest.NewRequest(http.MethodPost, "/", badBody)

		actual := s.service.parseFormEncodedRegistrationRequest(s.ctx, req)

		assert.Nil(t, actual)
	})

	T.Run("with invalid form", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		req := httptest.NewRequest(http.MethodPost, "/verify", nil)

		actual := s.service.parseFormEncodedRegistrationRequest(s.ctx, req)

		assert.Nil(t, actual)
	})
}

func TestService_handleRegistrationSubmission(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		expected := fakes.BuildFakeUserRegistrationInput()
		form := buildFormFromRegistrationRequest(expected)

		mockUsersService := &mocktypes.UsersService{}
		mockUsersService.On(
			"RegisterUser",
			testutils.ContextMatcher,
			expected,
		).Return(&types.UserCreationResponse{}, nil)
		s.service.usersService = mockUsersService

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
		res := httptest.NewRecorder()

		s.service.handleRegistrationSubmission(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockUsersService)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		req := httptest.NewRequest(http.MethodPost, "/", nil)
		res := httptest.NewRecorder()

		s.service.handleRegistrationSubmission(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	T.Run("with error registering user", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		expected := fakes.BuildFakeUserRegistrationInput()
		form := buildFormFromRegistrationRequest(expected)

		mockUsersService := &mocktypes.UsersService{}
		mockUsersService.On(
			"RegisterUser",
			testutils.ContextMatcher,
			expected,
		).Return((*types.UserCreationResponse)(nil), errors.New("blah"))
		s.service.usersService = mockUsersService

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
		res := httptest.NewRecorder()

		s.service.handleRegistrationSubmission(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockUsersService)
	})

	T.Run("with fake data", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)
		s.service.useFakeData = true

		expected := fakes.BuildFakeUserRegistrationInput()
		form := buildFormFromRegistrationRequest(expected)

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
		res := httptest.NewRecorder()

		s.service.handleRegistrationSubmission(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
	})
}

func buildFormFromTOTPSecretVerificationRequest(input *types.TOTPSecretVerificationInput) url.Values {
	form := url.Values{}

	form.Set(totpTokenFormKey, input.TOTPToken)
	form.Set(userIDFormKey, strconv.FormatUint(input.UserID, 10))

	return form
}

func TestService_parseFormEncodedTOTPSecretVerificationRequest(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		expected := fakes.BuildFakeTOTPSecretVerificationInput()
		form := buildFormFromTOTPSecretVerificationRequest(expected)
		req := httptest.NewRequest(http.MethodPost, "/verify", strings.NewReader(form.Encode()))

		actual := s.service.parseFormEncodedTOTPSecretVerificationRequest(s.ctx, req)

		assert.NotNil(t, actual)
	})

	T.Run("with invalid request body", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		badBody := &testutils.MockReadCloser{}
		badBody.On("Read", mock.IsType([]byte{})).Return(0, errors.New("blah"))

		req := httptest.NewRequest(http.MethodPost, "/", badBody)

		actual := s.service.parseFormEncodedTOTPSecretVerificationRequest(s.ctx, req)

		assert.Nil(t, actual)
	})

	T.Run("with invalid user ID format", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		form := url.Values{
			userIDFormKey: {"not a number lol"},
		}

		req := httptest.NewRequest(http.MethodPost, "/verify", strings.NewReader(form.Encode()))

		actual := s.service.parseFormEncodedTOTPSecretVerificationRequest(s.ctx, req)

		assert.Nil(t, actual)
	})

	T.Run("with invalid form", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		form := url.Values{
			userIDFormKey: {"0"},
		}

		req := httptest.NewRequest(http.MethodPost, "/verify", strings.NewReader(form.Encode()))

		actual := s.service.parseFormEncodedTOTPSecretVerificationRequest(s.ctx, req)

		assert.Nil(t, actual)
	})
}

func TestService_handleTOTPVerificationSubmission(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		res := httptest.NewRecorder()

		expected := fakes.BuildFakeTOTPSecretVerificationInput()
		form := buildFormFromTOTPSecretVerificationRequest(expected)
		req := httptest.NewRequest(http.MethodPost, "/verify", strings.NewReader(form.Encode()))

		mockUsersService := &mocktypes.UsersService{}
		mockUsersService.On(
			"VerifyUserTwoFactorSecret",
			testutils.ContextMatcher,
			expected,
		).Return(nil)
		s.service.usersService = mockUsersService

		s.service.handleTOTPVerificationSubmission(res, req)

		assert.Equal(t, http.StatusAccepted, res.Code)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		res := httptest.NewRecorder()

		req := httptest.NewRequest(http.MethodPost, "/verify", nil)

		s.service.handleTOTPVerificationSubmission(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	T.Run("with error writing to datastore", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		res := httptest.NewRecorder()

		expected := fakes.BuildFakeTOTPSecretVerificationInput()
		form := buildFormFromTOTPSecretVerificationRequest(expected)
		req := httptest.NewRequest(http.MethodPost, "/verify", strings.NewReader(form.Encode()))

		mockUsersService := &mocktypes.UsersService{}
		mockUsersService.On(
			"VerifyUserTwoFactorSecret",
			testutils.ContextMatcher,
			expected,
		).Return(errors.New("blah"))
		s.service.usersService = mockUsersService

		s.service.handleTOTPVerificationSubmission(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)
	})
}
