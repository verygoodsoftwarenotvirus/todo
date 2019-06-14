package users

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	dbclient "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/client"
	mauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/auth/v1/mock"
	mencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/v1/mock"
	mmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/metrics/v1/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	mockman "gitlab.com/verygoodsoftwarenotvirus/newsman/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func buildRequest(t *testing.T) *http.Request {
	t.Helper()

	req, err := http.NewRequest(
		http.MethodGet,
		"https://verygoodsoftwarenotvirus.ru",
		nil,
	)

	require.NotNil(t, req)
	assert.NoError(t, err)
	return req
}

func Test_randString(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		actual, err := randString()
		assert.NotEmpty(t, actual)
		assert.NoError(t, err)
	})
}

func TestService_validateCredentialChangeRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)
		expected := &models.User{
			ID:              uint64(123),
			HashedPassword:  "not really lol",
			Salt:            []byte(`nah`),
			TwoFactorSecret: "still no",
		}
		exampleTOTPToken := "123456"
		examplePassword := "password"

		req := buildRequest(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.
			On("GetUser", mock.Anything, expected.ID).
			Return(expected, nil)
		s.database = mockDB

		auth := &mauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.Anything,
			expected.HashedPassword,
			examplePassword,
			expected.TwoFactorSecret,
			exampleTOTPToken,
			expected.Salt,
		).Return(true, nil)
		s.authenticator = auth

		actual, sc := s.validateCredentialChangeRequest(
			req,
			expected.ID,
			examplePassword,
			exampleTOTPToken,
		)
		assert.Equal(t, expected, actual)
		assert.Equal(t, http.StatusOK, sc)
	})

	T.Run("with no rows found in database", func(t *testing.T) {
		s := buildTestService(t)
		expected := &models.User{
			ID:              uint64(123),
			HashedPassword:  "not really lol",
			Salt:            []byte(`nah`),
			TwoFactorSecret: "still no",
		}
		exampleTOTPToken := "123456"
		examplePassword := "password"

		req := buildRequest(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.
			On("GetUser", mock.Anything, expected.ID).
			Return((*models.User)(nil), sql.ErrNoRows)
		s.database = mockDB

		actual, sc := s.validateCredentialChangeRequest(
			req,
			expected.ID,
			examplePassword,
			exampleTOTPToken,
		)
		assert.Nil(t, actual)
		assert.Equal(t, http.StatusNotFound, sc)
	})

	T.Run("with error fetching from database", func(t *testing.T) {
		s := buildTestService(t)
		expected := &models.User{
			ID:              uint64(123),
			HashedPassword:  "not really lol",
			Salt:            []byte(`nah`),
			TwoFactorSecret: "still no",
		}
		exampleTOTPToken := "123456"
		examplePassword := "password"

		req := buildRequest(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.
			On("GetUser", mock.Anything, expected.ID).
			Return((*models.User)(nil), errors.New("blah"))
		s.database = mockDB

		actual, sc := s.validateCredentialChangeRequest(
			req,
			expected.ID,
			examplePassword,
			exampleTOTPToken,
		)
		assert.Nil(t, actual)
		assert.Equal(t, http.StatusInternalServerError, sc)
	})

	T.Run("with error validating login", func(t *testing.T) {
		s := buildTestService(t)
		expected := &models.User{
			ID:              uint64(123),
			HashedPassword:  "not really lol",
			Salt:            []byte(`nah`),
			TwoFactorSecret: "still no",
		}
		exampleTOTPToken := "123456"
		examplePassword := "password"

		req := buildRequest(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.
			On("GetUser", mock.Anything, expected.ID).
			Return(expected, nil)
		s.database = mockDB

		auth := &mauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.Anything,
			expected.HashedPassword,
			examplePassword,
			expected.TwoFactorSecret,
			exampleTOTPToken,
			expected.Salt,
		).Return(false, errors.New("blah"))
		s.authenticator = auth

		actual, sc := s.validateCredentialChangeRequest(
			req,
			expected.ID,
			examplePassword,
			exampleTOTPToken,
		)
		assert.Nil(t, actual)
		assert.Equal(t, http.StatusInternalServerError, sc)
	})

	T.Run("with invalid login", func(t *testing.T) {
		s := buildTestService(t)
		expected := &models.User{
			ID:              uint64(123),
			HashedPassword:  "not really lol",
			Salt:            []byte(`nah`),
			TwoFactorSecret: "still no",
		}
		exampleTOTPToken := "123456"
		examplePassword := "password"

		req := buildRequest(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.
			On("GetUser", mock.Anything, expected.ID).
			Return(expected, nil)
		s.database = mockDB

		auth := &mauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.Anything,
			expected.HashedPassword,
			examplePassword,
			expected.TwoFactorSecret,
			exampleTOTPToken,
			expected.Salt,
		).Return(false, nil)
		s.authenticator = auth

		actual, sc := s.validateCredentialChangeRequest(
			req,
			expected.ID,
			examplePassword,
			exampleTOTPToken,
		)
		assert.Nil(t, actual)
		assert.Equal(t, http.StatusUnauthorized, sc)
	})
}

func TestService_List(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.
			On("GetUsers", mock.Anything, mock.Anything).
			Return(&models.UserList{}, nil)
		s.database = mockDB

		ed := &mencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).
			Return(nil)
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)

		s.List(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
	})

	T.Run("with error reading from database", func(t *testing.T) {
		s := buildTestService(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.
			On("GetUsers", mock.Anything, mock.Anything).
			Return((*models.UserList)(nil), errors.New("blah"))
		s.database = mockDB

		ed := &mencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).
			Return(nil)
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)

		s.List(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)
	})

	T.Run("with error encoding response", func(t *testing.T) {
		s := buildTestService(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.
			On("GetUsers", mock.Anything, mock.Anything).
			Return(&models.UserList{}, nil)
		s.database = mockDB

		ed := &mencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).
			Return(errors.New("blah"))
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)

		s.List(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
	})

}

