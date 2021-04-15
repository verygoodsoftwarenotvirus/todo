package accounts

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type accountsServiceMiddlewareTestHelper struct {
	ctx            context.Context
	service        *service
	exampleUser    *types.User
	exampleAccount *types.Account
}

func buildMiddlewareTestHelper(t *testing.T) *accountsServiceMiddlewareTestHelper {
	t.Helper()

	h := &accountsServiceMiddlewareTestHelper{}

	h.ctx = context.Background()
	h.service = buildTestService()
	h.exampleUser = fakes.BuildFakeUser()
	h.exampleAccount = fakes.BuildFakeAccount()

	sessionCtxData, err := types.SessionContextDataFromUser(
		h.exampleUser,
		h.exampleAccount.ID,
		map[uint64]*types.UserAccountMembershipInfo{
			h.exampleAccount.ID: {
				AccountName: h.exampleAccount.Name,
				Permissions: testutil.BuildMaxUserPerms(),
			},
		},
	)
	require.NoError(t, err)

	h.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)
	h.service.sessionContextDataFetcher = func(_ *http.Request) (*types.SessionContextData, error) {
		return sessionCtxData, nil
	}
	h.service.accountIDFetcher = func(req *http.Request) uint64 {
		return h.exampleAccount.ID
	}

	return h
}

func TestService_CreationInputMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildMiddlewareTestHelper(t)

		exampleCreationInput := fakes.BuildFakeAccountCreationInput()
		jsonBytes, err := json.Marshal(&exampleCreationInput)
		require.NoError(t, err)

		mh := &testutil.MockHTTPHandler{}
		mh.On(
			"ServeHTTP",
			mock.IsType(http.ResponseWriter(httptest.NewRecorder())),
			mock.IsType(&http.Request{}),
		).Return()

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, req)

		actual := s.service.CreationInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		t.Parallel()

		s := buildMiddlewareTestHelper(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"DecodeRequest",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.MatchedBy(testutil.RequestMatcher()),
			mock.IsType(&types.AccountCreationInput{}),
		).Return(errors.New("blah"))
		encoderDecoder.On(
			"EncodeErrorResponse",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.IsType(http.ResponseWriter(httptest.NewRecorder())),
			"invalid request content",
			http.StatusBadRequest,
		)
		s.service.encoderDecoder = encoderDecoder

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		mh := &testutil.MockHTTPHandler{}
		actual := s.service.CreationInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)

		mock.AssertExpectationsForObjects(t, encoderDecoder, mh)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		s := buildMiddlewareTestHelper(t)

		exampleCreationInput := &types.AccountCreationInput{}
		jsonBytes, err := json.Marshal(&exampleCreationInput)
		require.NoError(t, err)

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, req)

		mh := &testutil.MockHTTPHandler{}
		actual := s.service.CreationInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})
}

func TestService_UpdateInputMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildMiddlewareTestHelper(t)

		exampleCreationInput := fakes.BuildFakeAccountUpdateInputFromAccount(fakes.BuildFakeAccount())
		jsonBytes, err := json.Marshal(&exampleCreationInput)
		require.NoError(t, err)

		mh := &testutil.MockHTTPHandler{}
		mh.On(
			"ServeHTTP",
			mock.IsType(http.ResponseWriter(httptest.NewRecorder())),
			mock.IsType(&http.Request{}),
		).Return()

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, req)

		actual := s.service.UpdateInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		t.Parallel()

		s := buildMiddlewareTestHelper(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"DecodeRequest",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.MatchedBy(testutil.RequestMatcher()),
			mock.IsType(&types.AccountUpdateInput{}),
		).Return(errors.New("blah"))
		encoderDecoder.On(
			"EncodeErrorResponse",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.IsType(http.ResponseWriter(httptest.NewRecorder())),
			"invalid request content",
			http.StatusBadRequest,
		)
		s.service.encoderDecoder = encoderDecoder

		mh := &testutil.MockHTTPHandler{}

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		actual := s.service.UpdateInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})

	T.Run("with empty input", func(t *testing.T) {
		t.Parallel()

		s := buildMiddlewareTestHelper(t)
		mh := &testutil.MockHTTPHandler{}

		exampleCreationInput := &types.AccountUpdateInput{}
		jsonBytes, err := json.Marshal(&exampleCreationInput)
		require.NoError(t, err)

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, req)

		actual := s.service.UpdateInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})
}

