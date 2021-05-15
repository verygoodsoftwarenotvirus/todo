package apiclients

import (
	"bytes"
	"database/sql"
	"errors"
	"math"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/mock"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/metrics/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/passwords"
	random "gitlab.com/verygoodsoftwarenotvirus/todo/internal/random"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/mock"
	testutil "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAPIClientsService_ListHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleAPIClientList := fakes.BuildFakeAPIClientList()

		mockDB := database.BuildMockDatabase()
		mockDB.APIClientDataManager.On(
			"GetAPIClients",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleAPIClientList, nil)
		helper.service.apiClientDataManager = mockDB

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
			mock.IsType(&types.APIClientList{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ListHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, encoderDecoder)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ListHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusUnauthorized, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with no results returned from datastore", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.APIClientDataManager.On(
			"GetAPIClients",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.APIClientList)(nil), sql.ErrNoRows)
		helper.service.apiClientDataManager = mockDB
		helper.service.userDataManager = mockDB

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
			mock.IsType(&types.APIClientList{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ListHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, encoderDecoder)
	})

	T.Run("with error retrieving clients from the datastore", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.APIClientDataManager.On(
			"GetAPIClients",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.APIClientList)(nil), errors.New("blah"))
		helper.service.apiClientDataManager = mockDB
		helper.service.userDataManager = mockDB

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ListHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, encoderDecoder)
	})
}

