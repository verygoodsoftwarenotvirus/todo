package accounts

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAccountsServiceMiddleware(t *testing.T) {
	suite.Run(t, new(accountsServiceMiddlewareTestSuite))
}

type accountsServiceMiddlewareTestSuite struct {
	suite.Suite

	ctx            context.Context
	service        *service
	exampleUser    *types.User
	exampleAccount *types.Account
}

var _ suite.SetupTestSuite = (*accountsServiceMiddlewareTestSuite)(nil)

func (s *accountsServiceMiddlewareTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.service = buildTestService()
	s.exampleUser = fakes.BuildFakeUser()
	s.exampleAccount = fakes.BuildFakeAccount()

	reqCtx, err := types.RequestContextFromUser(s.exampleUser, s.exampleAccount.ID, map[uint64]permissions.ServiceUserPermissions{
		s.exampleAccount.ID: testutil.BuildMaxUserPerms(),
	})
	require.NoError(s.T(), err)

	s.service.encoderDecoder = encoding.ProvideServerEncoderDecoder(logging.NewNonOperationalLogger(), encoding.ContentTypeJSON)
	s.service.requestContextFetcher = func(_ *http.Request) (*types.RequestContext, error) {
		return reqCtx, nil
	}
	s.service.accountIDFetcher = func(req *http.Request) uint64 {
		return s.exampleAccount.ID
	}
}

func (s *accountsServiceMiddlewareTestSuite) TestServiceCreationInputMiddleware() {
	t := s.T()

	exampleCreationInput := fakes.BuildFakeAccountCreationInput()
	jsonBytes, err := json.Marshal(&exampleCreationInput)
	require.NoError(t, err)

	mh := &testutil.MockHTTPHandler{}
	mh.On("ServeHTTP", mock.IsType(http.ResponseWriter(httptest.NewRecorder())), mock.IsType(&http.Request{})).Return()

	res := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
	require.NoError(t, err)
	require.NotNil(t, req)

	actual := s.service.CreationInputMiddleware(mh)
	actual.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

	mock.AssertExpectationsForObjects(t, mh)
}

func (s *accountsServiceMiddlewareTestSuite) TestServiceCreationInputMiddlewareWithInvalidInput() {
	t := s.T()

	exampleCreationInput := &types.AccountCreationInput{}
	jsonBytes, err := json.Marshal(&exampleCreationInput)
	require.NoError(t, err)

	res := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
	require.NoError(t, err)
	require.NotNil(t, req)

	actual := s.service.CreationInputMiddleware(&testutil.MockHTTPHandler{})
	actual.ServeHTTP(res, req)

	assert.Equal(t, http.StatusBadRequest, res.Code)
}

func (s *accountsServiceMiddlewareTestSuite) TestServiceCreationInputMiddlewareWithErrorDecodingRequest() {
	t := s.T()

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("DecodeRequest", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.RequestMatcher()), mock.IsType(&types.AccountCreationInput{})).Return(errors.New("blah"))
	ed.On(
		"EncodeErrorResponse",
		mock.MatchedBy(testutil.ContextMatcher),
		mock.IsType(http.ResponseWriter(httptest.NewRecorder())),
		"invalid request content",
		http.StatusBadRequest,
	)
	s.service.encoderDecoder = ed

	res := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
	require.NoError(t, err)
	require.NotNil(t, req)

	mh := &testutil.MockHTTPHandler{}
	actual := s.service.CreationInputMiddleware(mh)
	actual.ServeHTTP(res, req)

	assert.Equal(t, http.StatusBadRequest, res.Code)

	mock.AssertExpectationsForObjects(t, ed, mh)
}

func (s *accountsServiceMiddlewareTestSuite) TestServiceUpdateInputMiddleware() {
	t := s.T()

	exampleCreationInput := fakes.BuildFakeAccountUpdateInputFromAccount(fakes.BuildFakeAccount())
	jsonBytes, err := json.Marshal(&exampleCreationInput)
	require.NoError(t, err)

	mh := &testutil.MockHTTPHandler{}
	mh.On("ServeHTTP", mock.IsType(http.ResponseWriter(httptest.NewRecorder())), mock.IsType(&http.Request{})).Return()

	res := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", bytes.NewReader(jsonBytes))
	require.NoError(t, err)
	require.NotNil(t, req)

	actual := s.service.UpdateInputMiddleware(mh)
	actual.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

	mock.AssertExpectationsForObjects(t, mh)
}

func (s *accountsServiceMiddlewareTestSuite) TestServiceUpdateInputMiddlewareWithInvalidInput() {
	t := s.T()

	ed := mockencoding.NewMockEncoderDecoder()
	ed.On("DecodeRequest", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.RequestMatcher()), mock.IsType(&types.AccountUpdateInput{})).Return(errors.New("blah"))
	ed.On(
		"EncodeErrorResponse",
		mock.MatchedBy(testutil.ContextMatcher),
		mock.IsType(http.ResponseWriter(httptest.NewRecorder())),
		"invalid request content",
		http.StatusBadRequest,
	)
	s.service.encoderDecoder = ed

	res := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
	require.NoError(t, err)
	require.NotNil(t, req)

	mh := &testutil.MockHTTPHandler{}
	actual := s.service.UpdateInputMiddleware(mh)
	actual.ServeHTTP(res, req)

	assert.Equal(t, http.StatusBadRequest, res.Code)

	mock.AssertExpectationsForObjects(t, ed, mh)
}
