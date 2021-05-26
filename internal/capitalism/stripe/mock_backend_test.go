package stripe

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/form"
)

var _ stripe.Backend = (*mockBackend)(nil)

type mockBackend struct {
	mock.Mock

	anticipatedReturns [][]byte
}

func (m *mockBackend) AnticipateCall(t *testing.T, v interface{}) {
	t.Helper()

	b, err := json.Marshal(v)
	require.NoError(t, err)

	m.anticipatedReturns = append(m.anticipatedReturns, b)
}

func (m *mockBackend) Call(method, path, key string, params stripe.ParamsContainer, v stripe.LastResponseSetter) error {
	b := m.anticipatedReturns[0]
	m.anticipatedReturns = append(m.anticipatedReturns[:0], m.anticipatedReturns[1:]...)

	if err := json.Unmarshal(b, v); err != nil {
		panic(err)
	}

	return m.Called(method, path, key, params, v).Error(0)
}

func (m *mockBackend) CallRaw(method, path, key string, body *form.Values, params *stripe.Params, v stripe.LastResponseSetter) error {
	b := m.anticipatedReturns[0]
	m.anticipatedReturns = append(m.anticipatedReturns[:0], m.anticipatedReturns[1:]...)

	if err := json.Unmarshal(b, v); err != nil {
		panic(err)
	}

	return m.Called(method, path, key, body, params, v).Error(0)
}

func (m *mockBackend) CallMultipart(method, path, key, boundary string, body *bytes.Buffer, params *stripe.Params, v stripe.LastResponseSetter) error {
	b := m.anticipatedReturns[0]
	m.anticipatedReturns = append(m.anticipatedReturns[:0], m.anticipatedReturns[1:]...)

	if err := json.Unmarshal(b, v); err != nil {
		panic(err)
	}

	return m.Called(method, path, key, boundary, body, params, v).Error(0)
}

func (m *mockBackend) SetMaxNetworkRetries(maxNetworkRetries int64) {
	m.Called(maxNetworkRetries)
}
