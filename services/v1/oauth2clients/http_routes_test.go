package oauth2clients

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	mockauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/auth/mock"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding/mock"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/metrics/mock"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_randString(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		actual := randString()
		assert.NotEmpty(t, actual)
	})
}

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

func Test_fetchUserID(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		req := buildRequest(t)
		expected := uint64(123)

		// for the service.fetchUserID() call
		req = req.WithContext(
			context.WithValue(req.Context(), models.UserIDKey, expected),
		)
		s := buildTestService(t)

		actual := s.fetchUserID(req)
		assert.Equal(t, expected, actual)
	})

	T.Run("without context value present", func(t *testing.T) {
		req := buildRequest(t)

		expected := uint64(0)
		s := buildTestService(t)

		actual := s.fetchUserID(req)
		assert.Equal(t, expected, actual)
	})
}

func TestService_ListHandler(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)
		userID := uint64(1)

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2Clients",
			mock.Anything,
			mock.Anything,
			userID,
		).Return(&models.OAuth2ClientList{}, nil)
		s.database = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).Return(nil)
		s.encoderDecoder = ed

		req := buildRequest(t)
		// for the service.fetchUserID() call
		req = req.WithContext(
			context.WithValue(req.Context(), models.UserIDKey, userID),
		)
		res := httptest.NewRecorder()

		s.ListHandler()(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
	})

	T.Run("with no rows returned", func(t *testing.T) {
		s := buildTestService(t)
		userID := uint64(1)

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2Clients",
			mock.Anything,
			mock.Anything,
			userID,
		).Return((*models.OAuth2ClientList)(nil), sql.ErrNoRows)
		s.database = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).Return(nil)
		s.encoderDecoder = ed

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), models.UserIDKey, userID),
		)
		res := httptest.NewRecorder()

		s.ListHandler()(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
	})

	T.Run("with error fetching from database", func(t *testing.T) {
		s := buildTestService(t)
		userID := uint64(1)

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2Clients",
			mock.Anything,
			mock.Anything,
			userID,
		).Return((*models.OAuth2ClientList)(nil), errors.New("blah"))
		s.database = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).Return(nil)
		s.encoderDecoder = ed

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), models.UserIDKey, userID),
		)
		res := httptest.NewRecorder()

		s.ListHandler()(res, req)
		assert.Equal(t, http.StatusInternalServerError, res.Code)
	})

	T.Run("with error encoding response", func(t *testing.T) {
		s := buildTestService(t)
		userID := uint64(1)

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2Clients",
			mock.Anything,
			mock.Anything,
			userID,
		).Return(&models.OAuth2ClientList{}, nil)
		s.database = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).Return(errors.New("blah"))
		s.encoderDecoder = ed

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), models.UserIDKey, userID),
		)
		res := httptest.NewRecorder()

		s.ListHandler()(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
	})
}