func TestService_Create(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)
		exampleInput := &models.UserInput{
			Username: "username",
			Password: "password",
		}
		expectedUser := &models.User{
			Username:       exampleInput.Username,
			HashedPassword: "blahblah",
		}
		auth := &mauth.Authenticator{}
		auth.On("HashPassword", mock.Anything, exampleInput.Password).
			Return(expectedUser.HashedPassword, nil)
		s.authenticator = auth

		db := database.BuildMockDatabase()
		db.UserDataManager.
			On("CreateUser", mock.Anything, exampleInput).
			Return(expectedUser, nil)
		s.database = db

		mc := &mmetrics.UnitCounter{}
		mc.On("Increment", mock.Anything)
		s.userCounter = mc

		r := &mockman.Reporter{}
		r.On("Report", mock.Anything).Return()
		s.reporter = r

		ed := &mencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).Return(nil)
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)

		req = req.WithContext(context.WithValue(req.Context(), UserCreationMiddlewareCtxKey, exampleInput))

		s.userCreationEnabled = true
		s.Create(res, req)

		assert.Equal(t, http.StatusCreated, res.Code)
	})

	T.Run("with user creation disabled", func(t *testing.T) {
		s := buildTestService(t)

		res, req := httptest.NewRecorder(), buildRequest(t)

		s.userCreationEnabled = false
		s.Create(res, req)

		assert.Equal(t, http.StatusForbidden, res.Code)
	})

	T.Run("with missing input", func(t *testing.T) {
		s := buildTestService(t)

		res, req := httptest.NewRecorder(), buildRequest(t)

		s.userCreationEnabled = true
		s.Create(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	T.Run("with error hashing password", func(t *testing.T) {
		s := buildTestService(t)
		exampleInput := &models.UserInput{
			Username: "username",
			Password: "password",
		}
		expectedUser := &models.User{
			Username:       exampleInput.Username,
			HashedPassword: "blahblah",
		}
		auth := &mauth.Authenticator{}
		auth.On("HashPassword", mock.Anything, exampleInput.Password).
			Return(expectedUser.HashedPassword, errors.New("blah"))
		s.authenticator = auth

		res, req := httptest.NewRecorder(), buildRequest(t)

		req = req.WithContext(context.WithValue(req.Context(), UserCreationMiddlewareCtxKey, exampleInput))

		s.userCreationEnabled = true
		s.Create(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)
	})

	T.Run("with error creating entry in database", func(t *testing.T) {
		s := buildTestService(t)
		exampleInput := &models.UserInput{
			Username: "username",
			Password: "password",
		}
		expectedUser := &models.User{
			Username:       exampleInput.Username,
			HashedPassword: "blahblah",
		}
		auth := &mauth.Authenticator{}
		auth.On("HashPassword", mock.Anything, exampleInput.Password).
			Return(expectedUser.HashedPassword, nil)
		s.authenticator = auth

		db := database.BuildMockDatabase()
		db.UserDataManager.
			On("CreateUser", mock.Anything, exampleInput).
			Return(expectedUser, errors.New("blah"))
		s.database = db

		res, req := httptest.NewRecorder(), buildRequest(t)

		req = req.WithContext(context.WithValue(req.Context(), UserCreationMiddlewareCtxKey, exampleInput))

		s.userCreationEnabled = true
		s.Create(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)
	})

	T.Run("with pre-existing entry in database", func(t *testing.T) {
		s := buildTestService(t)
		exampleInput := &models.UserInput{
			Username: "username",
			Password: "password",
		}
		expectedUser := &models.User{
			Username:       exampleInput.Username,
			HashedPassword: "blahblah",
		}
		auth := &mauth.Authenticator{}
		auth.On("HashPassword", mock.Anything, exampleInput.Password).
			Return(expectedUser.HashedPassword, nil)
		s.authenticator = auth

		db := database.BuildMockDatabase()
		db.UserDataManager.
			On("CreateUser", mock.Anything, exampleInput).
			Return(expectedUser, dbclient.ErrUserExists)
		s.database = db

		res, req := httptest.NewRecorder(), buildRequest(t)

		req = req.WithContext(context.WithValue(req.Context(), UserCreationMiddlewareCtxKey, exampleInput))

		s.userCreationEnabled = true
		s.Create(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	T.Run("with error encoding response", func(t *testing.T) {
		s := buildTestService(t)
		exampleInput := &models.UserInput{
			Username: "username",
			Password: "password",
		}
		expectedUser := &models.User{
			Username:       exampleInput.Username,
			HashedPassword: "blahblah",
		}
		auth := &mauth.Authenticator{}
		auth.On("HashPassword", mock.Anything, exampleInput.Password).
			Return(expectedUser.HashedPassword, nil)
		s.authenticator = auth

		db := database.BuildMockDatabase()
		db.UserDataManager.
			On("CreateUser", mock.Anything, exampleInput).
			Return(expectedUser, nil)
		s.database = db

		mc := &mmetrics.UnitCounter{}
		mc.On("Increment", mock.Anything)
		s.userCounter = mc

		r := &mockman.Reporter{}
		r.On("Report", mock.Anything).Return()
		s.reporter = r

		ed := &mencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).
			Return(errors.New("blah"))
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)

		req = req.WithContext(context.WithValue(req.Context(), UserCreationMiddlewareCtxKey, exampleInput))

		s.userCreationEnabled = true
		s.Create(res, req)

		assert.Equal(t, http.StatusCreated, res.Code)
	})
}

func TestService_Read(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, mock.Anything).
			Return(&models.User{}, nil)
		s.database = mockDB

		ed := &mencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).
			Return(nil)
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)

		s.Read(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
	})

	T.Run("with no rows found", func(t *testing.T) {
		s := buildTestService(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, mock.Anything).
			Return(&models.User{}, sql.ErrNoRows)
		s.database = mockDB

		res, req := httptest.NewRecorder(), buildRequest(t)

		s.Read(res, req)

		assert.Equal(t, http.StatusNotFound, res.Code)
	})

	T.Run("with error reading from database", func(t *testing.T) {
		s := buildTestService(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, mock.Anything).
			Return(&models.User{}, errors.New("blah"))
		s.database = mockDB

		res, req := httptest.NewRecorder(), buildRequest(t)

		s.Read(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)
	})

	T.Run("with error encoding response", func(t *testing.T) {
		s := buildTestService(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, mock.Anything).
			Return(&models.User{}, nil)
		s.database = mockDB

		ed := &mencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).
			Return(errors.New("blah"))
		s.encoderDecoder = ed

		res, req := httptest.NewRecorder(), buildRequest(t)

		s.Read(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
	})

}

