package frontend

/*
import (
	"bytes"
	"image/png"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tebeka/selenium"
)

func saveScreenshotTo(t *testing.T, driver selenium.WebDriver, path string) {
	t.Helper()

	screenshotAsBytes, err := driver.Screenshot()
	require.NoError(t, err)

	im, err := png.Decode(bytes.NewReader(screenshotAsBytes))
	require.NoError(t, err)

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_CREATE, 0744)
	require.NoError(t, err)

	require.NoError(t, png.Encode(f, im))
}
*/
