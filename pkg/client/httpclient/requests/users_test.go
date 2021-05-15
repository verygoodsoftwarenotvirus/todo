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

		helper := buildTestHelper()

		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, helper.exampleUser.ID)

		actual, err := helper.builder.BuildGetUserRequest(helper.ctx, helper.exampleUser.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildGetUserRequest(helper.ctx, 0)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildGetUsersRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/users"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

		actual, err := helper.builder.BuildGetUsersRequest(helper.ctx, nil)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})
}

func TestBuilder_BuildSearchForUsersByUsernameRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/users/search"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		exampleUsername := fakes.BuildFakeUser().Username
		spec := newRequestSpec(false, http.MethodGet, fmt.Sprintf("q=%s", exampleUsername), expectedPath)

		actual, err := helper.builder.BuildSearchForUsersByUsernameRequest(helper.ctx, exampleUsername)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with empty username", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildSearchForUsersByUsernameRequest(helper.ctx, "")
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildCreateUserRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/users"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		exampleInput := fakes.BuildFakeUserCreationInput()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

		actual, err := helper.builder.BuildCreateUserRequest(helper.ctx, exampleInput)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildCreateUserRequest(helper.ctx, nil)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func TestBuilder_BuildArchiveUserRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/users/%d"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, helper.exampleUser.ID)

		actual, err := helper.builder.BuildArchiveUserRequest(helper.ctx, helper.exampleUser.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildArchiveUserRequest(helper.ctx, 0)
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

		helper := buildTestHelper()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)
		avatar := buildArbitraryImage(123)
		avatarBytes := buildPNGBytes(t, avatar)

		actual, err := helper.builder.BuildAvatarUploadRequest(helper.ctx, avatarBytes, "jpeg")
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("standard png", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)
		avatar := buildArbitraryImage(123)
		avatarBytes := buildPNGBytes(t, avatar)

		actual, err := helper.builder.BuildAvatarUploadRequest(helper.ctx, avatarBytes, "png")
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("standard gif", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)
		avatar := buildArbitraryImage(123)
		avatarBytes := buildPNGBytes(t, avatar)

		actual, err := helper.builder.BuildAvatarUploadRequest(helper.ctx, avatarBytes, "gif")
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})

	T.Run("with empty avatar", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildAvatarUploadRequest(helper.ctx, nil, "jpeg")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid extension", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		avatar := buildArbitraryImage(123)
		avatarBytes := buildPNGBytes(t, avatar)

		actual, err := helper.builder.BuildAvatarUploadRequest(helper.ctx, avatarBytes, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with error building request", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()
		helper.builder = buildTestRequestBuilderWithInvalidURL()

		avatar := buildArbitraryImage(123)
		avatarBytes := buildPNGBytes(t, avatar)

		actual, err := helper.builder.BuildAvatarUploadRequest(helper.ctx, avatarBytes, "png")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestBuilder_BuildGetAuditLogForUserRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/users/%d/audit"

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildGetAuditLogForUserRequest(helper.ctx, helper.exampleUser.ID)
		require.NotNil(t, actual)
		assert.NoError(t, err)

		spec := newRequestSpec(true, http.MethodGet, "", expectedPath, helper.exampleUser.ID)
		assertRequestQuality(t, actual, spec)
	})

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper()

		actual, err := helper.builder.BuildGetAuditLogForUserRequest(helper.ctx, 0)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}