func TestService_CreateHandler(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		exampleUser := &models.User{
			ID:              123,
			HashedPassword:  "hashed_pass",
			Salt:            []byte("blah"),
			TwoFactorSecret: "SUPER SECRET",
		}

		exampleInput := &models.OAuth2ClientCreationInput{
			UserLoginInput: models.UserLoginInput{
				Username:  "username",
				Password:  "password",
				TOTPToken: "123456",
			},
		}
		s := buildTestService(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUserByUsername",
			mock.Anything,
			exampleInput.Username,
		).Return(exampleUser, nil)
		mockDB.OAuth2ClientDataManager.On(
			"CreateOAuth2Client",
			mock.Anything,
			exampleInput,
		).Return(&models.OAuth2Client{}, nil)
		s.database = mockDB

		a := &mockauth.Authenticator{}
		a.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleInput.Password,
			exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
			exampleUser.Salt,
		).Return(true, nil)
		s.authenticator = a

		uc := &mockmetrics.UnitCounter{}
		uc.On("Increment", mock.Anything).Return()
		s.oauth2ClientCounter = uc

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).Return(nil)
		s.encoderDecoder = ed

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), CreationMiddlewareCtxKey, exampleInput),
		)
		req = req.WithContext(
			context.WithValue(req.Context(), models.UserIDKey, exampleUser.ID),
		)
		res := httptest.NewRecorder()

		s.CreateHandler()(res, req)
		assert.Equal(t, http.StatusCreated, res.Code)
	})

	T.Run("with missing input", func(t *testing.T) {
		s := buildTestService(t)

		req := buildRequest(t)
		res := httptest.NewRecorder()

		s.CreateHandler()(res, req)
		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	T.Run("with error getting user", func(t *testing.T) {
		exampleUser := &models.User{
			ID:              123,
			HashedPassword:  "hashed_pass",
			Salt:            []byte("blah"),
			TwoFactorSecret: "SUPER SECRET",
		}

		exampleInput := &models.OAuth2ClientCreationInput{
			UserLoginInput: models.UserLoginInput{
				Username:  "username",
				Password:  "password",
				TOTPToken: "123456",
			},
		}
		s := buildTestService(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUserByUsername",
			mock.Anything,
			exampleInput.Username,
		).Return((*models.User)(nil), errors.New("blah"))
		s.database = mockDB

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), CreationMiddlewareCtxKey, exampleInput),
		)
		req = req.WithContext(
			context.WithValue(req.Context(), models.UserIDKey, exampleUser.ID),
		)
		res := httptest.NewRecorder()

		s.CreateHandler()(res, req)
		assert.Equal(t, http.StatusInternalServerError, res.Code)
	})

	T.Run("with invalid credentials", func(t *testing.T) {
		exampleUser := &models.User{
			ID:              123,
			HashedPassword:  "hashed_pass",
			Salt:            []byte("blah"),
			TwoFactorSecret: "SUPER SECRET",
		}

		exampleInput := &models.OAuth2ClientCreationInput{
			UserLoginInput: models.UserLoginInput{
				Username:  "username",
				Password:  "password",
				TOTPToken: "123456",
			},
		}
		s := buildTestService(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUserByUsername",
			mock.Anything,
			exampleInput.Username,
		).Return(exampleUser, nil)
		mockDB.OAuth2ClientDataManager.On(
			"CreateOAuth2Client",
			mock.Anything,
			exampleInput,
		).Return(&models.OAuth2Client{}, nil)
		s.database = mockDB

		a := &mockauth.Authenticator{}
		a.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleInput.Password,
			exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
			exampleUser.Salt,
		).Return(false, nil)
		s.authenticator = a

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), CreationMiddlewareCtxKey, exampleInput),
		)
		req = req.WithContext(
			context.WithValue(req.Context(), models.UserIDKey, exampleUser.ID),
		)
		res := httptest.NewRecorder()

		s.CreateHandler()(res, req)
		assert.Equal(t, http.StatusUnauthorized, res.Code)
	})

	T.Run("with error validating password", func(t *testing.T) {
		exampleUser := &models.User{
			ID:              123,
			HashedPassword:  "hashed_pass",
			Salt:            []byte("blah"),
			TwoFactorSecret: "SUPER SECRET",
		}

		exampleInput := &models.OAuth2ClientCreationInput{
			UserLoginInput: models.UserLoginInput{
				Username:  "username",
				Password:  "password",
				TOTPToken: "123456",
			},
		}
		s := buildTestService(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUserByUsername",
			mock.Anything,
			exampleInput.Username,
		).Return(exampleUser, nil)
		mockDB.OAuth2ClientDataManager.On(
			"CreateOAuth2Client",
			mock.Anything,
			exampleInput,
		).Return(&models.OAuth2Client{}, nil)
		s.database = mockDB

		a := &mockauth.Authenticator{}
		a.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleInput.Password,
			exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
			exampleUser.Salt,
		).Return(true, errors.New("blah"))
		s.authenticator = a

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), CreationMiddlewareCtxKey, exampleInput),
		)
		req = req.WithContext(
			context.WithValue(req.Context(), models.UserIDKey, exampleUser.ID),
		)
		res := httptest.NewRecorder()

		s.CreateHandler()(res, req)
		assert.Equal(t, http.StatusInternalServerError, res.Code)
	})

	T.Run("with error creating oauth2 client", func(t *testing.T) {
		exampleUser := &models.User{
			ID:              123,
			HashedPassword:  "hashed_pass",
			Salt:            []byte("blah"),
			TwoFactorSecret: "SUPER SECRET",
		}

		exampleInput := &models.OAuth2ClientCreationInput{
			UserLoginInput: models.UserLoginInput{
				Username:  "username",
				Password:  "password",
				TOTPToken: "123456",
			},
		}
		s := buildTestService(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUserByUsername",
			mock.Anything,
			exampleInput.Username,
		).Return(exampleUser, nil)
		mockDB.OAuth2ClientDataManager.On(
			"CreateOAuth2Client",
			mock.Anything,
			exampleInput,
		).Return((*models.OAuth2Client)(nil), errors.New("blah"))
		s.database = mockDB

		a := &mockauth.Authenticator{}
		a.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleInput.Password,
			exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
			exampleUser.Salt,
		).Return(true, nil)
		s.authenticator = a

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), CreationMiddlewareCtxKey, exampleInput),
		)
		req = req.WithContext(
			context.WithValue(req.Context(), models.UserIDKey, exampleUser.ID),
		)
		res := httptest.NewRecorder()

		s.CreateHandler()(res, req)
		assert.Equal(t, http.StatusInternalServerError, res.Code)
	})

	T.Run("with error encoding response", func(t *testing.T) {
		exampleUser := &models.User{
			ID:              123,
			HashedPassword:  "hashed_pass",
			Salt:            []byte("blah"),
			TwoFactorSecret: "SUPER SECRET",
		}

		exampleInput := &models.OAuth2ClientCreationInput{
			UserLoginInput: models.UserLoginInput{
				Username:  "username",
				Password:  "password",
				TOTPToken: "123456",
			},
		}
		s := buildTestService(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUserByUsername",
			mock.Anything,
			exampleInput.Username,
		).Return(exampleUser, nil)
		mockDB.OAuth2ClientDataManager.On(
			"CreateOAuth2Client",
			mock.Anything,
			exampleInput,
		).Return(&models.OAuth2Client{}, nil)
		s.database = mockDB

		a := &mockauth.Authenticator{}
		a.On(
			"ValidateLogin",
			mock.Anything,
			exampleUser.HashedPassword,
			exampleInput.Password,
			exampleUser.TwoFactorSecret,
			exampleInput.TOTPToken,
			exampleUser.Salt,
		).Return(true, nil)
		s.authenticator = a

		uc := &mockmetrics.UnitCounter{}
		uc.On("Increment", mock.Anything).Return()
		s.oauth2ClientCounter = uc

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).Return(errors.New("blah"))
		s.encoderDecoder = ed

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), CreationMiddlewareCtxKey, exampleInput),
		)
		req = req.WithContext(
			context.WithValue(req.Context(), models.UserIDKey, exampleUser.ID),
		)
		res := httptest.NewRecorder()

		s.CreateHandler()(res, req)
		assert.Equal(t, http.StatusCreated, res.Code)
	})
}

