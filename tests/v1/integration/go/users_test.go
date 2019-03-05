package integration

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"net/http"
	"strconv"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/icrowley/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	tfs, err := randString()
	if err != nil {
		panic(err)
	}

	fake.Seed(time.Now().UnixNano())

	userInput := &models.UserInput{
		Username:        fake.UserName(),
		Password:        fake.Password(8, 64, true, true, true),
		TwoFactorSecret: tfs,
	}

	return userInput
}

func buildDummyUser(t *testing.T) (*models.UserCreationResponse, *models.UserInput, *http.Cookie) {
	t.Helper()
	ctx := context.Background()

	// build user creation route input
	userInput := buildDummyUserInput(t)
	user, err := todoClient.CreateUser(ctx, userInput)
	assert.NotNil(t, user)
	require.NoError(t, err)
	t.Logf("created dummy user #%d: %q", user.ID, user.Username)

	if user == nil || err != nil {
		t.FailNow()
	}

	cookie := loginUser(t, userInput.Username, userInput.Password, user.TwoFactorSecret)

	require.NoError(t, err)
	require.NotNil(t, cookie)

	return user, userInput, cookie
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
			tctx := context.Background()

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
			tctx := context.Background()

			// Fetch user
			actual, err := todoClient.GetUser(tctx, "nonexistent")
			assert.Nil(t, actual)
			assert.Error(t, err)
		})

		T.Run("it should be readable", func(t *testing.T) {
			tctx := context.Background()

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
			if err != nil {
				t.Logf("error encountered trying to fetch user %q: %v\n", premade.Username, err)
			}

			checkValueAndError(t, actual, err)

			// Assert user equality
			checkUserEquality(t, expected, actual)

			// Clean up
			assert.NoError(t, todoClient.DeleteUser(tctx, actual.Username))
		})
	})

	test.Run("Deleting", func(T *testing.T) {
		T.Run("should be able to be deleted", func(t *testing.T) {
			tctx := context.Background()

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
			tctx := context.Background()

			// Create users
			var expected []*models.UserCreationResponse
			for i := 0; i < 5; i++ {
				user, _, c := buildDummyUser(t)
				assert.NotNil(t, c)
				expected = append(expected, user)
			}

			// Assert user list equality
			actual, err := todoClient.GetUsers(tctx, nil)
			checkValueAndError(t, actual, err)
			assert.True(t, len(expected) <= len(actual.Users))

			// Clean up
			for _, user := range actual.Users {
				err = todoClient.DeleteUser(tctx, strconv.FormatUint(user.ID, 10))
				assert.NoError(t, err)
			}
		})
	})

}
