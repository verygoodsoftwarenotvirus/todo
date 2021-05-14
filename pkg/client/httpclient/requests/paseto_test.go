package requests

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_setSignatureForRequest(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleBody := []byte(t.Name())
		exampleSecretKey := []byte(strings.Repeat("A", validClientSecretSize))
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		expected := "_l92KZfsYpDrPeP8CGTgHQiAtpEg3TyECry5Bd0ibdI"

		assert.NoError(t, setSignatureForRequest(req, exampleBody, exampleSecretKey))
		assert.Equal(t, expected, req.Header.Get(signatureHeaderKey))
	})
}

func TestBuilder_BuildAPIClientAuthTokenRequest(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleInput := fakes.BuildFakePASETOCreationInput()
		exampleSecretKey := []byte(strings.Repeat("A", validClientSecretSize))

		actual, err := h.builder.BuildAPIClientAuthTokenRequest(h.ctx, exampleInput, exampleSecretKey)
		assert.NoError(t, err)
		assert.NotNil(t, actual)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleSecretKey := []byte(strings.Repeat("A", validClientSecretSize))

		actual, err := h.builder.BuildAPIClientAuthTokenRequest(h.ctx, nil, exampleSecretKey)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid secret key", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleInput := fakes.BuildFakePASETOCreationInput()

		actual, err := h.builder.BuildAPIClientAuthTokenRequest(h.ctx, exampleInput, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleInput := &types.PASETOCreationInput{}
		exampleSecretKey := []byte(strings.Repeat("A", validClientSecretSize))

		actual, err := h.builder.BuildAPIClientAuthTokenRequest(h.ctx, exampleInput, exampleSecretKey)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error building request", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		h.builder = buildTestRequestBuilderWithInvalidURL()
		exampleInput := fakes.BuildFakePASETOCreationInput()
		exampleSecretKey := []byte(strings.Repeat("A", validClientSecretSize))

		actual, err := h.builder.BuildAPIClientAuthTokenRequest(h.ctx, exampleInput, exampleSecretKey)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error encoding input to buffer", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleInput := fakes.BuildFakePASETOCreationInput()
		exampleSecretKey := []byte(strings.Repeat("A", validClientSecretSize))

		clientEncoder := &mockencoding.ClientEncoder{}
		clientEncoder.On(
			"EncodeReader",
			mock.Anything, // context.Context
			exampleInput,
		).Return(io.Reader(bytes.NewReader([]byte(""))), nil)
		clientEncoder.On("ContentType").Return("application/fart")
		clientEncoder.On(
			"Encode",
			mock.Anything, // context.Context
			mock.IsType(&bytes.Buffer{}),
			exampleInput,
		).Return(errors.New("blah"))
		h.builder.encoder = clientEncoder

		actual, err := h.builder.BuildAPIClientAuthTokenRequest(h.ctx, exampleInput, exampleSecretKey)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error setting signature", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleInput := fakes.BuildFakePASETOCreationInput()
		exampleSecretKey := []byte("A") // invalid key means the signature fails

		actual, err := h.builder.BuildAPIClientAuthTokenRequest(h.ctx, exampleInput, exampleSecretKey)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}
