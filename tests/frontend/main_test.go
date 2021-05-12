package frontend

import (
	"fmt"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestRegistrationFlow(t *testing.T) {
	t.Parallel()

	pw, err := playwright.Run()
	if err != nil {
		require.NoError(t, err, "could not start playwright")
	}
	browser, err := pw.Chromium.Launch()
	if err != nil {
		require.NoError(t, err, "could not launch browser")
	}
	page, err := browser.NewPage()
	if err != nil {
		require.NoError(t, err, "could not create page")
	}
	if _, err = page.Goto("https://news.ycombinator.com"); err != nil {
		require.NoError(t, err, "could not goto")
	}
	entries, err := page.QuerySelectorAll(".athing")
	if err != nil {
		require.NoError(t, err, "could not get entries")
	}
	for i, entry := range entries {
		titleElement, titleErr := entry.QuerySelector("td.title > a")
		if titleErr != nil {
			require.NoError(t, titleErr, "could not get title element")
		}

		title, titleContentErr := titleElement.TextContent()
		if titleContentErr != nil {
			require.NoError(t, titleContentErr, "could not get text content")
		}
		fmt.Printf("%d: %s\n", i+1, title)
	}
	if err = browser.Close(); err != nil {
		require.NoError(t, err, "could not close browser")
	}
	if err = pw.Stop(); err != nil {
		require.NoError(t, err, "could not stop Playwright")
	}
}
