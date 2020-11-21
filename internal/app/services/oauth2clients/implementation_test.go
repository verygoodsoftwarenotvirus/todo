package oauth2clients

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	oauth2 "gopkg.in/oauth2.v3"
	oauth2errors "gopkg.in/oauth2.v3/errors"
)

const (
	apiURLPrefix = "/api/v1"
)

func TestService_OAuth2InternalErrorHandler(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s := buildTestService(t)
		expected := errors.New("blah")

		actual := s.OAuth2InternalErrorHandler(expected)
		assert.Equal(t, expected, actual.Error)
	})
}

func TestService_OAuth2ResponseErrorHandler(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		exampleInput := &oauth2errors.Response{}
		buildTestService(t).OAuth2ResponseErrorHandler(exampleInput)
	})
}

func TestService_AuthorizeScopeHandler(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s := buildTestService(t)

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		req := buildRequest(t)
		res := httptest.NewRecorder()

		req = req.WithContext(
			context.WithValue(req.Context(), types.OAuth2ClientKey, exampleOAuth2Client),
		)
		req.URL.Path = fmt.Sprintf("%s/%s", apiURLPrefix, exampleOAuth2Client.Scopes[0])
		actual, err := s.AuthorizeScopeHandler(res, req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Equal(t, exampleOAuth2Client.Scopes[0], actual)
	})

	T.Run("without client attached to request but with client ID attached", func(t *testing.T) {
		t.Parallel()
		s := buildTestService(t)

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			exampleOAuth2Client.ClientID,
		).Return(exampleOAuth2Client, nil)
		s.clientDataManager = mockDB

		req := buildRequest(t)
		res := httptest.NewRecorder()

		req = req.WithContext(
			context.WithValue(req.Context(), clientIDKey, exampleOAuth2Client.ClientID),
		)
		req.URL.Path = fmt.Sprintf("%s/%s", apiURLPrefix, exampleOAuth2Client.Scopes[0])
		actual, err := s.AuthorizeScopeHandler(res, req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Equal(t, exampleOAuth2Client.Scopes[0], actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("without client attached to request and now rows found fetching client info", func(t *testing.T) {
		t.Parallel()
		s := buildTestService(t)

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			exampleOAuth2Client.ClientID,
		).Return((*types.OAuth2Client)(nil), sql.ErrNoRows)
		s.clientDataManager = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeErrorResponse", mock.Anything, "no such oauth2 client", http.StatusUnauthorized)
		s.encoderDecoder = ed

		req := buildRequest(t)
		res := httptest.NewRecorder()
		req = req.WithContext(
			context.WithValue(req.Context(), clientIDKey, exampleOAuth2Client.ClientID),
		)
		actual, err := s.AuthorizeScopeHandler(res, req)

		assert.Error(t, err)
		assert.Equal(t, http.StatusUnauthorized, res.Code)
		assert.Empty(t, actual)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})

	T.Run("without client attached to request and error fetching client info", func(t *testing.T) {
		t.Parallel()
		s := buildTestService(t)

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			exampleOAuth2Client.ClientID,
		).Return((*types.OAuth2Client)(nil), errors.New("blah"))
		s.clientDataManager = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything)
		s.encoderDecoder = ed

		req := buildRequest(t)
		res := httptest.NewRecorder()
		req = req.WithContext(
			context.WithValue(req.Context(), clientIDKey, exampleOAuth2Client.ClientID),
		)
		actual, err := s.AuthorizeScopeHandler(res, req)

		assert.Error(t, err)
		assert.Equal(t, http.StatusInternalServerError, res.Code)
		assert.Empty(t, actual)

		mock.AssertExpectationsForObjects(t, mockDB, ed)
	})

	T.Run("without client attached to request", func(t *testing.T) {
		t.Parallel()
		s := buildTestService(t)

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeErrorResponse", mock.Anything, "no scope information found", http.StatusBadRequest)
		s.encoderDecoder = ed

		req, res := buildRequest(t), httptest.NewRecorder()
		actual, err := s.AuthorizeScopeHandler(res, req)

		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, res.Code)
		assert.Empty(t, actual)

		mock.AssertExpectationsForObjects(t, ed)
	})

	T.Run("with invalid scope & client ID but no client", func(t *testing.T) {
		t.Parallel()
		s := buildTestService(t)

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			exampleOAuth2Client.ClientID,
		).Return(exampleOAuth2Client, nil)
		s.clientDataManager = mockDB

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeErrorResponse", mock.Anything, "not authorized for scope", http.StatusUnauthorized)
		s.encoderDecoder = ed

		req := buildRequest(t)
		req.URL.Path = fmt.Sprintf("%s/blah", apiURLPrefix)
		res := httptest.NewRecorder()
		req = req.WithContext(
			context.WithValue(req.Context(), clientIDKey, exampleOAuth2Client.ClientID),
		)
		actual, err := s.AuthorizeScopeHandler(res, req)

		assert.Error(t, err)
		assert.Equal(t, http.StatusUnauthorized, res.Code)
		assert.Empty(t, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestService_UserAuthorizationHandler(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s := buildTestService(t)

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		expected := fmt.Sprintf("%d", exampleOAuth2Client.BelongsToUser)

		req := buildRequest(t)
		res := httptest.NewRecorder()
		req = req.WithContext(
			context.WithValue(req.Context(), types.OAuth2ClientKey, exampleOAuth2Client),
		)

		actual, err := s.UserAuthorizationHandler(res, req)
		assert.NoError(t, err)
		assert.Equal(t, actual, expected)
	})

	T.Run("without client attached to request", func(t *testing.T) {
		t.Parallel()
		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		expected := fmt.Sprintf("%d", exampleUser.ID)

		req := buildRequest(t)
		res := httptest.NewRecorder()
		req = req.WithContext(
			context.WithValue(req.Context(), types.SessionInfoKey, exampleUser.ToSessionInfo()),
		)

		actual, err := s.UserAuthorizationHandler(res, req)
		assert.NoError(t, err)
		assert.Equal(t, actual, expected)
	})

	T.Run("with no user info attached", func(t *testing.T) {
		t.Parallel()
		s := buildTestService(t)
		req := buildRequest(t)
		res := httptest.NewRecorder()

		actual, err := s.UserAuthorizationHandler(res, req)
		assert.Error(t, err)
		assert.Empty(t, actual)
	})
}