func TestService_NewTOTPSecret(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)
		exampleInput := &models.TOTPSecretRefreshInput{}
		exampleUser := &models.User{
			ID:              uint64(123),
			HashedPassword:  "not really lol",
			Salt:            []byte(`nah`),
			TwoFactorSecret: "still no",
		}

		res, req := httptest.NewRecorder(), buildRequest(t)

		req = req.WithContext(
			context.WithValue(
				req.Context(), TOTPSecretRefreshMiddlewareCtxKey, exampleInput,
			))
		req = req.WithContext(
			context.WithValue(
				req.Context(), models.UserIDKey, exampleUser.ID,
			))

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).
			Return(exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUser",
			mock.Anything,
			mock.Anything, // bonus points for making this second expectation explicit
		).Return(nil)
		s.database = mockDB

		auth := &mauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
			exampleUser.Salt,
		).Return(true, nil)
		s.authenticator = auth

		ed := &mencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).
			Return(nil)
		s.encoderDecoder = ed

		s.NewTOTPSecret(res, req)

		assert.Equal(t, http.StatusAccepted, res.Code)
	})

	T.Run("without input attached to request", func(t *testing.T) {
		s := buildTestService(t)

		res, req := httptest.NewRecorder(), buildRequest(t)

		s.NewTOTPSecret(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	T.Run("with input attached but without user information", func(t *testing.T) {
		s := buildTestService(t)
		exampleInput := &models.TOTPSecretRefreshInput{}

		res, req := httptest.NewRecorder(), buildRequest(t)

		req = req.WithContext(
			context.WithValue(
				req.Context(), TOTPSecretRefreshMiddlewareCtxKey, exampleInput,
			))

		s.NewTOTPSecret(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)
	})

	T.Run("with error validating login", func(t *testing.T) {
		s := buildTestService(t)
		exampleInput := &models.TOTPSecretRefreshInput{}
		exampleUser := &models.User{
			ID:              uint64(123),
			HashedPassword:  "not really lol",
			Salt:            []byte(`nah`),
			TwoFactorSecret: "still no",
		}

		res, req := httptest.NewRecorder(), buildRequest(t)

		req = req.WithContext(
			context.WithValue(
				req.Context(), TOTPSecretRefreshMiddlewareCtxKey, exampleInput,
			))
		req = req.WithContext(
			context.WithValue(
				req.Context(), models.UserIDKey, exampleUser.ID,
			))

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).
			Return(exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUser",
			mock.Anything,
			mock.Anything, // bonus points for making this second expectation explicit
		).Return(nil)
		s.database = mockDB

		auth := &mauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
			exampleUser.Salt,
		).Return(false, errors.New("blah"))
		s.authenticator = auth

		ed := &mencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).
			Return(nil)
		s.encoderDecoder = ed

		s.NewTOTPSecret(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)
	})

	T.Run("with error updating in database", func(t *testing.T) {
		s := buildTestService(t)
		exampleInput := &models.TOTPSecretRefreshInput{}
		exampleUser := &models.User{
			ID:              uint64(123),
			HashedPassword:  "not really lol",
			Salt:            []byte(`nah`),
			TwoFactorSecret: "still no",
		}

		res, req := httptest.NewRecorder(), buildRequest(t)

		req = req.WithContext(
			context.WithValue(
				req.Context(), TOTPSecretRefreshMiddlewareCtxKey, exampleInput,
			))
		req = req.WithContext(
			context.WithValue(
				req.Context(), models.UserIDKey, exampleUser.ID,
			))

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).
			Return(exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUser",
			mock.Anything,
			mock.Anything, // bonus points for making this second expectation explicit
		).Return(errors.New("blah"))
		s.database = mockDB

		auth := &mauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
			exampleUser.Salt,
		).Return(true, nil)
		s.authenticator = auth

		ed := &mencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).
			Return(nil)
		s.encoderDecoder = ed

		s.NewTOTPSecret(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)
	})

	T.Run("with error encoding response", func(t *testing.T) {
		s := buildTestService(t)
		exampleInput := &models.TOTPSecretRefreshInput{}
		exampleUser := &models.User{
			ID:              uint64(123),
			HashedPassword:  "not really lol",
			Salt:            []byte(`nah`),
			TwoFactorSecret: "still no",
		}

		res, req := httptest.NewRecorder(), buildRequest(t)

		req = req.WithContext(
			context.WithValue(
				req.Context(), TOTPSecretRefreshMiddlewareCtxKey, exampleInput,
			))
		req = req.WithContext(
			context.WithValue(
				req.Context(), models.UserIDKey, exampleUser.ID,
			))

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).
			Return(exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUser",
			mock.Anything,
			mock.Anything, // bonus points for making this second expectation explicit
		).Return(nil)
		s.database = mockDB

		auth := &mauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
			exampleUser.Salt,
		).Return(true, nil)
		s.authenticator = auth

		ed := &mencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).
			Return(errors.New("blah"))
		s.encoderDecoder = ed

		s.NewTOTPSecret(res, req)

		assert.Equal(t, http.StatusAccepted, res.Code)
	})
}

