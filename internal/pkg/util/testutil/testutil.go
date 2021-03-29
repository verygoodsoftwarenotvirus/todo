package testutil

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	fake "github.com/brianvoe/gofakeit/v5"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/permissions"
)

const (
	base64ImagePrefix = `data:image/jpeg;base64,`
)

func init() {
	fake.Seed(time.Now().UnixNano())
}

// BuildMaxServiceAdminPerms builds a helpful ServiceAdminPermissionChecker.
func BuildMaxServiceAdminPerms() permissions.ServiceAdminPermissions {
	return permissions.NewServiceAdminPermissions(math.MaxUint32)
}

// BuildMaxUserPerms builds a helpful ServiceAdminPermissionChecker.
func BuildMaxUserPerms() permissions.ServiceUserPermissions {
	return permissions.NewServiceUserPermissions(math.MaxUint32)
}

// BuildNoAdminPerms builds a helpful ServiceAdminPermissionChecker.
func BuildNoAdminPerms() permissions.ServiceAdminPermissions {
	return permissions.NewServiceAdminPermissions(0)
}

// BuildNoUserPerms builds a helpful ServiceAdminPermissionChecker.
func BuildNoUserPerms() permissions.ServiceUserPermissions {
	return permissions.NewServiceUserPermissions(0)
}

// DetermineServiceURL returns the url, if properly configured.
func DetermineServiceURL() *url.URL {
	ta := os.Getenv("TARGET_ADDRESS")
	if ta == "" {
		panic("must provide target address!")
	}

	u, err := url.Parse(ta)
	if err != nil {
		panic(err)
	}

	svcAddr := u.String()

	log.Printf("using target address: %q\n", svcAddr)

	return u
}

// EnsureServerIsUp checks that a server is up and doesn't return until it's certain one way or the other.
func EnsureServerIsUp(ctx context.Context, address string) {
	var (
		isDown           = true
		interval         = time.Second
		maxAttempts      = 50
		numberOfAttempts = 0
	)

	for isDown {
		if !IsUp(ctx, address) {
			log.Printf("waiting %s before pinging %q again", interval, address)
			time.Sleep(interval)

			numberOfAttempts++
			if numberOfAttempts >= maxAttempts {
				log.Fatal("Maximum number of attempts made, something's gone awry")
			}
		} else {
			isDown = false
		}
	}
}

// IsUp can check if an instance of our server is alive.
func IsUp(ctx context.Context, address string) bool {
	uri := fmt.Sprintf("%s/_meta_/ready", address)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}

	if err = res.Body.Close(); err != nil {
		log.Println("error closing body")
	}

	return res.StatusCode == http.StatusOK
}

// ParseTwoFactorSecretFromBase64EncodedQRCode accepts a base64-encoded QR code representing an otpauth:// URI,
// parses the QR code and extracts the 2FA secret from the URI. It can also return an error.
func ParseTwoFactorSecretFromBase64EncodedQRCode(qrCode string) (string, error) {
	qrCode = strings.TrimPrefix(qrCode, base64ImagePrefix)

	unbased, err := base64.StdEncoding.DecodeString(qrCode)
	if err != nil {
		return "", fmt.Errorf("Cannot decode b64: %w", err)
	}

	im, err := png.Decode(bytes.NewReader(unbased))
	if err != nil {
		return "", fmt.Errorf("Bad png: %w", err)
	}

	bb, err := gozxing.NewBinaryBitmapFromImage(im)
	if err != nil {
		return "", fmt.Errorf("Bad binary bitmap: %w", err)
	}

	res, err := qrcode.NewQRCodeReader().DecodeWithoutHints(bb)
	if err != nil {
		return "", fmt.Errorf("decoding: %w", err)
	}

	totpDetails := res.String()

	u, err := url.Parse(totpDetails)
	if err != nil {
		return "", fmt.Errorf("parsing URI: %w", err)
	}

	return u.Query().Get("secret"), nil
}

// CreateBodyFromStruct takes any value in and returns an io.ReadCloser for an http.Request's body.
func CreateBodyFromStruct(t *testing.T, in interface{}) io.ReadCloser {
	t.Helper()

	out, err := json.Marshal(in)
	require.NoError(t, err)

	return ioutil.NopCloser(bytes.NewReader(out))
}

// BuildArbitraryImage builds an image with a bunch of colors in it.
func BuildArbitraryImage(widthAndHeight int) image.Image {
	img := image.NewRGBA(image.Rectangle{Min: image.Point{}, Max: image.Point{X: widthAndHeight, Y: widthAndHeight}})

	// Set color for each pixel.
	for x := 0; x < widthAndHeight; x++ {
		for y := 0; y < widthAndHeight; y++ {
			img.Set(x, y, color.RGBA{R: uint8(x % math.MaxUint8), G: uint8(y % math.MaxUint8), B: uint8(x + y%math.MaxUint8), A: math.MaxUint8})
		}
	}

	return img
}

// BuildArbitraryImagePNGBytes builds an image with a bunch of colors in it.
func BuildArbitraryImagePNGBytes(widthAndHeight int) []byte {
	var b bytes.Buffer
	if err := png.Encode(&b, BuildArbitraryImage(widthAndHeight)); err != nil {
		panic(err)
	}

	return b.Bytes()
}

// AssertAppropriateNumberOfTestsRan ensures the expected number of tests are run in a given suite.
func AssertAppropriateNumberOfTestsRan(t *testing.T, totalExpectedTestCount uint, stats *suite.SuiteInformation) {
	t.Helper()

	/*
		Acknowledged that this:
			1. a corny thing to do
			2. an annoying thing to have to update when you add new tests
			3. the source of a false negative when debugging a singular test

		That said, in the event someone boo-boos and leaves something in globalClientExceptions, this part will fail,
		which is worth it.
	*/

	if stats.Passed() {
		require.Equal(t, int(totalExpectedTestCount), len(stats.TestStats), "expected total number of tests run to equal %d, but it was %d", totalExpectedTestCount, len(stats.TestStats))
	}
}
