package integration

import (
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/bxcodec/faker"
	"github.com/franela/goblin"
	. "github.com/onsi/gomega"
	"github.com/pquerna/otp/totp"
)

func buildDummyOauth2ClientInput(t *testing.T, username, password, totpSecret string) *models.Oauth2ClientInput {
	t.Helper()

	code, err := totp.GenerateCode(totpSecret, time.Now())
	Ω(err).ShouldNot(HaveOccurred())

	x := &models.Oauth2ClientInput{
		UserLoginInput: models.UserLoginInput{
			Username:  username,
			Password:  password,
			TOTPToken: code,
		},
		Scopes: []string{"*"},
		Domain: faker.Internet{}.DomainName(),
	}

	return x
}

func buildDummyOauth2Client(t *testing.T, username, password, totpSecret string) *models.Oauth2Client {
	t.Helper()

	x, err := todoClient.CreateOauth2Client(
		buildDummyOauth2ClientInput(t, username, password, totpSecret),
	)
	Ω(err).ShouldNot(HaveOccurred())

	return x
}

func TestOauth2Clients(t *testing.T) {
	g := goblin.Goblin(t)

	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Oauth2Clients", func() {
		// create user
		_, x, c := buildDummyUser(t)
		Ω(c).ShouldNot(BeNil())

		g.It("should be able to be created", func() {
			// Create oauth2Client
			input := buildDummyOauth2ClientInput(t, x.Username, x.Password, x.TOTPSecret)
			actual, err := todoClient.CreateOauth2Client(input)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(actual).ShouldNot(BeNil())

			// Assert oauth2Client equality
			Ω(actual.ID).ShouldNot(BeZero())
			Ω(actual.Domain).Should(Equal(input.Domain))
			Ω(actual.Scopes).ShouldNot(Equal(input.Scopes))
			Ω(actual.CreatedOn).ShouldNot(BeZero())
			Ω(actual.UpdatedOn).Should(BeNil())
			Ω(actual.ArchivedOn).Should(BeNil())

			// Clean up
			err = todoClient.DeleteOauth2Client(actual.ID)
			Ω(err).ShouldNot(HaveOccurred())
		})

		g.It("should be able to be read", func() {
			// Create oauth2Client
			// Create oauth2Client
			input := buildDummyOauth2ClientInput(t, x.Username, x.Password, x.TOTPSecret)
			premade, err := todoClient.CreateOauth2Client(input)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(premade).ShouldNot(BeNil())

			// Fetch oauth2Client
			actual, err := todoClient.GetOauth2Client(premade.ClientID)
			Ω(err).ShouldNot(HaveOccurred())

			// Assert oauth2Client equality
			Ω(actual.ID).ShouldNot(BeZero())
			Ω(actual.Domain).Should(Equal(input.Domain))
			Ω(actual.Scopes).ShouldNot(Equal(input.Scopes))
			Ω(actual.CreatedOn).ShouldNot(BeZero())
			Ω(actual.UpdatedOn).Should(BeNil())
			Ω(actual.ArchivedOn).Should(BeNil())

			// Clean up
			err = todoClient.DeleteOauth2Client(actual.ID)
			Ω(err).ShouldNot(HaveOccurred())
		})

		g.It("should return an error when fetching a nonexistent one", func() {
			// Fetch oauth2Client
			_, err := todoClient.GetOauth2Client(nonexistentID)
			Ω(err).Should(HaveOccurred())
			// assert.Nil(t, actual)
		})

		g.It("should be able to be read in a list", func() {
			// Create oauth2Clients
			expected := []*models.Oauth2Client{}
			for i := 0; i < 5; i++ {
				expected = append(expected, buildDummyOauth2Client(t, x.Username, x.Password, x.TOTPSecret))
			}

			// Assert oauth2Client list equality
			actual, err := todoClient.GetOauth2Clients(nil)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(actual).ShouldNot(BeNil())
			Ω(len(expected) <= len(actual.Clients)).Should(BeTrue())

			// Clean up
			for _, oauth2Client := range actual.Clients {
				err := todoClient.DeleteOauth2Client(oauth2Client.ID)
				Ω(err).ShouldNot(HaveOccurred())
			}
		})

		g.It("should be able to be deleted", func() {
			// Create oauth2Client
			premade := buildDummyOauth2Client(t, x.Username, x.Password, x.TOTPSecret)

			// Clean up
			err := todoClient.DeleteOauth2Client(premade.ID)
			Ω(err).ShouldNot(HaveOccurred())
		})
	})
}