func TestService_UpdatePassword(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)

		res, req := httptest.NewRecorder(), buildRequest(t)
		exampleUser := &models.User{
			ID:              uint64(123),
			HashedPassword:  "not really lol",
			Salt:            []byte(`nah`),
			TwoFactorSecret: "still no",
		}
		exampleInput := &models.PasswordUpdateInput{
			NewPassword:     "new_password",
			CurrentPassword: "old_password",
			TOTPToken:       "123456",
		}

		req = req.WithContext(
			context.WithValue(
				req.Context(), PasswordChangeMiddlewareCtxKey, exampleInput,
			))
		req = req.WithContext(
			context.WithValue(
				req.Context(), models.UserIDKey, exampleUser.ID,
			))

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).
			Return(exampleUser, nil)
		mockDB.UserDataManager.On("UpdateUser",
			mock.Anything,
			mock.Anything, // bonus points for making this second expectation explicit
		).Return(nil)
		s.database = mockDB

		auth := &mauth.Authenticator{}

		auth.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
			exampleUser.Salt,
		).Return(true, nil)
		auth.On(
			"HashPassword",
			mock.Anything,
			exampleInput.NewPassword,
		).Return("blah", nil)

		s.authenticator = auth

		ed := &mencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).
			Return(nil)
		s.encoderDecoder = ed

		s.UpdatePassword(res, req)

		assert.Equal(t, http.StatusAccepted, res.Code)
	})

	T.Run("without input attached to request", func(t *testing.T) {
		s := buildTestService(t)

		res, req := httptest.NewRecorder(), buildRequest(t)

		s.UpdatePassword(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	T.Run("with input but without user info", func(t *testing.T) {
		s := buildTestService(t)

		res, req := httptest.NewRecorder(), buildRequest(t)
		exampleInput := &models.PasswordUpdateInput{
			NewPassword:     "new_password",
			CurrentPassword: "old_password",
			TOTPToken:       "123456",
		}

		req = req.WithContext(
			context.WithValue(
				req.Context(), PasswordChangeMiddlewareCtxKey, exampleInput,
			))

		s.UpdatePassword(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)
	})

	T.Run("with error validating login", func(t *testing.T) {
		s := buildTestService(t)

		res, req := httptest.NewRecorder(), buildRequest(t)
		exampleUser := &models.User{
			ID:              uint64(123),
			HashedPassword:  "not really lol",
			Salt:            []byte(`nah`),
			TwoFactorSecret: "still no",
		}
		exampleInput := &models.PasswordUpdateInput{
			NewPassword:     "new_password",
			CurrentPassword: "old_password",
			TOTPToken:       "123456",
		}

		req = req.WithContext(
			context.WithValue(
				req.Context(), PasswordChangeMiddlewareCtxKey, exampleInput,
			))
		req = req.WithContext(
			context.WithValue(
				req.Context(), models.UserIDKey, exampleUser.ID,
			))

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.
			On("GetUser", mock.Anything, exampleUser.ID).
			Return(exampleUser, nil)
		mockDB.UserDataManager.
			On("UpdateUser",
				mock.Anything,
				mock.Anything, // bonus points for making this second expectation explicit
			).Return(nil)
		s.database = mockDB

		auth := &mauth.Authenticator{}

		auth.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
			exampleUser.Salt,
		).Return(false, errors.New("blah"))
		s.authenticator = auth

		s.UpdatePassword(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)
	})

	T.Run("with error hashing password", func(t *testing.T) {
		s := buildTestService(t)

		res, req := httptest.NewRecorder(), buildRequest(t)
		exampleUser := &models.User{
			ID:              uint64(123),
			HashedPassword:  "not really lol",
			Salt:            []byte(`nah`),
			TwoFactorSecret: "still no",
		}
		exampleInput := &models.PasswordUpdateInput{
			NewPassword:     "new_password",
			CurrentPassword: "old_password",
			TOTPToken:       "123456",
		}

		req = req.WithContext(
			context.WithValue(
				req.Context(), PasswordChangeMiddlewareCtxKey, exampleInput,
			))
		req = req.WithContext(
			context.WithValue(
				req.Context(), models.UserIDKey, exampleUser.ID,
			))

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.
			On("GetUser", mock.Anything, exampleUser.ID).
			Return(exampleUser, nil)
		mockDB.UserDataManager.
			On("UpdateUser",
				mock.Anything,
				mock.Anything, // bonus points for making this second expectation explicit
			).Return(nil)
		s.database = mockDB

		auth := &mauth.Authenticator{}

		auth.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
			exampleUser.Salt,
		).Return(true, nil)
		auth.On(
			"HashPassword",
			mock.Anything,
			exampleInput.NewPassword,
		).Return("blah", errors.New("blah"))

		s.authenticator = auth

		ed := &mencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).
			Return(nil)
		s.encoderDecoder = ed

		s.UpdatePassword(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)
	})

	T.Run("with error updating user", func(t *testing.T) {
		s := buildTestService(t)

		res, req := httptest.NewRecorder(), buildRequest(t)
		exampleUser := &models.User{
			ID:              uint64(123),
			HashedPassword:  "not really lol",
			Salt:            []byte(`nah`),
			TwoFactorSecret: "still no",
		}
		exampleInput := &models.PasswordUpdateInput{
			NewPassword:     "new_password",
			CurrentPassword: "old_password",
			TOTPToken:       "123456",
		}

		req = req.WithContext(
			context.WithValue(
				req.Context(), PasswordChangeMiddlewareCtxKey, exampleInput,
			))
		req = req.WithContext(
			context.WithValue(
				req.Context(), models.UserIDKey, exampleUser.ID,
			))

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.
			On("GetUser", mock.Anything, exampleUser.ID).
			Return(exampleUser, nil)
		mockDB.UserDataManager.
			On("UpdateUser",
				mock.Anything,
				mock.Anything, // bonus points for making this second expectation explicit
			).Return(errors.New("blah"))
		s.database = mockDB

		auth := &mauth.Authenticator{}

		auth.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleInput.CurrentPassword,
			exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
			exampleUser.Salt,
		).Return(true, nil)
		auth.On(
			"HashPassword",
			mock.Anything,
			exampleInput.NewPassword,
		).Return("blah", nil)

		s.authenticator = auth

		s.UpdatePassword(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)
	})

}