func TestAPIClientsService_CreateHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, helper.exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = mockDB

		a := &passwords.MockAuthenticator{}
		a.On(
			"ValidateLogin",
			testutil.ContextMatcher,
			helper.exampleUser.HashedPassword,
			helper.exampleInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleInput.TOTPToken,
		).Return(true, nil)
		helper.service.authenticator = a

		sg := &random.MockGenerator{}
		sg.On(
			"GenerateBase64EncodedString",
			testutil.ContextMatcher,
			clientIDSize,
		).Return(helper.exampleAPIClient.ClientID, nil)
		sg.On(
			"GenerateRawBytes",
			testutil.ContextMatcher,
			clientSecretSize,
		).Return(helper.exampleAPIClient.ClientSecret, nil)
		helper.service.secretGenerator = sg

		mockDB.APIClientDataManager.On(
			"CreateAPIClient",
			testutil.ContextMatcher,
			helper.exampleInput,
			helper.exampleUser.ID,
		).Return(helper.exampleAPIClient, nil)
		helper.service.apiClientDataManager = mockDB

		uc := &mockmetrics.UnitCounter{}
		uc.On("Increment", testutil.ContextMatcher).Return()
		helper.service.apiClientCounter = uc

		helper.service.CreateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusCreated, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, a, sg, uc)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

		helper.service.CreateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)
	})

	T.Run("without input attached to request", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(nil))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.service.CreateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
	})

	T.Run("with invalid input attached to request", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

		helper.service.cfg.minimumPasswordLength = math.MaxUint8
		helper.exampleInput.Password = ""
		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, helper.exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		helper.service.CreateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
	})

	T.Run("with error retrieving user", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, helper.exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return((*types.User)(nil), errors.New("blah"))
		helper.service.apiClientDataManager = mockDB
		helper.service.userDataManager = mockDB

		helper.service.CreateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with invalid credentials", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, helper.exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		mockDB.APIClientDataManager.On(
			"CreateAPIClient",
			testutil.ContextMatcher,
			helper.exampleInput,
			helper.exampleUser.ID,
		).Return(helper.exampleAPIClient, nil)
		helper.service.apiClientDataManager = mockDB
		helper.service.userDataManager = mockDB

		a := &passwords.MockAuthenticator{}
		a.On(
			"ValidateLogin",
			testutil.ContextMatcher,
			helper.exampleUser.HashedPassword,
			helper.exampleInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleInput.TOTPToken,
		).Return(false, nil)
		helper.service.authenticator = a

		helper.service.CreateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, a)
	})

	T.Run("with invalid password", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, helper.exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		mockDB.APIClientDataManager.On(
			"CreateAPIClient",
			testutil.ContextMatcher,
			helper.exampleInput,
			helper.exampleUser.ID,
		).Return(helper.exampleAPIClient, nil)
		helper.service.apiClientDataManager = mockDB
		helper.service.userDataManager = mockDB

		a := &passwords.MockAuthenticator{}
		a.On(
			"ValidateLogin",
			testutil.ContextMatcher,
			helper.exampleUser.HashedPassword,
			helper.exampleInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleInput.TOTPToken,
		).Return(true, errors.New("blah"))
		helper.service.authenticator = a

		helper.service.CreateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, a)
	})

	T.Run("with error generating client ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, helper.exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)

		a := &passwords.MockAuthenticator{}
		a.On(
			"ValidateLogin",
			testutil.ContextMatcher,
			helper.exampleUser.HashedPassword,
			helper.exampleInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleInput.TOTPToken,
		).Return(true, nil)
		helper.service.authenticator = a

		sg := &random.MockGenerator{}
		sg.On(
			"GenerateBase64EncodedString",
			testutil.ContextMatcher,
			clientIDSize,
		).Return("", errors.New("blah"))
		helper.service.secretGenerator = sg

		mockDB.APIClientDataManager.On(
			"CreateAPIClient",
			testutil.ContextMatcher,
			helper.exampleInput,
			helper.exampleUser.ID,
		).Return(helper.exampleAPIClient, nil)

		helper.service.apiClientDataManager = mockDB
		helper.service.userDataManager = mockDB

		helper.service.CreateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, a, sg)
	})

	T.Run("with error generating client secret", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, helper.exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)
		helper.service.userDataManager = mockDB

		a := &passwords.MockAuthenticator{}
		a.On(
			"ValidateLogin",
			testutil.ContextMatcher,
			helper.exampleUser.HashedPassword,
			helper.exampleInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleInput.TOTPToken,
		).Return(true, nil)
		helper.service.authenticator = a

		sg := &random.MockGenerator{}
		sg.On(
			"GenerateBase64EncodedString",
			testutil.ContextMatcher,
			clientIDSize,
		).Return(helper.exampleAPIClient.ClientID, nil)
		sg.On(
			"GenerateRawBytes",
			testutil.ContextMatcher,
			clientSecretSize,
		).Return([]byte(nil), errors.New("blah"))
		helper.service.secretGenerator = sg

		mockDB.APIClientDataManager.On(
			"CreateAPIClient",
			testutil.ContextMatcher,
			helper.exampleInput,
			helper.exampleUser.ID,
		).Return(helper.exampleAPIClient, nil)
		helper.service.apiClientDataManager = mockDB

		helper.service.CreateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, a, sg)
	})

	T.Run("with error creating API Client in data store", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)

		jsonBytes := helper.service.encoderDecoder.MustEncode(helper.ctx, helper.exampleInput)

		var err error
		helper.req, err = http.NewRequestWithContext(helper.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, helper.req)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutil.ContextMatcher,
			helper.exampleUser.ID,
		).Return(helper.exampleUser, nil)

		a := &passwords.MockAuthenticator{}
		a.On(
			"ValidateLogin",
			testutil.ContextMatcher,
			helper.exampleUser.HashedPassword,
			helper.exampleInput.Password,
			helper.exampleUser.TwoFactorSecret,
			helper.exampleInput.TOTPToken,
		).Return(true, nil)
		helper.service.authenticator = a

		sg := &random.MockGenerator{}
		sg.On(
			"GenerateBase64EncodedString",
			testutil.ContextMatcher,
			clientIDSize,
		).Return(helper.exampleAPIClient.ClientID, nil)
		sg.On(
			"GenerateRawBytes",
			testutil.ContextMatcher,
			clientSecretSize,
		).Return(helper.exampleAPIClient.ClientSecret, nil)
		helper.service.secretGenerator = sg

		mockDB.APIClientDataManager.On(
			"CreateAPIClient",
			testutil.ContextMatcher,
			helper.exampleInput,
			helper.exampleUser.ID,
		).Return((*types.APIClient)(nil), errors.New("blah"))

		helper.service.apiClientDataManager = mockDB
		helper.service.userDataManager = mockDB

		helper.service.CreateHandler(helper.res, helper.req)
		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, a, sg)
	})
}

