package models

import (
	"errors"
	"net/http"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/logging/v1/noop"

	"github.com/stretchr/testify/assert"
)

func TestWebhook_Update(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		actual := &Webhook{
			Name:        "name",
			ContentType: "application/json",
			URL:         "https://verygoodsoftwarenotvirus.ru",
			Method:      http.MethodPost,
			Events:      []string{"things"},
			DataTypes:   []string{"stuff"},
			Topics:      []string{"blah"},
		}

		exampleInput := &WebhookInput{
			Name:        "new name",
			ContentType: "application/xml",
			URL:         "https://blah.verygoodsoftwarenotvirus.ru",
			Method:      http.MethodPatch,
			Events:      []string{"more_things"},
			DataTypes:   []string{"new_stuff"},
			Topics:      []string{"blah-blah"},
		}

		expected := &Webhook{
			Name:        "new name",
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
		w := &Webhook{}
		w.ToListener(noop.ProvideNoopLogger())

	})
}

func Test_buildErrorLogFunc(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		w := &Webhook{}

		actual := buildErrorLogFunc(w, noop.ProvideNoopLogger())
		actual(errors.New("blah"))
	})
}
