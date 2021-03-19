package requests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testingType struct {
	Name string `json:"name"`
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
