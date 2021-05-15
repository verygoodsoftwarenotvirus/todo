package testutils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"net/http"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/permissions"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	fake "github.com/brianvoe/gofakeit/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func init() {
	fake.Seed(time.Now().UnixNano())
}

// BuildMaxServiceAdminPerms builds a helpful ServiceAdminPermissionChecker.
func BuildMaxServiceAdminPerms() permissions.ServiceAdminPermission {
	return permissions.NewServiceAdminPermissions(math.MaxInt64)
}

// BuildMaxUserPerms builds a helpful ServiceAdminPermissionChecker.
func BuildMaxUserPerms() permissions.ServiceUserPermission {
	return permissions.NewServiceUserPermissions(math.MaxInt64)
}

// BuildNoAdminPerms builds a helpful ServiceAdminPermissionChecker.
func BuildNoAdminPerms() permissions.ServiceAdminPermission {
	return permissions.NewServiceAdminPermissions(0)
}

// BuildNoUserPerms builds a helpful ServiceAdminPermissionChecker.
func BuildNoUserPerms() permissions.ServiceUserPermission {
	return permissions.NewServiceUserPermissions(0)
}

// errArbitrary is an arbitrary error.
var errArbitrary = errors.New("blah")

// BrokenSessionContextDataFetcher is a deliberately broken sessionContextDataFetcher.
func BrokenSessionContextDataFetcher(*http.Request) (*types.SessionContextData, error) {
	return nil, errArbitrary
}

// CreateBodyFromStruct takes any value in and returns an io.ReadCloser for an http.Request's body.
func CreateBodyFromStruct(t *testing.T, in interface{}) io.ReadCloser {
	t.Helper()

	out, err := json.Marshal(in)
	require.NoError(t, err)

	return io.NopCloser(bytes.NewReader(out))
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

// BuildTestRequest builds an arbitrary *http.Request.
func BuildTestRequest(t *testing.T) *http.Request {
	t.Helper()

	ctx := context.Background()
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodOptions,
		"https://todo.verygoodsoftwarenotvirus.ru",
		nil,
	)

	require.NotNil(t, req)
	assert.NoError(t, err)

	return req
}
