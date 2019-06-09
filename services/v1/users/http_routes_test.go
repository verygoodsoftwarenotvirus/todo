package users

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	dbclient "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/client"
	mauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/auth/v1/mock"
	mencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/v1/mock"
	mmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/metrics/v1/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	mmodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/mock"

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

		mockDB := &mmodels.UserDataManager{}
		mockDB.On("GetUser", mock.Anything, expected.ID).
			Return(expected, nil)
		s.database = mockDB

		auth := &mauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.Anything,
			expected.HashedPassword,
			expected.Salt,
			examplePassword,
			expected.TwoFactorSecret,
			exampleTOTPToken,
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

		mockDB := &mmodels.UserDataManager{}
		mockDB.On("GetUser", mock.Anything, expected.ID).
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

		mockDB := &mmodels.UserDataManager{}
		mockDB.On("GetUser", mock.Anything, expected.ID).
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

		mockDB := &mmodels.UserDataManager{}
		mockDB.On("GetUser", mock.Anything, expected.ID).
			Return(expected, nil)
		s.database = mockDB

		auth := &mauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.Anything,
			expected.HashedPassword,
			expected.Salt,
			examplePassword,
			expected.TwoFactorSecret,
			exampleTOTPToken,
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

		mockDB := &mmodels.UserDataManager{}
		mockDB.On("GetUser", mock.Anything, expected.ID).
			Return(expected, nil)
		s.database = mockDB

		auth := &mauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.Anything,
			expected.HashedPassword,
			expected.Salt,
			examplePassword,
			expected.TwoFactorSecret,
			exampleTOTPToken,
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

		mockDB := &mmodels.UserDataManager{}
		mockDB.On("GetUsers", mock.Anything, mock.Anything).
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

		mockDB := &mmodels.UserDataManager{}
		mockDB.On("GetUsers", mock.Anything, mock.Anything).
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

		mockDB := &mmodels.UserDataManager{}
		mockDB.On("GetUsers", mock.Anything, mock.Anything).
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

		db := &mmodels.UserDataManager{}
		db.On("CreateUser", mock.Anything, exampleInput).
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

		s.Create(res, req)

		assert.Equal(t, http.StatusCreated, res.Code)
	})

	T.Run("with missing input", func(t *testing.T) {
		s := buildTestService(t)

		res, req := httptest.NewRecorder(), buildRequest(t)

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

		db := &mmodels.UserDataManager{}
		db.On("CreateUser", mock.Anything, exampleInput).
			Return(expectedUser, errors.New("blah"))
		s.database = db

		res, req := httptest.NewRecorder(), buildRequest(t)

		req = req.WithContext(context.WithValue(req.Context(), UserCreationMiddlewareCtxKey, exampleInput))

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

		db := &mmodels.UserDataManager{}
		db.On("CreateUser", mock.Anything, exampleInput).
			Return(expectedUser, dbclient.ErrUserExists)
		s.database = db

		res, req := httptest.NewRecorder(), buildRequest(t)

		req = req.WithContext(context.WithValue(req.Context(), UserCreationMiddlewareCtxKey, exampleInput))

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

		db := &mmodels.UserDataManager{}
		db.On("CreateUser", mock.Anything, exampleInput).
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

		s.Create(res, req)

		assert.Equal(t, http.StatusCreated, res.Code)
	})
}

