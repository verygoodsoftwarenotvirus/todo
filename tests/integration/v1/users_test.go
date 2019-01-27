package integration

import (
	"net/http"
	"reflect"
	"strconv"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/bxcodec/faker"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
)

func buildDummyUserInput(t *testing.T) *models.UserInput {
	t.Helper()

	u, _ := (&faker.Internet{}).UserName(reflect.ValueOf(nil))
	p, _ := (&faker.Internet{}).Password(reflect.ValueOf(nil))

	x := &models.UserInput{
		Username: u.(string),
		Password: p.(string),
	}

	return x
}

func buildDummyUser(t *testing.T) (*models.UserCreationResponse, *models.UserInput, *http.Cookie) {
	t.Helper()

	// build user creation route input
	y := buildDummyUserInput(t)
	x, err := todoClient.CreateUser(y)
	assert.NoError(t, err)

	code, err := totp.GenerateCode(x.TwoFactorSecret, time.Now())
	assert.NoError(t, err)
	cookie, err := todoClient.Login(x.Username, y.Password, code)
	assert.NoError(t, err)

	return x, y, cookie
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
			// Create user
			expected := buildDummyUserInput(t)
			actual, err := todoClient.CreateUser(
				&models.UserInput{
					Username: expected.Username,
					Password: expected.Password,
				},
			)
			checkValueAndError(t, actual, err)

			// Assert user equality
			checkUserCreationEquality(t, expected, actual)

			// Clean up
			assert.NoError(t, todoClient.DeleteUser(strconv.FormatUint(actual.ID, 10)))
		})
	})

	test.Run("Reading", func(T *testing.T) {
		T.Run("it should return an error when trying to read something that doesn't exist", func(t *testing.T) {
			// Fetch user
			actual, err := todoClient.GetUser("nonexistent")
			assert.Nil(t, actual)
			assert.Error(t, err)
		})

		T.Run("it should be readable", func(t *testing.T) {
			// Create user
			expected := buildDummyUserInput(t)
			premade, err := todoClient.CreateUser(
				&models.UserInput{
					Username: expected.Username,
					Password: expected.Password,
				},
			)
			checkValueAndError(t, premade, err)
			assert.NotEmpty(t, premade.TwoFactorSecret)

			// Fetch user
			actual, err := todoClient.GetUser(premade.Username)
			checkValueAndError(t, actual, err)

			// Assert user equality
			checkUserEquality(t, expected, actual)

			// Clean up
			assert.NoError(t, todoClient.DeleteUser(actual.Username))
		})
	})

	test.Run("Updating", func(T *testing.T) {
		T.Run("it should be updatable", func(t *testing.T) {
			t.SkipNow()
		})
	})

	test.Run("Deleting", func(T *testing.T) {
		T.Run("should be able to be deleted", func(t *testing.T) {
			// Create user
			premade, _, c := buildDummyUser(t)
			if c == nil {
				t.Log("TestUsers deleted test cookie is nil")
			}

			// Clean up
			err := todoClient.DeleteUser(strconv.FormatUint(premade.ID, 10))
			assert.NoError(t, err)
		})
	})

	test.Run("Listing", func(T *testing.T) {
		T.Run("should be able to be read in a list", func(t *testing.T) {
			// Create users
			expected := []*models.UserCreationResponse{}
			for i := 0; i < 5; i++ {
				user, _, c := buildDummyUser(t)
				assert.NotNil(t, c)
				expected = append(expected, user)
			}

			// Assert user list equality
			actual, err := todoClient.GetUsers(nil)
			checkValueAndError(t, actual, err)
			assert.True(t, len(expected) <= len(actual.Users))

			// Clean up
			for _, user := range actual.Users {
				err := todoClient.DeleteUser(strconv.FormatUint(user.ID, 10))
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