func TestService_AddMemberInputMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildMiddlewareTestHelper(t)

		exampleCreationInput := fakes.BuildFakeAddUserToAccountInput()
		jsonBytes, err := json.Marshal(&exampleCreationInput)
		require.NoError(t, err)

		mh := &testutil.MockHTTPHandler{}
		mh.On(
			"ServeHTTP",
			mock.IsType(http.ResponseWriter(httptest.NewRecorder())),
			mock.IsType(&http.Request{}),
		).Return()

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, req)

		actual := s.service.AddMemberInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		t.Parallel()

		s := buildMiddlewareTestHelper(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"DecodeRequest",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.MatchedBy(testutil.RequestMatcher()),
			mock.IsType(&types.AddUserToAccountInput{}),
		).Return(errors.New("blah"))
		encoderDecoder.On(
			"EncodeErrorResponse",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.IsType(http.ResponseWriter(httptest.NewRecorder())),
			"invalid request content",
			http.StatusBadRequest,
		)
		s.service.encoderDecoder = encoderDecoder

		mh := &testutil.MockHTTPHandler{}

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		actual := s.service.AddMemberInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})

	T.Run("with empty input", func(t *testing.T) {
		t.Parallel()

		s := buildMiddlewareTestHelper(t)
		mh := &testutil.MockHTTPHandler{}

		exampleCreationInput := &types.AddUserToAccountInput{}
		jsonBytes, err := json.Marshal(&exampleCreationInput)
		require.NoError(t, err)

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, req)

		actual := s.service.AddMemberInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})
}

func TestService_ModifyMemberPermissionsInputMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildMiddlewareTestHelper(t)

		exampleCreationInput := fakes.BuildFakeUserPermissionModificationInput()
		jsonBytes, err := json.Marshal(&exampleCreationInput)
		require.NoError(t, err)

		mh := &testutil.MockHTTPHandler{}
		mh.On(
			"ServeHTTP",
			mock.IsType(http.ResponseWriter(httptest.NewRecorder())),
			mock.IsType(&http.Request{}),
		).Return()

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, req)

		actual := s.service.ModifyMemberPermissionsInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		t.Parallel()

		s := buildMiddlewareTestHelper(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"DecodeRequest",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.MatchedBy(testutil.RequestMatcher()),
			mock.IsType(&types.ModifyUserPermissionsInput{}),
		).Return(errors.New("blah"))
		encoderDecoder.On(
			"EncodeErrorResponse",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.IsType(http.ResponseWriter(httptest.NewRecorder())),
			"invalid request content",
			http.StatusBadRequest,
		)
		s.service.encoderDecoder = encoderDecoder

		mh := &testutil.MockHTTPHandler{}

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		actual := s.service.ModifyMemberPermissionsInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})

	T.Run("with empty input", func(t *testing.T) {
		t.Parallel()

		s := buildMiddlewareTestHelper(t)
		mh := &testutil.MockHTTPHandler{}

		exampleCreationInput := &types.ModifyUserPermissionsInput{}
		jsonBytes, err := json.Marshal(&exampleCreationInput)
		require.NoError(t, err)

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, req)

		actual := s.service.ModifyMemberPermissionsInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})
}

func TestService_AccountTransferInputMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildMiddlewareTestHelper(t)

		exampleCreationInput := fakes.BuildFakeTransferAccountOwnershipInput()
		jsonBytes, err := json.Marshal(&exampleCreationInput)
		require.NoError(t, err)

		mh := &testutil.MockHTTPHandler{}
		mh.On(
			"ServeHTTP",
			mock.IsType(http.ResponseWriter(httptest.NewRecorder())),
			mock.IsType(&http.Request{}),
		).Return()

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, req)

		actual := s.service.AccountTransferInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		t.Parallel()

		s := buildMiddlewareTestHelper(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"DecodeRequest",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.MatchedBy(testutil.RequestMatcher()),
			mock.IsType(&types.TransferAccountOwnershipInput{}),
		).Return(errors.New("blah"))
		encoderDecoder.On(
			"EncodeErrorResponse",
			mock.MatchedBy(testutil.ContextMatcher),
			mock.IsType(http.ResponseWriter(httptest.NewRecorder())),
			"invalid request content",
			http.StatusBadRequest,
		)
		s.service.encoderDecoder = encoderDecoder

		mh := &testutil.MockHTTPHandler{}

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		actual := s.service.AccountTransferInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})

	T.Run("with empty input", func(t *testing.T) {
		t.Parallel()

		s := buildMiddlewareTestHelper(t)
		mh := &testutil.MockHTTPHandler{}

		exampleCreationInput := &types.TransferAccountOwnershipInput{}
		jsonBytes, err := json.Marshal(&exampleCreationInput)
		require.NoError(t, err)

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, "https://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
		require.NoError(t, err)
		require.NotNil(t, req)

		actual := s.service.AccountTransferInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})
}
