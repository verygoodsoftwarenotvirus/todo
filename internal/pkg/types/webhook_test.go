package types

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

func TestWebhook_Update(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		exampleInput := &WebhookUpdateInput{
			Name:        "whatever",
			ContentType: "application/xml",
			URL:         "https://blah.verygoodsoftwarenotvirus.ru",
			Method:      http.MethodPatch,
			Events:      []string{"more_things"},
			DataTypes:   []string{"new_stuff"},
			Topics:      []string{"blah-blah"},
		}

		actual := &Webhook{
			Name:        "something_else",
			ContentType: "application/json",
			URL:         "https://verygoodsoftwarenotvirus.ru",
			Method:      http.MethodPost,
			Events:      []string{"things"},
			DataTypes:   []string{"stuff"},
			Topics:      []string{"blah"},
		}
		expected := &Webhook{
			Name:        exampleInput.Name,
			ContentType: "application/xml",
			URL:         "https://blah.verygoodsoftwarenotvirus.ru",
			Method:      http.MethodPatch,
			Events:      []string{"more_things"},
			DataTypes:   []string{"new_stuff"},
			Topics:      []string{"blah-blah"},
		}

		actual.Update(exampleInput)
		assert.Equal(t, expected, actual)
	})
}

func TestWebhook_ToListener(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		w := &Webhook{}
		w.ToListener(noop.NewLogger())
	})
}

func Test_buildErrorLogFunc(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		w := &Webhook{}
		actual := buildErrorLogFunc(w, noop.NewLogger())
		actual(errors.New("blah"))
	})
}

func TestWebhookCreationInput_Validate(T *testing.T) {
	T.Parallel()

	buildValidWebhookCreationInput := func() *WebhookCreationInput {
		return &WebhookCreationInput{
			Name:        "whatever",
			ContentType: "application/xml",
			URL:         "https://blah.verygoodsoftwarenotvirus.ru",
			Method:      http.MethodPatch,
			Events:      []string{"more_things"},
			DataTypes:   []string{"new_stuff"},
			Topics:      []string{"blah-blah"},
		}
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, buildValidWebhookCreationInput().Validate())
	})

	T.Run("bad name", func(t *testing.T) {
		t.Parallel()
		exampleInput := buildValidWebhookCreationInput()
		exampleInput.Name = ""

		assert.Error(t, exampleInput.Validate())
	})

	T.Run("bad URL", func(t *testing.T) {
		t.Parallel()
		exampleInput := buildValidWebhookCreationInput()
		exampleInput.URL = fmt.Sprintf(`%s://verygoodsoftwarenotvirus.ru`, string(byte(127)))

		assert.Error(t, exampleInput.Validate())
	})

	T.Run("bad method", func(t *testing.T) {
		t.Parallel()
		exampleInput := buildValidWebhookCreationInput()
		exampleInput.Method = "balogna"

		assert.Error(t, exampleInput.Validate())
	})

	T.Run("bad content type", func(t *testing.T) {
		t.Parallel()
		exampleInput := buildValidWebhookCreationInput()
		exampleInput.ContentType = "application/balogna"

		assert.Error(t, exampleInput.Validate())
	})

	T.Run("empty events", func(t *testing.T) {
		t.Parallel()
		exampleInput := buildValidWebhookCreationInput()
		exampleInput.Events = []string{}

		assert.Error(t, exampleInput.Validate())
	})

	T.Run("empty data types", func(t *testing.T) {
		t.Parallel()
		exampleInput := buildValidWebhookCreationInput()
		exampleInput.DataTypes = []string{}

		assert.Error(t, exampleInput.Validate())
	})

	T.Run("empty topics", func(t *testing.T) {
		t.Parallel()
		exampleInput := buildValidWebhookCreationInput()
		exampleInput.Topics = []string{}

		assert.Error(t, exampleInput.Validate())
	})
}

func TestWebhookUpdateInput_Validate(T *testing.T) {
	T.Parallel()

	buildValidWebhookCreationInput := func() *WebhookUpdateInput {
		return &WebhookUpdateInput{
			Name:        "whatever",
			ContentType: "application/xml",
			URL:         "https://blah.verygoodsoftwarenotvirus.ru",
			Method:      http.MethodPatch,
			Events:      []string{"more_things"},
			DataTypes:   []string{"new_stuff"},
			Topics:      []string{"blah-blah"},
		}
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, buildValidWebhookCreationInput().Validate())
	})

	T.Run("bad name", func(t *testing.T) {
		t.Parallel()
		exampleInput := buildValidWebhookCreationInput()
		exampleInput.Name = ""

		assert.Error(t, exampleInput.Validate())
	})

	T.Run("bad URL", func(t *testing.T) {
		t.Parallel()
		exampleInput := buildValidWebhookCreationInput()
		exampleInput.URL = fmt.Sprintf(`%s://verygoodsoftwarenotvirus.ru`, string(byte(127)))

		assert.Error(t, exampleInput.Validate())
	})

	T.Run("bad method", func(t *testing.T) {
		t.Parallel()
		exampleInput := buildValidWebhookCreationInput()
		exampleInput.Method = "balogna"

		assert.Error(t, exampleInput.Validate())
	})

	T.Run("bad content type", func(t *testing.T) {
		t.Parallel()
		exampleInput := buildValidWebhookCreationInput()
		exampleInput.ContentType = "application/balogna"

		assert.Error(t, exampleInput.Validate())
	})

	T.Run("empty events", func(t *testing.T) {
		t.Parallel()
		exampleInput := buildValidWebhookCreationInput()
		exampleInput.Events = []string{}

		assert.Error(t, exampleInput.Validate())
	})

	T.Run("empty data types", func(t *testing.T) {
		t.Parallel()
		exampleInput := buildValidWebhookCreationInput()
		exampleInput.DataTypes = []string{}

		assert.Error(t, exampleInput.Validate())
	})

	T.Run("empty topics", func(t *testing.T) {
		t.Parallel()
		exampleInput := buildValidWebhookCreationInput()
		exampleInput.Topics = []string{}

		assert.Error(t, exampleInput.Validate())
	})
}