func TestService_Read(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)

		mockDB := &mmodels.UserDataManager{}
		mockDB.On("GetUser", mock.Anything, mock.Anything).
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

		mockDB := &mmodels.UserDataManager{}
		mockDB.On("GetUser", mock.Anything, mock.Anything).
			Return(&models.User{}, sql.ErrNoRows)
		s.database = mockDB

		res, req := httptest.NewRecorder(), buildRequest(t)

		s.Read(res, req)

		assert.Equal(t, http.StatusNotFound, res.Code)
	})

	T.Run("with error reading from database", func(t *testing.T) {
		s := buildTestService(t)

		mockDB := &mmodels.UserDataManager{}
		mockDB.On("GetUser", mock.Anything, mock.Anything).
			Return(&models.User{}, errors.New("blah"))
		s.database = mockDB

		res, req := httptest.NewRecorder(), buildRequest(t)

		s.Read(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)
	})

	T.Run("with error encoding response", func(t *testing.T) {
		s := buildTestService(t)

		mockDB := &mmodels.UserDataManager{}
		mockDB.On("GetUser", mock.Anything, mock.Anything).
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

		mockDB := &mmodels.UserDataManager{}
		mockDB.On("GetUser", mock.Anything, exampleUser.ID).
			Return(exampleUser, nil)
		mockDB.On("UpdateUser",
			mock.Anything,
			mock.Anything, // bonus points for making this second expectation explicit
		).Return(nil)
		s.database = mockDB

		auth := &mauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleUser.Salt,
			exampleInput.CurrentPassword,
			exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
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

		mockDB := &mmodels.UserDataManager{}
		mockDB.On("GetUser", mock.Anything, exampleUser.ID).
			Return(exampleUser, nil)
		mockDB.On("UpdateUser",
			mock.Anything,
			mock.Anything, // bonus points for making this second expectation explicit
		).Return(nil)
		s.database = mockDB

		auth := &mauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleUser.Salt,
			exampleInput.CurrentPassword,
			exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
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

		mockDB := &mmodels.UserDataManager{}
		mockDB.On("GetUser", mock.Anything, exampleUser.ID).
			Return(exampleUser, nil)
		mockDB.On("UpdateUser",
			mock.Anything,
			mock.Anything, // bonus points for making this second expectation explicit
		).Return(errors.New("blah"))
		s.database = mockDB

		auth := &mauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleUser.Salt,
			exampleInput.CurrentPassword,
			exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
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

		mockDB := &mmodels.UserDataManager{}
		mockDB.On("GetUser", mock.Anything, exampleUser.ID).
			Return(exampleUser, nil)
		mockDB.On("UpdateUser",
			mock.Anything,
			mock.Anything, // bonus points for making this second expectation explicit
		).Return(nil)
		s.database = mockDB

		auth := &mauth.Authenticator{}
		auth.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleUser.Salt,
			exampleInput.CurrentPassword,
			exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
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

		mockDB := &mmodels.UserDataManager{}
		mockDB.On("GetUser", mock.Anything, exampleUser.ID).
			Return(exampleUser, nil)
		mockDB.On("UpdateUser",
			mock.Anything,
			mock.Anything, // bonus points for making this second expectation explicit
		).Return(nil)
		s.database = mockDB

		auth := &mauth.Authenticator{}

		auth.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleUser.Salt,
			exampleInput.CurrentPassword,
			exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
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

		mockDB := &mmodels.UserDataManager{}
		mockDB.On("GetUser", mock.Anything, exampleUser.ID).
			Return(exampleUser, nil)
		mockDB.On("UpdateUser",
			mock.Anything,
			mock.Anything, // bonus points for making this second expectation explicit
		).Return(nil)
		s.database = mockDB

		auth := &mauth.Authenticator{}

		auth.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleUser.Salt,
			exampleInput.CurrentPassword,
			exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
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

		mockDB := &mmodels.UserDataManager{}
		mockDB.On("GetUser", mock.Anything, exampleUser.ID).
			Return(exampleUser, nil)
		mockDB.On("UpdateUser",
			mock.Anything,
			mock.Anything, // bonus points for making this second expectation explicit
		).Return(nil)
		s.database = mockDB

		auth := &mauth.Authenticator{}

		auth.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleUser.Salt,
			exampleInput.CurrentPassword,
			exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
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

		mockDB := &mmodels.UserDataManager{}
		mockDB.On("GetUser", mock.Anything, exampleUser.ID).
			Return(exampleUser, nil)
		mockDB.On("UpdateUser",
			mock.Anything,
			mock.Anything, // bonus points for making this second expectation explicit
		).Return(errors.New("blah"))
		s.database = mockDB

		auth := &mauth.Authenticator{}

		auth.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleUser.Salt,
			exampleInput.CurrentPassword,
			exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
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

		mockDB := &mmodels.UserDataManager{}
		mockDB.On("DeleteUser", mock.Anything, expectedUserID).
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

		mockDB := &mmodels.UserDataManager{}
		mockDB.On("DeleteUser", mock.Anything, expectedUserID).
			Return(errors.New("blah"))
		s.database = mockDB

		s.Delete(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)
	})
}
