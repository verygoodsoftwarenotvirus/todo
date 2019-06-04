package frontend

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tebeka/selenium"
)

func TestLoginPage(t *testing.T) {
	// Connect to the WebDriver instance running in docker-compose.
	caps := selenium.Capabilities{"browserName": "firefox"}
	wd, err := selenium.NewRemote(caps, seleniumHubAddr)
	if err != nil {
		panic(err)
	}
	defer wd.Quit()

	t.Log("hitting", urlToUse+"/login")

	// Navigate to the login page.
	if err := wd.Get(urlToUse + "/login"); err != nil {
		panic(err)
	}

	ps, err := wd.PageSource()
	t.Log(ps)
	require.NoError(t, err)

	// fetch the button.
	elem, err := wd.FindElement(selenium.ByID, "loginButton")
	if err != nil {
		panic(err)
	}

	// check that it is visible
	actual, err := elem.IsDisplayed()
	assert.NoError(t, err)
	assert.True(t, actual)
}
