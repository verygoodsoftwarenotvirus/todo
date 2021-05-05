package frontend

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_detailsForLanguage(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		for _, lang := range supportedLanguages {
			assert.NotNil(t, detailsForLanguage(lang))
		}
	})

	T.Run("returns default for nil", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, englishDetails, detailsForLanguage(nil))
	})
}

func Test_determineLanguage(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		for expected, deets := range languageDetails {
			req := httptest.NewRequest(http.MethodGet, "/things", nil)
			req.Header.Set("Accept-Language", deets.Abbreviation)

			actual := determineLanguage(req)
			assert.Equal(t, expected, actual, "expected result to be interpreted as English")
		}
	})

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/things", nil)
		req.Header.Set("Accept-Language", "en-US")

		actual := determineLanguage(req)
		assert.Equal(t, english, actual, "expected result to be interpreted as English")
	})

	T.Run("returns default for nil", func(t *testing.T) {
		t.Parallel()

		actual := determineLanguage(nil)
		assert.Equal(t, defaultLanguage, actual, "expected result to be interpreted as English")
	})

	T.Run("returns default for invalid language header", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/things", nil)
		req.Header.Set("Accept-Language", "")

		actual := determineLanguage(req)
		assert.Equal(t, defaultLanguage, actual, "expected result to be interpreted as English")
	})

	T.Run("returns default for language header that yields no results but does not error", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/things", nil)
		req.Header.Set("Accept-Language", "fleeb-FLORP")

		actual := determineLanguage(req)
		assert.Equal(t, defaultLanguage, actual, "expected result to be interpreted as English")
	})

	T.Run("returns default for language not found", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/things", nil)
		req.Header.Set("Accept-Language", "zu-HM")

		actual := determineLanguage(req)
		assert.Equal(t, defaultLanguage, actual, "expected result to be interpreted as English")
	})
}
