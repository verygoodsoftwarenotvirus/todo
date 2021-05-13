package frontend

import (
	"fmt"
	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var (
	ChromeDisabled  bool
	FirefoxDisabled bool
	WebkitDisabled  bool
)

func init() {
	ChromeDisabled = true // strings.ToLower(strings.TrimSpace(os.Getenv("CHROME_DISABLED"))) == "y"
	FirefoxDisabled = strings.ToLower(strings.TrimSpace(os.Getenv("FIREFOX_DISABLED"))) == "y"
	WebkitDisabled = true // strings.ToLower(strings.TrimSpace(os.Getenv("WEBKIT_DISABLED"))) == "y"
}

func stringPointer(s string) *string {
	return &s
}

type testHelper struct {
	pw                      *playwright.Playwright
	Firefox, Chrome, Webkit playwright.Browser
}

func setupTestHelper(t *testing.T) *testHelper {
	t.Helper()

	pw, err := playwright.Run()
	require.NoError(t, err, "could not start playwright")

	th := &testHelper{pw: pw}

	if !ChromeDisabled {
		th.Chrome, err = pw.Chromium.Launch()
		require.NotNil(t, th.Chrome)
		require.NoError(t, err, "could not launch browser")
	}

	if !FirefoxDisabled {
		th.Firefox, err = pw.Firefox.Launch()
		require.NotNil(t, th.Firefox)
		require.NoError(t, err, "could not launch browser")
	}

	if !WebkitDisabled {
		th.Webkit, err = pw.WebKit.Launch()
		require.NotNil(t, th.Webkit)
		require.NoError(t, err, "could not launch browser")
	}

	return th
}

func (h *testHelper) runForAllBrowsers(t *testing.T, testName string, testFunc func(playwright.Browser) func(*testing.T)) {
	if !ChromeDisabled {
		t.Run(fmt.Sprintf("%s with chrome", testName), testFunc(h.Chrome))
	}
	if !FirefoxDisabled {
		t.Run(fmt.Sprintf("%s with firefox", testName), testFunc(h.Firefox))
	}
	if !WebkitDisabled {
		t.Run(fmt.Sprintf("%s with webkit", testName), testFunc(h.Webkit))
	}
}

func boolPointer(b bool) *bool {
	return &b
}

func saveScreenshotTo(t *testing.T, page playwright.Page, path string) {
	t.Helper()

	opts := playwright.PageScreenshotOptions{
		FullPage: boolPointer(true),
		Path:     stringPointer(filepath.Join("/home/vgsnv/src/gitlab.com/verygoodsoftwarenotvirus/todo/artifacts", fmt.Sprintf("%s.png", path))),
		Type:     playwright.ScreenshotTypePng,
	}

	_, err := page.Screenshot(opts)
	require.NoError(t, err)
}
