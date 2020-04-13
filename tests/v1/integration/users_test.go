package integration

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"net/http"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/tracing"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	fakemodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/fake"

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
	// Note that err == nil only if we read len(b) bytes
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base32.StdEncoding.EncodeToString(b), nil
}

func buildDummyUser(t *testing.T) (*models.UserCreationResponse, *models.UserCreationInput, *http.Cookie) {
	t.Helper()
	ctx := context.Background()

	// build user creation route input
	userInput := fakemodels.BuildFakeUserCreationInput()
	user, err := todoClient.CreateUser(ctx, userInput)
	assert.NotNil(t, user)
	require.NoError(t, err)

	if user == nil || err != nil {
		t.FailNow()
	}
	cookie := loginUser(t, userInput.Username, userInput.Password, user.TwoFactorSecret)

	require.NoError(t, err)
	require.NotNil(t, cookie)

	return user, userInput, cookie
}

func checkUserCreationEquality(t *testing.T, expected *models.UserCreationInput, actual *models.UserCreationResponse) {
	t.Helper()

	assert.NotZero(t, actual.ID)
	assert.Equal(t, expected.Username, actual.Username)
	assert.NotEmpty(t, actual.TwoFactorSecret)
	assert.NotZero(t, actual.CreatedOn)
	assert.Nil(t, actual.UpdatedOn)
	assert.Nil(t, actual.ArchivedOn)
}

func checkUserEquality(t *testing.T, expected *models.UserCreationInput, actual *models.User) {
	t.Helper()

	assert.NotZero(t, actual.ID)
	assert.Equal(t, expected.Username, actual.Username)
	assert.NotZero(t, actual.CreatedOn)
	assert.Nil(t, actual.UpdatedOn)
	assert.Nil(t, actual.ArchivedOn)
}

func TestUsers(test *testing.T) {
	test.Parallel()

	test.Run("Creating", func(T *testing.T) {
		T.Run("should be creatable", func(t *testing.T) {
			ctx := context.Background()
			ctx, span := tracing.StartSpan(ctx, t.Name())
			defer span.End()

			// Create user
			exampleUserInput := fakemodels.BuildFakeUserCreationInput()
			actual, err := todoClient.CreateUser(ctx, exampleUserInput)
			checkValueAndError(t, actual, err)

			// Assert user equality
			checkUserCreationEquality(t, exampleUserInput, actual)

			// Clean up
			assert.NoError(t, todoClient.ArchiveUser(ctx, actual.ID))
		})
	})

	test.Run("Reading", func(T *testing.T) {
		T.Run("it should return an error when trying to read something that doesn't exist", func(t *testing.T) {
			ctx := context.Background()
			ctx, span := tracing.StartSpan(ctx, t.Name())
			defer span.End()

			// Fetch user
			actual, err := todoClient.GetUser(ctx, nonexistentID)
			assert.Nil(t, actual)
			assert.Error(t, err)
		})

		T.Run("it should be readable", func(t *testing.T) {
			ctx := context.Background()
			ctx, span := tracing.StartSpan(ctx, t.Name())
			defer span.End()

			// Create user
			exampleUserInput := fakemodels.BuildFakeUserCreationInput()
			premade, err := todoClient.CreateUser(ctx, exampleUserInput)
			checkValueAndError(t, premade, err)
			assert.NotEmpty(t, premade.TwoFactorSecret)

			// Fetch user
			actual, err := todoClient.GetUser(ctx, premade.ID)
			if err != nil {
				t.Logf("error encountered trying to fetch user %q: %v\n", premade.Username, err)
			}
			checkValueAndError(t, actual, err)

			// Assert user equality
			checkUserEquality(t, exampleUserInput, actual)

			// Clean up
			assert.NoError(t, todoClient.ArchiveUser(ctx, actual.ID))
		})
	})

	test.Run("Deleting", func(T *testing.T) {
		T.Run("should be able to be deleted", func(t *testing.T) {
			ctx := context.Background()
			ctx, span := tracing.StartSpan(ctx, t.Name())
			defer span.End()

			// Create user
			exampleUserInput := fakemodels.BuildFakeUserCreationInput()
			u, err := todoClient.CreateUser(ctx, exampleUserInput)
			assert.NoError(t, err)
			assert.NotNil(t, u)

			if u == nil || err != nil {
				t.Log("something has gone awry, user returned is nil")
				t.FailNow()
			}

			// Execute
			err = todoClient.ArchiveUser(ctx, u.ID)
			assert.NoError(t, err)
		})
	})

	test.Run("Listing", func(T *testing.T) {
		T.Run("should be able to be read in a list", func(t *testing.T) {
			ctx := context.Background()
			ctx, span := tracing.StartSpan(ctx, t.Name())
			defer span.End()

			// Create users
			var expected []*models.UserCreationResponse
			for i := 0; i < 5; i++ {
				user, _, c := buildDummyUser(t)
				assert.NotNil(t, c)
				expected = append(expected, user)
			}

			// Assert user list equality
			actual, err := todoClient.GetUsers(ctx, nil)
			checkValueAndError(t, actual, err)
			assert.True(t, len(expected) <= len(actual.Users))

			// Clean up
			for _, user := range actual.Users {
				err = todoClient.ArchiveUser(ctx, user.ID)
				assert.NoError(t, err)
			}
		})
	})
}