func TestAPIClientsService_ReadHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		apiClientDataManager := &mocktypes.APIClientDataManager{}
		apiClientDataManager.On(
			"GetAPIClientByDatabaseID",
			testutil.ContextMatcher,
			helper.exampleAPIClient.ID,
			helper.exampleUser.ID,
		).Return(helper.exampleAPIClient, nil)
		helper.service.apiClientDataManager = apiClientDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
			mock.IsType(&types.APIClient{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ReadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, apiClientDataManager, encoderDecoder)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ReadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with no such API client in the database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		apiClientDataManager := &mocktypes.APIClientDataManager{}
		apiClientDataManager.On(
			"GetAPIClientByDatabaseID",
			testutil.ContextMatcher,
			helper.exampleAPIClient.ID,
			helper.exampleUser.ID,
		).Return((*types.APIClient)(nil), sql.ErrNoRows)
		helper.service.apiClientDataManager = apiClientDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ReadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, apiClientDataManager, encoderDecoder)
	})

	T.Run("with error fetching from database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		apiClientDataManager := &mocktypes.APIClientDataManager{}
		apiClientDataManager.On(
			"GetAPIClientByDatabaseID",
			testutil.ContextMatcher,
			helper.exampleAPIClient.ID,
			helper.exampleUser.ID,
		).Return((*types.APIClient)(nil), errors.New("blah"))
		helper.service.apiClientDataManager = apiClientDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
		)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ReadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, apiClientDataManager, encoderDecoder)
	})
}

func TestAPIClientsService_ArchiveHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		apiClientDataManager := &mocktypes.APIClientDataManager{}
		apiClientDataManager.On(
			"ArchiveAPIClient",
			testutil.ContextMatcher,
			helper.exampleAPIClient.ID,
			helper.exampleAccount.ID,
			helper.exampleUser.ID,
		).Return(nil)
		helper.service.apiClientDataManager = apiClientDataManager

		unitCounter := &mockmetrics.UnitCounter{}
		unitCounter.On("Decrement", testutil.ContextMatcher).Return()
		helper.service.apiClientCounter = unitCounter

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNoContent, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, apiClientDataManager, unitCounter)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with no such API client in the database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		apiClientDataManager := &mocktypes.APIClientDataManager{}
		apiClientDataManager.On(
			"ArchiveAPIClient",
			testutil.ContextMatcher,
			helper.exampleAPIClient.ID,
			helper.exampleAccount.ID,
			helper.exampleUser.ID,
		).Return(sql.ErrNoRows)
		helper.service.apiClientDataManager = apiClientDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, apiClientDataManager, encoderDecoder)
	})

	T.Run("with error fetching from database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		apiClientDataManager := &mocktypes.APIClientDataManager{}
		apiClientDataManager.On(
			"ArchiveAPIClient",
			testutil.ContextMatcher,
			helper.exampleAPIClient.ID,
			helper.exampleAccount.ID,
			helper.exampleUser.ID,
		).Return(errors.New("blah"))
		helper.service.apiClientDataManager = apiClientDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
		)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, apiClientDataManager, encoderDecoder)
	})
}

func TestAPIClientsService_AuditEntryHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleAuditLogEntries := fakes.BuildFakeAuditLogEntryList().Entries

		apiClientDataManager := &mocktypes.APIClientDataManager{}
		apiClientDataManager.On(
			"GetAuditLogEntriesForAPIClient",
			testutil.ContextMatcher,
			helper.exampleAPIClient.ID,
		).Return(exampleAuditLogEntries, nil)
		helper.service.apiClientDataManager = apiClientDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
			mock.IsType([]*types.AuditLogEntry{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.AuditEntryHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, apiClientDataManager, encoderDecoder)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.AuditEntryHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)

		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		apiClientDataManager := &mocktypes.APIClientDataManager{}
		apiClientDataManager.On(
			"GetAuditLogEntriesForAPIClient",
			testutil.ContextMatcher,
			helper.exampleAPIClient.ID,
		).Return([]*types.AuditLogEntry(nil), sql.ErrNoRows)
		helper.service.apiClientDataManager = apiClientDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.AuditEntryHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)

		mock.AssertExpectationsForObjects(t, apiClientDataManager, encoderDecoder)
	})

	T.Run("with error reading from database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		apiClientDataManager := &mocktypes.APIClientDataManager{}
		apiClientDataManager.On(
			"GetAuditLogEntriesForAPIClient",
			testutil.ContextMatcher,
			helper.exampleAPIClient.ID,
		).Return([]*types.AuditLogEntry(nil), errors.New("blah"))
		helper.service.apiClientDataManager = apiClientDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.HTTPResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.AuditEntryHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)

		mock.AssertExpectationsForObjects(t, apiClientDataManager, encoderDecoder)
	})
}