func TestService_ReadHandler(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)
		userID := uint64(1)
		exampleOAuth2ClientID := uint64(2)

		s.urlClientIDExtractor = func(req *http.Request) uint64 {
			return exampleOAuth2ClientID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2Client",
			mock.Anything,
			exampleOAuth2ClientID,
			userID,
		).Return(&models.OAuth2Client{}, nil)
		s.database = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).Return(nil)
		s.encoderDecoder = ed

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), models.UserIDKey, userID),
		)
		res := httptest.NewRecorder()

		s.ReadHandler()(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
	})

	T.Run("with no rows found", func(t *testing.T) {
		s := buildTestService(t)
		userID := uint64(1)
		exampleOAuth2ClientID := uint64(2)

		s.urlClientIDExtractor = func(req *http.Request) uint64 {
			return exampleOAuth2ClientID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2Client",
			mock.Anything,
			exampleOAuth2ClientID,
			userID,
		).Return(&models.OAuth2Client{}, sql.ErrNoRows)
		s.database = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).Return(nil)
		s.encoderDecoder = ed

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), models.UserIDKey, userID),
		)
		res := httptest.NewRecorder()

		s.ReadHandler()(res, req)
		assert.Equal(t, http.StatusNotFound, res.Code)
	})

	T.Run("with error fetching client from database", func(t *testing.T) {
		s := buildTestService(t)
		userID := uint64(1)
		exampleOAuth2ClientID := uint64(2)

		s.urlClientIDExtractor = func(req *http.Request) uint64 {
			return exampleOAuth2ClientID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2Client",
			mock.Anything,
			exampleOAuth2ClientID,
			userID,
		).Return((*models.OAuth2Client)(nil), errors.New("blah"))
		s.database = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).Return(nil)
		s.encoderDecoder = ed

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), models.UserIDKey, userID),
		)
		res := httptest.NewRecorder()

		s.ReadHandler()(res, req)
		assert.Equal(t, http.StatusInternalServerError, res.Code)
	})

	T.Run("with error encoding response", func(t *testing.T) {
		s := buildTestService(t)
		userID := uint64(1)
		exampleOAuth2ClientID := uint64(2)

		s.urlClientIDExtractor = func(req *http.Request) uint64 {
			return exampleOAuth2ClientID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2Client",
			mock.Anything,
			exampleOAuth2ClientID,
			userID,
		).Return(&models.OAuth2Client{}, nil)
		s.database = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).Return(errors.New("blah"))
		s.encoderDecoder = ed

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), models.UserIDKey, userID),
		)
		res := httptest.NewRecorder()

		s.ReadHandler()(res, req)
		assert.Equal(t, http.StatusOK, res.Code)
	})
}

