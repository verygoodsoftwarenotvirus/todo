package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_newPASETORoundTripper(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		c, _ := buildSimpleTestClient(t)
		exampleClientID := "example_client_id"
		exampleSecret := make([]byte, validClientSecretSize)

		assert.NotNil(t, newPASETORoundTripper(c, exampleClientID, exampleSecret))
	})
}

func Test_pasetoRoundTripper_RoundTrip(T *testing.T) {
	T.Parallel()

	originalPASETORoundTripperClient := pasetoRoundTripperClient

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		anticipatedResult := "v2.local.QAxIpVe-ECVNI1z4xQbm_qQYomyT3h8FtV8bxkz8pBJWkT8f7HtlOpbroPDEZUKop_vaglyp76CzYy375cHmKCW8e1CCkV0Lflu4GTDyXMqQdpZMM1E6OaoQW27gaRSvWBrR3IgbFIa0AkuUFw.UGFyYWdvbiBJbml0aWF0aXZlIEVudGVycHJpc2Vz"

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				response := &types.PASETOResponse{Token: anticipatedResult}

				assert.NotEmpty(t, req.Header.Get(signatureHeaderKey))

				require.NoError(t, json.NewEncoder(res).Encode(response))
			},
		))

		c := buildTestClient(t, ts)

		pasetoRoundTripperClient = ts.Client()

		exampleClientID := "example_client_id"
		exampleSecret := make([]byte, validClientSecretSize)
		rt := newPASETORoundTripper(c, exampleClientID, exampleSecret)

		exampleResponse := &http.Response{
			StatusCode: http.StatusTeapot,
		}

		mrt := &mockRoundTripper{}
		mrt.On("RoundTrip", mock.IsType(&http.Request{})).Return(exampleResponse, nil)
		rt.base = mrt

		req := httptest.NewRequest(http.MethodPost, c.URL().String(), nil)

		res, err := rt.RoundTrip(req)
		assert.NoError(t, err)
		assert.NotNil(t, res)

		assert.Equal(t, exampleResponse, res)
		mock.AssertExpectationsForObjects(t, mrt)
	})

	T.Run("with error fetching token", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assert.NotEmpty(t, req.Header.Get(signatureHeaderKey))

				res.WriteHeader(http.StatusUnauthorized)
			},
		))

		c := buildTestClient(t, ts)

		pasetoRoundTripperClient = ts.Client()

		exampleClientID := "example_client_id"
		exampleSecret := make([]byte, validClientSecretSize)
		rt := newPASETORoundTripper(c, exampleClientID, exampleSecret)

		req := httptest.NewRequest(http.MethodPost, c.URL().String(), nil)

		res, err := rt.RoundTrip(req)
		assert.Error(t, err)
		assert.Nil(t, res)
	})

	T.Run("with error executing RoundTrip", func(t *testing.T) {
		t.Parallel()

		anticipatedResult := "v2.local.QAxIpVe-ECVNI1z4xQbm_qQYomyT3h8FtV8bxkz8pBJWkT8f7HtlOpbroPDEZUKop_vaglyp76CzYy375cHmKCW8e1CCkV0Lflu4GTDyXMqQdpZMM1E6OaoQW27gaRSvWBrR3IgbFIa0AkuUFw.UGFyYWdvbiBJbml0aWF0aXZlIEVudGVycHJpc2Vz"

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				response := &types.PASETOResponse{Token: anticipatedResult}

				assert.NotEmpty(t, req.Header.Get(signatureHeaderKey))

				require.NoError(t, json.NewEncoder(res).Encode(response))
			},
		))

		c := buildTestClient(t, ts)

		pasetoRoundTripperClient = ts.Client()

		exampleClientID := "example_client_id"
		exampleSecret := make([]byte, validClientSecretSize)
		rt := newPASETORoundTripper(c, exampleClientID, exampleSecret)

		mrt := &mockRoundTripper{}
		mrt.On("RoundTrip", mock.IsType(&http.Request{})).Return((*http.Response)(nil), errors.New("blah"))
		rt.base = mrt

		req := httptest.NewRequest(http.MethodPost, c.URL().String(), nil)

		res, err := rt.RoundTrip(req)
		assert.Error(t, err)
		assert.Nil(t, res)

		mock.AssertExpectationsForObjects(t, mrt)
	})

	pasetoRoundTripperClient = originalPASETORoundTripperClient
}