func TestService_Delete(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)
		expectedUserID := uint64(123)
		s.userIDFetcher = func(req *http.Request) uint64 {
			return expectedUserID
		}

		res, req := httptest.NewRecorder(), buildRequest(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On("DeleteUser", mock.Anything, expectedUserID).
			Return(nil)
		s.database = mockDB

		r := &mockman.Reporter{}
		r.On("Report", mock.Anything).Return()

		mc := &mmetrics.UnitCounter{}
		mc.On("Decrement", mock.Anything)
		s.userCounter = mc

		ed := &mencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).
			Return(nil)
		s.encoderDecoder = ed

		s.Delete(res, req)

		assert.Equal(t, http.StatusNoContent, res.Code)
	})

	T.Run("with error updating database", func(t *testing.T) {
		s := buildTestService(t)
		expectedUserID := uint64(123)
		s.userIDFetcher = func(req *http.Request) uint64 {
			return expectedUserID
		}

		res, req := httptest.NewRecorder(), buildRequest(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.
			On("DeleteUser", mock.Anything, expectedUserID).
			Return(errors.New("blah"))
		s.database = mockDB

		s.Delete(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)
	})
}

func TestService_ExportData(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)
		expectedLength := 1128
		expectedUser := &models.User{ID: 123}
		expectedData := &models.DataExport{
			User:          *expectedUser,
			Items:         []models.Item{{ID: 123}},
			OAuth2Clients: []models.OAuth2Client{{ID: 123}},
			Webhooks:      []models.Webhook{{ID: 123}},
		}

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(context.WithValue(req.Context(), models.UserKey, expectedUser))

		mockDB := database.BuildMockDatabase()
		mockDB.
			On("ExportData", mock.Anything, expectedUser).
			Return(expectedData, nil)
		s.database = mockDB

		s.ExportData(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
		assert.Len(t, res.Body.String(), expectedLength)
	})
}