func TestService_ArchiveHandler(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)
		userID := uint64(1)
		exampleOAuth2ClientID := uint64(2)

		s.urlClientIDExtractor = func(req *http.Request) uint64 {
			return exampleOAuth2ClientID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"ArchiveOAuth2Client",
			mock.Anything,
			exampleOAuth2ClientID,
			userID,
		).Return(nil)
		s.database = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).Return(nil)
		s.encoderDecoder = ed

		uc := &mockmetrics.UnitCounter{}
		uc.On("Decrement", mock.Anything).Return()
		s.oauth2ClientCounter = uc

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), models.UserIDKey, userID),
		)
		res := httptest.NewRecorder()

		s.ArchiveHandler()(res, req)
		assert.Equal(t, http.StatusNoContent, res.Code)
	})

	T.Run("with no rows found", func(t *testing.T) {
		s := buildTestService(t)
		userID := uint64(1)
		exampleOAuth2ClientID := uint64(2)

		s.urlClientIDExtractor = func(req *http.Request) uint64 {
			return exampleOAuth2ClientID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"ArchiveOAuth2Client",
			mock.Anything,
			exampleOAuth2ClientID,
			userID,
		).Return(sql.ErrNoRows)
		s.database = mockDB

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), models.UserIDKey, userID),
		)
		res := httptest.NewRecorder()

		s.ArchiveHandler()(res, req)
		assert.Equal(t, http.StatusNotFound, res.Code)
	})

	T.Run("with error deleting record", func(t *testing.T) {
		s := buildTestService(t)
		userID := uint64(1)
		exampleOAuth2ClientID := uint64(2)

		s.urlClientIDExtractor = func(req *http.Request) uint64 {
			return exampleOAuth2ClientID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"ArchiveOAuth2Client",
			mock.Anything,
			exampleOAuth2ClientID,
			userID,
		).Return(errors.New("blah"))
		s.database = mockDB

		req := buildRequest(t)
		req = req.WithContext(
			context.WithValue(req.Context(), models.UserIDKey, userID),
		)
		res := httptest.NewRecorder()

		s.ArchiveHandler()(res, req)
		assert.Equal(t, http.StatusInternalServerError, res.Code)
	})
}
