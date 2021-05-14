package encoding

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"

	"github.com/stretchr/testify/assert"
)

func TestProvideClientEncoder(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, ProvideClientEncoder(logging.NewNonOperationalLogger(), ContentTypeJSON))
	})
}

func Test_clientEncoder_Unmarshal(T *testing.T) {
	T.Parallel()

	T.Run("with JSON", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		e := ProvideClientEncoder(logging.NewNonOperationalLogger(), ContentTypeJSON)

		expected := &example{Name: "name"}
		actual := &example{}

		assert.NoError(t, e.Unmarshal(ctx, []byte(`{"name": "name"}`), &actual))
		assert.Equal(t, expected, actual)
	})

	T.Run("with XML", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		e := ProvideClientEncoder(logging.NewNonOperationalLogger(), ContentTypeXML)

		expected := &example{Name: "name"}
		actual := &example{}

		assert.NoError(t, e.Unmarshal(ctx, []byte(`<example><name>name</name></example>`), &actual))
		assert.Equal(t, expected, actual)
	})

	T.Run("with invalid data", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		e := ProvideClientEncoder(logging.NewNonOperationalLogger(), ContentTypeJSON)

		actual := &example{}

		assert.Error(t, e.Unmarshal(ctx, []byte(`{"name"   `), &actual))
		assert.Empty(t, actual.Name)
	})
}

func Test_clientEncoder_Encode(T *testing.T) {
	T.Parallel()

	T.Run("with JSON", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		e := ProvideClientEncoder(logging.NewNonOperationalLogger(), ContentTypeJSON)

		res := httptest.NewRecorder()

		assert.NoError(t, e.Encode(ctx, res, &example{Name: t.Name()}))
	})

	T.Run("with XML", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		e := ProvideClientEncoder(logging.NewNonOperationalLogger(), ContentTypeXML)

		res := httptest.NewRecorder()

		assert.NoError(t, e.Encode(ctx, res, &example{Name: t.Name()}))
	})

	T.Run("with invalid data", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		e := ProvideClientEncoder(logging.NewNonOperationalLogger(), ContentTypeJSON)

		assert.Error(t, e.Encode(ctx, nil, &broken{Name: json.Number(t.Name())}))
	})
}

func Test_clientEncoder_EncodeReader(T *testing.T) {
	T.Parallel()

	T.Run("with JSON", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		e := ProvideClientEncoder(logging.NewNonOperationalLogger(), ContentTypeJSON)

		actual, err := e.EncodeReader(ctx, &example{Name: t.Name()})
		assert.NoError(t, err)
		assert.NotNil(t, actual)
	})

	T.Run("with XML", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		e := ProvideClientEncoder(logging.NewNonOperationalLogger(), ContentTypeXML)

		actual, err := e.EncodeReader(ctx, &example{Name: t.Name()})
		assert.NoError(t, err)
		assert.NotNil(t, actual)
	})

	T.Run("with invalid data", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		e := ProvideClientEncoder(logging.NewNonOperationalLogger(), ContentTypeJSON)

		actual, err := e.EncodeReader(ctx, &broken{Name: json.Number(t.Name())})
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}