func TestService_BuildExportDataHandler(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)
		expectedLength := 1128
		expectedUser := &models.User{ID: 123}
		expectedData := &models.DataExport{
			User:          *expectedUser,
			Items:         []models.Item{{ID: 123}},
			OAuth2Clients: []models.OAuth2Client{{ID: 123}},
			Webhooks:      []models.Webhook{{ID: 123}},
		}

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(context.WithValue(req.Context(), models.UserKey, expectedUser))

		mockDB := database.BuildMockDatabase()
		mockDB.
			On("ExportData", mock.Anything, expectedUser).
			Return(expectedData, nil)
		s.database = mockDB

		hf := s.BuildExportDataHandler(false)
		hf(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
		assert.Len(t, res.Body.String(), expectedLength)
	})

	T.Run("without user attached to request", func(t *testing.T) {
		s := buildTestService(t)
		expectedUser := &models.User{ID: 123}
		expectedData := &models.DataExport{
			User:          *expectedUser,
			Items:         []models.Item{{ID: 123}},
			OAuth2Clients: []models.OAuth2Client{{ID: 123}},
			Webhooks:      []models.Webhook{{ID: 123}},
		}

		res, req := httptest.NewRecorder(), buildRequest(t)

		mockDB := database.BuildMockDatabase()
		mockDB.
			On("ExportData", mock.Anything, expectedUser).
			Return(expectedData, nil)
		s.database = mockDB

		hf := s.BuildExportDataHandler(false)
		hf(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)
	})

	T.Run("with error exporting data from database", func(t *testing.T) {
		s := buildTestService(t)
		expectedUser := &models.User{ID: 123}

		res, req := httptest.NewRecorder(), buildRequest(t)
		req = req.WithContext(context.WithValue(req.Context(), models.UserKey, expectedUser))

		mockDB := database.BuildMockDatabase()
		mockDB.
			On("ExportData", mock.Anything, expectedUser).
			Return((*models.DataExport)(nil), errors.New("blah"))
		s.database = mockDB

		hf := s.BuildExportDataHandler(false)
		hf(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)
	})

}
