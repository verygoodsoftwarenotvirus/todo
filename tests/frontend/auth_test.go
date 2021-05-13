package frontend

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"testing"
	"time"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestRegistrationFlow(T *testing.T) {
	helper := setupTestHelper(T)

	helper.runForAllBrowsers(T, "registration flow", func(browser playwright.Browser) func(*testing.T) {
		return func(t *testing.T) {
			user := fakes.BuildFakeUserCreationInput()

			page, err := browser.NewPage()
			require.NoError(t, err, "could not create page")

			_, err = page.Goto(urlToUse)
			require.NoError(t, err, "could not navigate to root page")

			registerLink, err := page.QuerySelector("#registerLink")
			require.NoError(t, err, "could not find register link on homepage")

			require.NoError(t, registerLink.Click(), "error clicking registration link")

			// fetch the username field and fill it
			usernameField, usernameFieldFindErr := page.QuerySelector("#usernameInput")
			require.NoError(t, usernameFieldFindErr, "could not find username input on registration page")

			// fetch the passwords field and fill it
			passwordField, passwordFieldFindErr := page.QuerySelector("#passwordInput")
			require.NoError(t, passwordFieldFindErr, "could not find password input on registration page")

			time.Sleep(time.Second)

			assert.Equal(t, urlToUse+"/register", page.URL())

			require.NoError(t, page.Type("#usernameInput", user.Username))
			require.NoError(t, page.Type("#passwordInput", user.Password))
			require.NoError(t, page.Click("#registrationButton"))

			time.Sleep(time.Second)

			saveScreenshotTo(t, page, "fart")

			require.NotNil(t, usernameField)
			require.NotNil(t, passwordField)
		}
	})
}