func TestService_ClientAuthorizedHandler(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s := buildTestService(t)

		exampleGrant := oauth2.AuthorizationCode
		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		stringID := fmt.Sprintf("%d", exampleOAuth2Client.ID)

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			stringID,
		).Return(exampleOAuth2Client, nil)
		s.clientDataManager = mockDB

		actual, err := s.ClientAuthorizedHandler(stringID, exampleGrant)
		assert.True(t, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with password credentials grant", func(t *testing.T) {
		t.Parallel()
		s := buildTestService(t)
		exampleGrant := oauth2.PasswordCredentials

		actual, err := s.ClientAuthorizedHandler("ID", exampleGrant)
		assert.False(t, actual)
		assert.Error(t, err)
	})

	T.Run("with error reading from clientDataManager", func(t *testing.T) {
		t.Parallel()
		s := buildTestService(t)
		exampleGrant := oauth2.AuthorizationCode
		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		stringID := fmt.Sprintf("%d", exampleOAuth2Client.ID)

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			stringID,
		).Return((*types.OAuth2Client)(nil), errors.New("blah"))
		s.clientDataManager = mockDB

		actual, err := s.ClientAuthorizedHandler(stringID, exampleGrant)
		assert.False(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with disallowed implicit", func(t *testing.T) {
		t.Parallel()
		s := buildTestService(t)

		exampleGrant := oauth2.Implicit
		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		stringID := fmt.Sprintf("%d", exampleOAuth2Client.ID)

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			stringID,
		).Return(exampleOAuth2Client, nil)
		s.clientDataManager = mockDB

		actual, err := s.ClientAuthorizedHandler(stringID, exampleGrant)
		assert.False(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestService_ClientScopeHandler(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		s := buildTestService(t)

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		stringID := fmt.Sprintf("%d", exampleOAuth2Client.ID)

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			stringID,
		).Return(exampleOAuth2Client, nil)
		s.clientDataManager = mockDB

		actual, err := s.ClientScopeHandler(stringID, exampleOAuth2Client.Scopes[0])
		assert.True(t, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error reading from clientDataManager", func(t *testing.T) {
		t.Parallel()
		s := buildTestService(t)

		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		stringID := fmt.Sprintf("%d", exampleOAuth2Client.ID)

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			stringID,
		).Return((*types.OAuth2Client)(nil), errors.New("blah"))
		s.clientDataManager = mockDB

		actual, err := s.ClientScopeHandler(stringID, exampleOAuth2Client.Scopes[0])
		assert.False(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("without valid scope", func(t *testing.T) {
		t.Parallel()
		s := buildTestService(t)

		exampleScope := "halb"
		exampleOAuth2Client := fakes.BuildFakeOAuth2Client()
		stringID := fmt.Sprintf("%d", exampleOAuth2Client.ID)

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			stringID,
		).Return(exampleOAuth2Client, nil)
		s.clientDataManager = mockDB

		actual, err := s.ClientScopeHandler(stringID, exampleScope)
		assert.False(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}
