package integration

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"net/http"
	"reflect"
	"strconv"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/bxcodec/faker"
	"github.com/stretchr/testify/assert"
)

func init() {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
}

// randString produces a random string
// https://blog.questionable.services/article/generating-secure-random-numbers-crypto-rand/
func randString() (string, error) {
	b := make([]byte, 64)
	// Note that err == nil only if we read len(b) bytes.
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base32.StdEncoding.EncodeToString(b), nil
}

func buildDummyUserInput(t *testing.T) *models.UserInput {
	t.Helper()

	u, _ := (&faker.Internet{}).UserName(reflect.ValueOf(nil))
	p, _ := (&faker.Internet{}).Password(reflect.ValueOf(nil))
	tfs, _ := randString()

	x := &models.UserInput{
		Username:        u.(string),
		Password:        p.(string),
		TwoFactorSecret: tfs,
	}

	return x
}

func buildDummyUser(ctx context.Context, t *testing.T) (*models.UserCreationResponse, *models.UserInput, *http.Cookie) {
	t.Helper()

	// build user creation route input
	y := buildDummyUserInput(t)
	u, err := todoClient.CreateUser(ctx, y)
	assert.NotNil(t, u)
	// require.NoError(t, err)

	t.Logf("created dummy user #%d", u.ID)
	// code, err := totp.GenerateCode(u.TwoFactorSecret, time.Now())
	// assert.NoError(t, err)
	// t.Logf("created 2FA code %q", code)
	// require.NotEmpty(t, code)

	cookie := loginUser(t, u.Username, y.Password, u.TwoFactorSecret)

	// cookie, err := todoClient.Login(ctx, u.Username, y.Password, code)
	t.Logf("received cookie: %v", cookie != nil)
	t.Logf("received error: %v", err != nil)

	assert.NoError(t, err)
	assert.NotNil(t, cookie)

	return u, y, cookie
}

func checkUserCreationEquality(t *testing.T, expected *models.UserInput, actual *models.UserCreationResponse) {
	t.Helper()

	assert.NotZero(t, actual.ID)
	assert.Equal(t, expected.Username, actual.Username)
	assert.NotEmpty(t, actual.TwoFactorSecret)
	// assert.Nil(t, actual.PasswordLastChangedOn)
	assert.NotZero(t, actual.CreatedOn)
	assert.Nil(t, actual.UpdatedOn)
	assert.Nil(t, actual.ArchivedOn)
}

func checkUserEquality(t *testing.T, expected *models.UserInput, actual *models.User) {
	t.Helper()

	assert.NotZero(t, actual.ID)
	assert.Equal(t, expected.Username, actual.Username)
	assert.Empty(t, actual.TwoFactorSecret)
	// assert.Nil(t, actual.PasswordLastChangedOn)
	assert.NotZero(t, actual.CreatedOn)
	assert.Nil(t, actual.UpdatedOn)
	assert.Nil(t, actual.ArchivedOn)
}

func TestUsers(test *testing.T) {
	// test.Parallel()

	test.Run("Creating", func(T *testing.T) {
		T.Run("should be creatable", func(t *testing.T) {
			tctx := buildSpanContext("create-user")

			// Create user
			expected := buildDummyUserInput(t)
			actual, err := todoClient.CreateUser(
				tctx,
				&models.UserInput{
					Username: expected.Username,
					Password: expected.Password,
				},
			)
			checkValueAndError(t, actual, err)

			// Assert user equality
			checkUserCreationEquality(t, expected, actual)

			// Clean up
			assert.NoError(t, todoClient.DeleteUser(tctx, strconv.FormatUint(actual.ID, 10)))
		})
	})

	test.Run("Reading", func(T *testing.T) {
		T.Run("it should return an error when trying to read something that doesn't exist", func(t *testing.T) {
			tctx := buildSpanContext("search-for-nonexistent-user")

			// Fetch user
			actual, err := todoClient.GetUser(tctx, "nonexistent")
			assert.Nil(t, actual)
			assert.Error(t, err)
		})

		T.Run("it should be readable", func(t *testing.T) {
			tctx := buildSpanContext("read-user")

			// Create user
			expected := buildDummyUserInput(t)
			premade, err := todoClient.CreateUser(
				tctx,
				&models.UserInput{
					Username: expected.Username,
					Password: expected.Password,
				},
			)
			checkValueAndError(t, premade, err)
			assert.NotEmpty(t, premade.TwoFactorSecret)

			// Fetch user
			actual, err := todoClient.GetUser(tctx, premade.Username)
			checkValueAndError(t, actual, err)

			// Assert user equality
			checkUserEquality(t, expected, actual)

			// Clean up
			assert.NoError(t, todoClient.DeleteUser(tctx, actual.Username))
		})
	})

	test.Run("Updating", func(T *testing.T) {
		T.Run("it should be updatable", func(t *testing.T) {
			t.SkipNow()
		})
	})

	test.Run("Deleting", func(T *testing.T) {
		T.Run("should be able to be deleted", func(t *testing.T) {
			tctx := buildSpanContext("delete-user")

			// Create user
			y := buildDummyUserInput(t)
			u, err := todoClient.CreateUser(tctx, y)
			assert.NoError(t, err)
			assert.NotNil(t, u)

			if u == nil || err != nil {
				t.Log("TestUsers something has gone awry, user returned is nil")
				t.FailNow()
			}

			// Execute
			err = todoClient.DeleteUser(tctx, strconv.FormatUint(u.ID, 10))
			assert.NoError(t, err)
		})
	})

	test.Run("Listing", func(T *testing.T) {
		T.Run("should be able to be read in a list", func(t *testing.T) {
			tctx := buildSpanContext("list-users")

			// Create users
			expected := []*models.UserCreationResponse{}
			for i := 0; i < 5; i++ {
				user, _, c := buildDummyUser(tctx, t)
				assert.NotNil(t, c)
				expected = append(expected, user)
			}

			// Assert user list equality
			actual, err := todoClient.GetUsers(tctx, nil)
			checkValueAndError(t, actual, err)
			assert.True(t, len(expected) <= len(actual.Users))

			// Clean up
			for _, user := range actual.Users {
				err := todoClient.DeleteUser(tctx, strconv.FormatUint(user.ID, 10))
				assert.NoError(t, err)
			}
		})
	})

	test.Run("Counting", func(T *testing.T) {
		T.Run("it should be able to be counted", func(t *testing.T) {
			t.Skip()
		})
	})
}
