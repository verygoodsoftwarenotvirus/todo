package frontend

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

var (
	chromeDisabled  bool
	firefoxDisabled bool
	webkitDisabled  bool
)

func init() {
	chromeDisabled = strings.EqualFold(strings.TrimSpace(os.Getenv("CHROME_DISABLED")), "y")
	firefoxDisabled = strings.EqualFold(strings.TrimSpace(os.Getenv("FIREFOX_DISABLED")), "y")
	webkitDisabled = strings.EqualFold(strings.TrimSpace(os.Getenv("WEBKIT_DISABLED")), "y")
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

	if !chromeDisabled {
		th.Chrome, err = pw.Chromium.Launch()
		require.NoError(t, err, "could not launch browser")
		require.NotNil(t, th.Chrome)
	}

	if !firefoxDisabled {
		th.Firefox, err = pw.Firefox.Launch()
		require.NoError(t, err, "could not launch browser")
		require.NotNil(t, th.Firefox)
	}

	if !webkitDisabled {
		th.Webkit, err = pw.WebKit.Launch()
		require.NoError(t, err, "could not launch browser")
		require.NotNil(t, th.Webkit)
	}

	return th
}

func (h *testHelper) runForAllBrowsers(t *testing.T, testName string, testFunc func(playwright.Browser) func(*testing.T)) {
	if !chromeDisabled {
		t.Run(fmt.Sprintf("%s with chrome", testName), testFunc(h.Chrome))
	}
	if !firefoxDisabled {
		t.Run(fmt.Sprintf("%s with firefox", testName), testFunc(h.Firefox))
	}
	if !webkitDisabled {
		t.Run(fmt.Sprintf("%s with webkit", testName), testFunc(h.Webkit))
	}
}

func boolPointer(b bool) *bool {
	return &b
}

const artifactsDir = "/relevant/absolute/path/should/go/here"

func saveScreenshotTo(t *testing.T, page playwright.Page, path string) {
	t.Helper()

	opts := playwright.PageScreenshotOptions{
		FullPage: boolPointer(true),
		Path:     stringPointer(filepath.Join(artifactsDir, fmt.Sprintf("%s.png", path))),
		Type:     playwright.ScreenshotTypePng,
	}

	_, err := page.Screenshot(opts)
	require.NoError(t, err)
}

func getScreenshotBytes(t *testing.T, ss playwright.ElementHandle) []byte {
	t.Helper()

	opts := playwright.ElementHandleScreenshotOptions{
		Type: playwright.ScreenshotTypePng,
	}

	data, err := ss.Screenshot(opts)
	require.NoError(t, err)

	return data
}
