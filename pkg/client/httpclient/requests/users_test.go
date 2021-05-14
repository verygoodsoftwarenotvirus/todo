package requests

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"net/http"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuilder_BuildGetUserRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/users/%d"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, h.exampleUser.ID)

		actual, err := h.builder.BuildGetUserRequest(h.ctx, h.exampleUser.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildGetUserRequest(h.ctx, 0)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildGetUsersRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/users"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

		actual, err := h.builder.BuildGetUsersRequest(h.ctx, nil)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})
}

func TestBuilder_BuildSearchForUsersByUsernameRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/users/search"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		exampleUsername := fakes.BuildFakeUser().Username
		spec := newRequestSpec(false, http.MethodGet, fmt.Sprintf("q=%s", exampleUsername), expectedPath)

		actual, err := h.builder.BuildSearchForUsersByUsernameRequest(h.ctx, exampleUsername)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with empty username", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildSearchForUsersByUsernameRequest(h.ctx, "")
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildCreateUserRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/users"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleInput := fakes.BuildFakeUserCreationInput()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

		actual, err := h.builder.BuildCreateUserRequest(h.ctx, exampleInput)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildCreateUserRequest(h.ctx, nil)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildArchiveUserRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/users/%d"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, h.exampleUser.ID)

		actual, err := h.builder.BuildArchiveUserRequest(h.ctx, h.exampleUser.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildArchiveUserRequest(h.ctx, 0)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

// buildArbitraryImage builds an image with a bunch of colors in it.
func buildArbitraryImage(widthAndHeight int) image.Image {
	img := image.NewRGBA(image.Rectangle{Min: image.Point{}, Max: image.Point{X: widthAndHeight, Y: widthAndHeight}})

	// Set color for each pixel.
	for x := 0; x < widthAndHeight; x++ {
		for y := 0; y < widthAndHeight; y++ {
			img.Set(x, y, color.RGBA{R: uint8(x % math.MaxUint8), G: uint8(y % math.MaxUint8), B: uint8(x + y%math.MaxUint8), A: math.MaxUint8})
		}
	}

	return img
}

func buildPNGBytes(t *testing.T, i image.Image) []byte {
	t.Helper()

	b := new(bytes.Buffer)
	require.NoError(t, png.Encode(b, i))

	return b.Bytes()
}

func TestBuilder_BuildAvatarUploadRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/users/avatar/upload"

	T.Run("standard jpeg", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)
		avatar := buildArbitraryImage(123)
		avatarBytes := buildPNGBytes(t, avatar)

		actual, err := h.builder.BuildAvatarUploadRequest(h.ctx, avatarBytes, "jpeg")
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("standard png", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)
		avatar := buildArbitraryImage(123)
		avatarBytes := buildPNGBytes(t, avatar)

		actual, err := h.builder.BuildAvatarUploadRequest(h.ctx, avatarBytes, "png")
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("standard gif", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)
		avatar := buildArbitraryImage(123)
		avatarBytes := buildPNGBytes(t, avatar)

		actual, err := h.builder.BuildAvatarUploadRequest(h.ctx, avatarBytes, "gif")
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with empty avatar", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildAvatarUploadRequest(h.ctx, nil, "jpeg")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid extension", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		avatar := buildArbitraryImage(123)
		avatarBytes := buildPNGBytes(t, avatar)

		actual, err := h.builder.BuildAvatarUploadRequest(h.ctx, avatarBytes, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error building request", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		h.builder = buildTestRequestBuilderWithInvalidURL()

		avatar := buildArbitraryImage(123)
		avatarBytes := buildPNGBytes(t, avatar)

		actual, err := h.builder.BuildAvatarUploadRequest(h.ctx, avatarBytes, "png")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildGetAuditLogForUserRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/users/%d/audit"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildGetAuditLogForUserRequest(h.ctx, h.exampleUser.ID)
		require.NotNil(t, actual)
		assert.NoError(t, err)

		spec := newRequestSpec(true, http.MethodGet, "", expectedPath, h.exampleUser.ID)
		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		actual, err := h.builder.BuildGetAuditLogForUserRequest(h.ctx, 0)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}
