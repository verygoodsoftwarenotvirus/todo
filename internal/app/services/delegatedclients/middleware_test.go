package delegatedclients

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_CreationInputMiddleware(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ed := &mockencoding.EncoderDecoder{}
		ed.On(
			"DecodeRequest",
			mock.Anything,
			mock.AnythingOfType("*http.Request"),
			mock.Anything,
		).Return(nil)
		s.encoderDecoder = ed

		mh := &testutil.MockHTTPHandler{}
		mh.On(
			"ServeHTTP",
			mock.Anything,
			mock.Anything,
		)

		h := s.CreationInputMiddleware(mh)
		req := buildRequest(t)
		res := httptest.NewRecorder()

		expected := fakes.BuildFakeDelegatedClientCreationInput()
		bs, err := json.Marshal(expected)
		require.NoError(t, err)
		req.Body = ioutil.NopCloser(bytes.NewReader(bs))

		h.ServeHTTP(res, req)
		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, ed, mh)
	})

	T.Run("with error decoding request", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ed := &mockencoding.EncoderDecoder{}
		ed.On(
			"DecodeRequest", mock.Anything,
			mock.AnythingOfType("*http.Request"),
			mock.Anything,
		).Return(errors.New("blah"))
		ed.On(
			"EncodeErrorResponse", mock.Anything,
			mock.Anything,
			"invalid request content",
			http.StatusBadRequest,
		)
		s.encoderDecoder = ed

		mh := &testutil.MockHTTPHandler{}
		h := s.CreationInputMiddleware(mh)
		req := buildRequest(t)
		res := httptest.NewRecorder()

		h.ServeHTTP(res, req)
		assert.Equal(t, http.StatusBadRequest, res.Code)

		mock.AssertExpectationsForObjects(t, ed, mh)
	})
}

func TestService_fetchDelegatedClientFromRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s := buildTestService(t)

		exampleDelegatedClient := fakes.BuildFakeDelegatedClient()

		req := buildRequest(t).WithContext(
			context.WithValue(
				ctx,
				types.DelegatedClientKey,
				exampleDelegatedClient,
			),
		)

		actual := s.fetchDelegatedClientFromRequest(req)
		assert.Equal(t, exampleDelegatedClient, actual)
	})

	T.Run("without value present", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		assert.Nil(t, s.fetchDelegatedClientFromRequest(buildRequest(t)))
	})
}

func TestService_fetchDelegatedClientIDFromRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s := buildTestService(t)
		exampleDelegatedClient := fakes.BuildFakeDelegatedClient()

		req := buildRequest(t).WithContext(
			context.WithValue(
				ctx,
				clientIDKey,
				exampleDelegatedClient.ClientID,
			),
		)

		actual := s.fetchDelegatedClientIDFromRequest(req)
		assert.Equal(t, exampleDelegatedClient.ClientID, actual)
	})

	T.Run("without value present", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		assert.Empty(t, s.fetchDelegatedClientIDFromRequest(buildRequest(t)))
	})
}
