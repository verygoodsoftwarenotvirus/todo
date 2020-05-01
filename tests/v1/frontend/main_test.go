package frontend

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tebeka/selenium"
)

func runTestOnAllSupportedBrowsers(t *testing.T, tp testProvider) {
	for _, bn := range []string{"firefox", "chrome"} {
		browserName := bn
		caps := selenium.Capabilities{"browserName": browserName}
		wd, err := selenium.NewRemote(caps, seleniumHubAddr)
		if err != nil {
			panic(err)
		}

		t.Run(bn, tp(wd))
		assert.NoError(t, wd.Quit())
	}
}

type testProvider func(driver selenium.WebDriver) func(t *testing.T)

func TestLoginPage(T *testing.T) {
	runTestOnAllSupportedBrowsers(T, func(driver selenium.WebDriver) func(t *testing.T) {
		return func(t *testing.T) {
			// Navigate to the login page.
			require.NoError(t, driver.Get(urlToUse+"/login"))

			// fetch the button.
			elem, err := driver.FindElement(selenium.ByID, "loginButton")
			if err != nil {
				panic(err)
			}

			// check that it is visible.
			actual, err := elem.IsDisplayed()
			assert.NoError(t, err)
			assert.True(t, actual)
		}
	})
}
