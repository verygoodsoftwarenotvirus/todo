package oauth2clients

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/oauth2.v3"
	oauth2errors "gopkg.in/oauth2.v3/errors"
)

const (
	apiURLPrefix = "/api/v1"
)

func TestService_OAuth2InternalErrorHandler(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)
		expected := errors.New("blah")

		actual := s.OAuth2InternalErrorHandler(expected)
		assert.Equal(t, expected, actual.Error)
	})
}

func TestService_OAuth2ResponseErrorHandler(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		exampleInput := &oauth2errors.Response{}
		buildTestService(t).OAuth2ResponseErrorHandler(exampleInput)
	})
}

func TestService_AuthorizeScopeHandler(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)
		expected := "blah"
		exampleClient := &models.OAuth2Client{
			Scopes: strings.Split(expected, ","),
		}

		req := buildRequest(t)
		res := httptest.NewRecorder()

		req = req.WithContext(
			context.WithValue(req.Context(), models.OAuth2ClientKey, exampleClient),
		)
		req.URL.Path = fmt.Sprintf("%s/blah", apiURLPrefix)
		actual, err := s.AuthorizeScopeHandler(res, req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Equal(t, expected, actual)
	})

	T.Run("without client attached to request but with client ID attached", func(t *testing.T) {
		s := buildTestService(t)
		expected := "blah"
		exampleClient := &models.OAuth2Client{
			ClientID: "blargh",
			Scopes:   strings.Split(expected, ","),
		}

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			exampleClient.ClientID,
		).Return(exampleClient, nil)
		s.database = mockDB

		req := buildRequest(t)
		res := httptest.NewRecorder()

		req = req.WithContext(
			context.WithValue(req.Context(), clientIDKey, exampleClient.ClientID),
		)
		req.URL.Path = fmt.Sprintf("%s/blah", apiURLPrefix)
		actual, err := s.AuthorizeScopeHandler(res, req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.Code)
		assert.Equal(t, expected, actual)
	})

	T.Run("without client attached to request and now rows found fetching client info", func(t *testing.T) {
		s := buildTestService(t)
		expected := "blah,flarg"
		exampleClient := &models.OAuth2Client{
			ClientID: "blargh",
			Scopes:   strings.Split(expected, ","),
		}

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			exampleClient.ClientID,
		).Return((*models.OAuth2Client)(nil), sql.ErrNoRows)
		s.database = mockDB

		req := buildRequest(t)
		res := httptest.NewRecorder()
		req = req.WithContext(
			context.WithValue(req.Context(), clientIDKey, exampleClient.ClientID),
		)
		actual, err := s.AuthorizeScopeHandler(res, req)

		assert.Error(t, err)
		assert.Equal(t, http.StatusNotFound, res.Code)
		assert.Empty(t, actual)
	})

	T.Run("without client attached to request and error fetching client info", func(t *testing.T) {
		s := buildTestService(t)
		expected := "blah,flarg"
		exampleClient := &models.OAuth2Client{
			ClientID: "blargh",
			Scopes:   strings.Split(expected, ","),
		}

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			exampleClient.ClientID,
		).Return((*models.OAuth2Client)(nil), errors.New("blah"))
		s.database = mockDB

		req := buildRequest(t)
		res := httptest.NewRecorder()
		req = req.WithContext(
			context.WithValue(req.Context(), clientIDKey, exampleClient.ClientID),
		)
		actual, err := s.AuthorizeScopeHandler(res, req)

		assert.Error(t, err)
		assert.Equal(t, http.StatusInternalServerError, res.Code)
		assert.Empty(t, actual)
	})

	T.Run("without client attached to request", func(t *testing.T) {
		s := buildTestService(t)
		req := buildRequest(t)
		res := httptest.NewRecorder()
		actual, err := s.AuthorizeScopeHandler(res, req)

		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, res.Code)
		assert.Empty(t, actual)
	})

	T.Run("with invalid scope & client ID but no client", func(t *testing.T) {
		s := buildTestService(t)
		exampleClient := &models.OAuth2Client{
			ClientID: "blargh",
			Scopes:   []string{},
		}

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			exampleClient.ClientID,
		).Return(exampleClient, nil)
		s.database = mockDB

		req := buildRequest(t)
		req.URL.Path = fmt.Sprintf("%s/blah", apiURLPrefix)
		res := httptest.NewRecorder()
		req = req.WithContext(
			context.WithValue(req.Context(), clientIDKey, exampleClient.ClientID),
		)
		actual, err := s.AuthorizeScopeHandler(res, req)

		assert.Error(t, err)
		assert.Equal(t, http.StatusUnauthorized, res.Code)
		assert.Empty(t, actual)
	})
}

