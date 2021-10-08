package types

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWebhookCreationInput_Validate(T *testing.T) {
	T.Parallel()

	buildValidWebhookCreationInput := func() *WebhookCreationRequestInput {
		return &WebhookCreationRequestInput{
			Name:        "whatever",
			ContentType: "application/xml",
			URL:         "https://blah.verygoodsoftwarenotvirus.ru",
			Method:      http.MethodPatch,
			Events:      []string{"more_things"},
			DataTypes:   []string{"new_stuff"},
			Topics:      []string{"blah-blah"},
		}
	}

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, buildValidWebhookCreationInput().ValidateWithContext(context.Background()))
	})

	T.Run("bad name", func(t *testing.T) {
		t.Parallel()
		exampleInput := buildValidWebhookCreationInput()
		exampleInput.Name = ""

		assert.Error(t, exampleInput.ValidateWithContext(context.Background()))
	})

	T.Run("bad url", func(t *testing.T) {
		t.Parallel()
		exampleInput := buildValidWebhookCreationInput()
		// much as we'd like to use testutils.InvalidRawURL here, it causes a cyclical import :'(
		exampleInput.URL = fmt.Sprintf(`%s://verygoodsoftwarenotvirus.ru`, string(byte(127)))

		assert.Error(t, exampleInput.ValidateWithContext(context.Background()))
	})

	T.Run("bad method", func(t *testing.T) {
		t.Parallel()
		exampleInput := buildValidWebhookCreationInput()
		exampleInput.Method = "balogna"

		assert.Error(t, exampleInput.ValidateWithContext(context.Background()))
	})

	T.Run("bad content type", func(t *testing.T) {
		t.Parallel()
		exampleInput := buildValidWebhookCreationInput()
		exampleInput.ContentType = "application/balogna"

		assert.Error(t, exampleInput.ValidateWithContext(context.Background()))
	})

	T.Run("empty events", func(t *testing.T) {
		t.Parallel()
		exampleInput := buildValidWebhookCreationInput()
		exampleInput.Events = []string{}

		assert.Error(t, exampleInput.ValidateWithContext(context.Background()))
	})

	T.Run("empty data types", func(t *testing.T) {
		t.Parallel()
		exampleInput := buildValidWebhookCreationInput()
		exampleInput.DataTypes = []string{}

		assert.Error(t, exampleInput.ValidateWithContext(context.Background()))
	})
}
