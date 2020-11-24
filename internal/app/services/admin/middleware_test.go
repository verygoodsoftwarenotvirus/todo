package admin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_AccountStatusUpdateInputMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)
		s.encoderDecoder = &encoding.ServerEncoderDecoder{}

		exampleInput := fakes.BuildFakeAccountStatusUpdateInput()
		jsonBytes, err := json.Marshal(&exampleInput)
		require.NoError(t, err)

		mh := &testutil.MockHTTPHandler{}
		mh.On("ServeHTTP", mock.Anything, mock.Anything).Return()

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", strings.NewReader(string(jsonBytes)))
		require.NoError(t, err)
		require.NotNil(t, req)

		actual := s.AccountStatusUpdateInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mh)
	})

	T.Run("bad input", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)
		s.encoderDecoder = &encoding.ServerEncoderDecoder{}

		exampleCreationInput := &types.AccountStatusUpdateInput{}
		jsonBytes, err := json.Marshal(&exampleCreationInput)
		require.NoError(t, err)

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", strings.NewReader(string(jsonBytes)))
		require.NoError(t, err)
		require.NotNil(t, req)

		actual := s.AccountStatusUpdateInputMiddleware(&testutil.MockHTTPHandler{})
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService(t)

		ed := &mockencoding.EncoderDecoder{}
		ed.On("DecodeRequest", mock.Anything, mock.Anything).Return(errors.New("blah"))
		ed.On(
			"EncodeErrorResponse",
			mock.Anything,
			"invalid request content",
			http.StatusBadRequest,
		)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://todo.verygoodsoftwarenotvirus.ru", nil)
		require.NoError(t, err)
		require.NotNil(t, req)

		mh := &testutil.MockHTTPHandler{}
		actual := s.AccountStatusUpdateInputMiddleware(mh)
		actual.ServeHTTP(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)

		mock.AssertExpectationsForObjects(t, ed, mh)
	})
}
