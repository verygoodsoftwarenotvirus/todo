package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type testingType struct {
	Name string `json:"name"`
}

func TestArgIsNotPointerOrNil(T *testing.T) {
	T.Parallel()

	T.Run("expected use", func(t *testing.T) {
		t.Parallel()
		err := argIsNotPointerOrNil(&testingType{})
		assert.NoError(t, err, "error should not be returned when a pointer is provided")
	})

	T.Run("with non-pointer", func(t *testing.T) {
		t.Parallel()
		err := argIsNotPointerOrNil(testingType{})
		assert.Error(t, err, "error should be returned when a non-pointer is provided")
	})

	T.Run("with nil", func(t *testing.T) {
		t.Parallel()
		err := argIsNotPointerOrNil(nil)
		assert.Error(t, err, "error should be returned when nil is provided")
	})
}

func TestArgIsNotPointer(T *testing.T) {
	T.Parallel()

	T.Run("expected use", func(t *testing.T) {
		t.Parallel()
		notAPointer, err := argIsNotPointer(&testingType{})
		assert.False(t, notAPointer, "expected `false` when a pointer is provided")
		assert.NoError(t, err, "error should not be returned when a pointer is provided")
	})

	T.Run("with non-pointer", func(t *testing.T) {
		t.Parallel()
		notAPointer, err := argIsNotPointer(testingType{})
		assert.True(t, notAPointer, "expected `true` when a non-pointer is provided")
		assert.Error(t, err, "error should be returned when a non-pointer is provided")
	})

	T.Run("with nil", func(t *testing.T) {
		t.Parallel()
		notAPointer, err := argIsNotPointer(nil)
		assert.True(t, notAPointer, "expected `true` when nil is provided")
		assert.Error(t, err, "error should be returned when nil is provided")
	})
}

func TestArgIsNotNil(T *testing.T) {
	T.Parallel()

	T.Run("without nil", func(t *testing.T) {
		t.Parallel()
		isNil, err := argIsNotNil(&testingType{})
		assert.False(t, isNil, "expected `false` when a pointer is provided")
		assert.NoError(t, err, "error should not be returned when a pointer is provided")
	})

	T.Run("with non-pointer", func(t *testing.T) {
		t.Parallel()
		isNil, err := argIsNotNil(testingType{})
		assert.False(t, isNil, "expected `true` when a non-pointer is provided")
		assert.NoError(t, err, "error should not be returned when a non-pointer is provided")
	})

	T.Run("with nil", func(t *testing.T) {
		t.Parallel()
		isNil, err := argIsNotNil(nil)
		assert.True(t, isNil, "expected `true` when nil is provided")
		assert.Error(t, err, "error should be returned when nil is provided")
	})
}

func TestUnmarshalBody(T *testing.T) {
	T.Parallel()

	T.Run("expected use", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expected := "whatever"
		res := &http.Response{
			Body:       ioutil.NopCloser(strings.NewReader(fmt.Sprintf(`{"name": %q}`, expected))),
			StatusCode: http.StatusOK,
		}
		var out testingType

		err := unmarshalBody(ctx, res, &out)
		assert.Equal(t, out.Name, expected, "expected marshaling to work")
		assert.NoError(t, err, "no error should be encountered unmarshaling into a valid struct")
	})

	T.Run("with good status but unmarshallable response", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		res := &http.Response{
			Body:       ioutil.NopCloser(strings.NewReader("BLAH")),
			StatusCode: http.StatusOK,
		}
		var out testingType

		err := unmarshalBody(ctx, res, &out)
		assert.Error(t, err, "error should be encountered unmarshaling invalid response into a valid struct")
	})

	T.Run("with an erroneous error code", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		res := &http.Response{
			Body: ioutil.NopCloser(
				strings.NewReader(
					func() string {
						bs, err := json.Marshal(&types.ErrorResponse{})
						require.NoError(t, err)
						return string(bs)
					}(),
				),
			),
			StatusCode: http.StatusBadRequest,
		}
		var out *testingType

		err := unmarshalBody(ctx, res, &out)
		assert.Nil(t, out, "expected nil to be returned")
		assert.Error(t, err, "error should be returned from the API")
	})

	T.Run("with an erroneous error code and unmarshallable body", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		res := &http.Response{
			Body:       ioutil.NopCloser(strings.NewReader("BLAH")),
			StatusCode: http.StatusBadRequest,
		}
		var out *testingType

		err := unmarshalBody(ctx, res, &out)
		assert.Nil(t, out, "expected nil to be returned")
		assert.Error(t, err, "error should be returned from the unmarshaller")
	})

	T.Run("with nil target variable", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		err := unmarshalBody(ctx, nil, nil)
		assert.Error(t, err, "error should be encountered when passed nil")
	})

	T.Run("with erroneous reader", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expected := errors.New("blah")

		rc := newMockReadCloser()
		rc.On("Read", mock.AnythingOfType("[]uint8")).Return(0, expected)

		res := &http.Response{
			Body:       rc,
			StatusCode: http.StatusOK,
		}
		var out testingType

		err := unmarshalBody(ctx, res, &out)
		assert.Equal(t, expected, err)
		assert.Error(t, err, "no error should be encountered unmarshaling into a valid struct")

		mock.AssertExpectationsForObjects(t, rc)
	})
}

type testBreakableStruct struct {
	Thing json.Number `json:"thing"`
}

func TestCreateBodyFromStruct(T *testing.T) {
	T.Parallel()

	T.Run("expected use", func(t *testing.T) {
		t.Parallel()
		name := "whatever"
		expected := fmt.Sprintf(`{"name":%q}`, name)
		x := &testingType{Name: name}

		actual, err := createBodyFromStruct(x)
		assert.NoError(t, err, "expected no error creating JSON from valid struct")

		bs, err := ioutil.ReadAll(actual)
		assert.NoError(t, err, "expected no error reading JSON from valid struct")
		assert.Equal(t, expected, string(bs), "expected and actual JSON bodies don't match")
	})

	T.Run("with unmarshallable struct", func(t *testing.T) {
		t.Parallel()
		x := &testBreakableStruct{Thing: "stuff"}
		_, err := createBodyFromStruct(x)

		assert.Error(t, err, "expected error creating JSON from invalid struct")
	})
}
