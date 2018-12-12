package integration

import (
	"net/http"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/bxcodec/faker"
	"github.com/franela/goblin"
	. "github.com/onsi/gomega"
)

func buildDummyUserInput(t *testing.T) *models.UserInput {
	t.Helper()

	x := &models.UserInput{
		Username: faker.Internet{}.UserName(),
		Password: faker.Internet{}.Password(),
	}

	return x
}

func buildDummyUser(t *testing.T) (*models.User, *models.UserInput, *http.Cookie) {
	t.Helper()

	y := buildDummyUserInput(t)
	u := *todoClient.URL
	u.Path = "users"

	req, err := http.NewRequest(http.MethodPost, u.String(), readerFromObject(y))
	Ω(err).ShouldNot(HaveOccurred())

	res, err := todoClient.Do(req)
	Ω(err).ShouldNot(HaveOccurred())

	x, err := todoClient.CreateUser(y)
	Ω(err).ShouldNot(HaveOccurred())

	var cookie *http.Cookie
	cookies := res.Cookies()
	if len(cookies) > 0 {
		cookie = res.Cookies()[0]
	}

	return x, y, cookie
}

func TestUsers(t *testing.T) {
	g := goblin.Goblin(t)

	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Users", func() {
		g.It("should be able to be created", func() {
			// Create user
			expected := buildDummyUserInput(t)
			actual, err := todoClient.CreateUser(
				&models.UserInput{
					Username: expected.Username,
					Password: expected.Password,
				},
			)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(actual).ShouldNot(BeNil())

			// Assert user equality
			Ω(actual.ID).ShouldNot(BeZero())
			Ω(expected.Username).Should(Equal(actual.Username))
			Ω(actual.HashedPassword).Should(BeEmpty())
			Ω(actual.TwoFactorSecret).Should(BeEmpty())
			Ω(actual.CreatedOn).ShouldNot(BeZero())
			Ω(actual.UpdatedOn).Should(BeNil())
			Ω(actual.ArchivedOn).Should(BeNil())

			// Clean up
			err = todoClient.DeleteUser(actual.ID)
			Ω(err).ShouldNot(HaveOccurred())
		})

		g.It("should be able to be read", func() {
			// Create user
			expected := buildDummyUserInput(t)
			premade, err := todoClient.CreateUser(
				&models.UserInput{
					Username: expected.Username,
					Password: expected.Password,
				},
			)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(premade).ShouldNot(BeNil())

			// Fetch user
			actual, err := todoClient.GetUser(premade.Username)
			Ω(err).ShouldNot(HaveOccurred())

			// Assert user equality
			Ω(actual.ID).ShouldNot(BeZero())
			Ω(expected.Username).Should(Equal(actual.Username))
			Ω(actual.HashedPassword).Should(BeEmpty())
			Ω(actual.TwoFactorSecret).Should(BeEmpty())
			Ω(actual.CreatedOn).ShouldNot(BeZero())
			Ω(actual.UpdatedOn).Should(BeNil())
			Ω(actual.ArchivedOn).Should(BeNil())
			// Clean up
			err = todoClient.DeleteUser(actual.Username)
			Ω(err).ShouldNot(HaveOccurred())
		})

		g.It("should return an error when fetching a nonexistent one", func() {
			// Fetch user
			_, err := todoClient.GetUser(nonexistentID)
			Ω(err).Should(HaveOccurred())
			// assert.Nil(t, actual)
		})

		g.It("should be able to be read in a list", func() {
			// Create users
			expected := []*models.User{}
			for i := 0; i < 5; i++ {
				user, _, c := buildDummyUser(t)
				Ω(c).ShouldNot(BeNil())
				expected = append(expected, user)
			}

			// Assert user list equality
			actual, err := todoClient.GetUsers(nil)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(actual).ShouldNot(BeNil())
			Ω(len(expected) <= len(actual.Users)).Should(BeTrue())

			// Clean up
			for _, user := range actual.Users {
				err := todoClient.DeleteUser(user.ID)
				Ω(err).ShouldNot(HaveOccurred())
			}
		})

		g.It("should be able to change their TOTP Token")
		g.It("should be able to change their password")

		g.It("should be able to be deleted", func() {
			// Create user
			premade, _, c := buildDummyUser(t)
			Ω(c).ShouldNot(BeNil())

			// Clean up
			err := todoClient.DeleteUser(premade.ID)
			Ω(err).ShouldNot(HaveOccurred())
		})
	})
}