func TestService_UserAuthorizationHandler(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)
		exampleClient := &models.OAuth2Client{BelongsTo: 1}
		expected := fmt.Sprintf("%d", exampleClient.BelongsTo)

		req := buildRequest(t)
		res := httptest.NewRecorder()
		req = req.WithContext(
			context.WithValue(req.Context(), models.OAuth2ClientKey, exampleClient),
		)

		actual, err := s.UserAuthorizationHandler(res, req)
		assert.NoError(t, err)
		assert.Equal(t, actual, expected)
	})

	T.Run("without client attached to request", func(t *testing.T) {
		s := buildTestService(t)
		exampleUser := &models.User{ID: 1}
		expected := fmt.Sprintf("%d", exampleUser.ID)

		req := buildRequest(t)
		res := httptest.NewRecorder()
		req = req.WithContext(
			context.WithValue(req.Context(), models.UserKey, exampleUser),
		)

		actual, err := s.UserAuthorizationHandler(res, req)
		assert.NoError(t, err)
		assert.Equal(t, actual, expected)
	})

	T.Run("with no user info attached", func(t *testing.T) {
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
		s := buildTestService(t)
		expected := true

		exampleGrant := oauth2.AuthorizationCode
		exampleClient := &models.OAuth2Client{
			ID:       1,
			ClientID: "blah",
			Scopes:   []string{},
		}
		stringID := fmt.Sprintf("%d", exampleClient.ID)

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			stringID,
		).Return(exampleClient, nil)
		s.database = mockDB

		actual, err := s.ClientAuthorizedHandler(stringID, exampleGrant)
		assert.Equal(t, expected, actual)
		assert.NoError(t, err)
	})

	T.Run("with password credentials grant", func(t *testing.T) {
		s := buildTestService(t)
		expected := false
		exampleGrant := oauth2.PasswordCredentials

		actual, err := s.ClientAuthorizedHandler("ID", exampleGrant)
		assert.Equal(t, expected, actual)
		assert.Error(t, err)
	})

	T.Run("with error reading from database", func(t *testing.T) {
		s := buildTestService(t)
		expected := false
		exampleGrant := oauth2.AuthorizationCode
		exampleClient := &models.OAuth2Client{
			ID:       1,
			ClientID: "blah",
			Scopes:   []string{},
		}
		stringID := fmt.Sprintf("%d", exampleClient.ID)

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			stringID,
		).Return((*models.OAuth2Client)(nil), errors.New("blah"))
		s.database = mockDB

		actual, err := s.ClientAuthorizedHandler(stringID, exampleGrant)
		assert.Equal(t, expected, actual)
		assert.Error(t, err)
	})

	T.Run("with disallowed implicit", func(t *testing.T) {
		s := buildTestService(t)
		expected := false

		exampleGrant := oauth2.Implicit
		exampleClient := &models.OAuth2Client{
			ID:       1,
			ClientID: "blah",
			Scopes:   []string{},
		}
		stringID := fmt.Sprintf("%d", exampleClient.ID)

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			stringID,
		).Return(exampleClient, nil)
		s.database = mockDB

		actual, err := s.ClientAuthorizedHandler(stringID, exampleGrant)
		assert.Equal(t, expected, actual)
		assert.Error(t, err)
	})
}

func TestService_ClientScopeHandler(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService(t)
		expected := true

		exampleScope := "halb"
		exampleClient := &models.OAuth2Client{
			ID:       1,
			ClientID: "blah",
			Scopes:   []string{exampleScope},
		}
		stringID := fmt.Sprintf("%d", exampleClient.ID)

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			stringID,
		).Return(exampleClient, nil)
		s.database = mockDB

		actual, err := s.ClientScopeHandler(stringID, exampleScope)
		assert.Equal(t, expected, actual)
		assert.NoError(t, err)
	})

	T.Run("with error reading from database", func(t *testing.T) {
		s := buildTestService(t)
		expected := false

		exampleScope := "halb"
		exampleClient := &models.OAuth2Client{
			ID:       1,
			ClientID: "blah",
			Scopes:   []string{exampleScope},
		}
		stringID := fmt.Sprintf("%d", exampleClient.ID)

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			stringID,
		).Return((*models.OAuth2Client)(nil), errors.New("blah"))
		s.database = mockDB

		actual, err := s.ClientScopeHandler(stringID, exampleScope)
		assert.Equal(t, expected, actual)
		assert.Error(t, err)
	})

	T.Run("without valid scope", func(t *testing.T) {
		s := buildTestService(t)
		expected := false

		exampleScope := "halb"
		exampleClient := &models.OAuth2Client{
			ID:       1,
			ClientID: "blah",
			Scopes:   []string{},
		}
		stringID := fmt.Sprintf("%d", exampleClient.ID)

		mockDB := database.BuildMockDatabase()
		mockDB.OAuth2ClientDataManager.On(
			"GetOAuth2ClientByClientID",
			mock.Anything,
			stringID,
		).Return(exampleClient, nil)
		s.database = mockDB

		actual, err := s.ClientScopeHandler(stringID, exampleScope)
		assert.Equal(t, expected, actual)
		assert.Error(t, err)
	})
}
