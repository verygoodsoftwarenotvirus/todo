package integration

import (
	"testing"

	// "gitlab.com/verygoodsoftwarenotvirus/todo/models"

	// "github.com/bxcodec/faker"
	"github.com/franela/goblin"
)

func TestAll(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("Auth", func() {
		g.It("should reject an unauthenticated request")
		g.Describe("credentials", func() {
			g.It("should accept a valid cookie")
			g.It("should reject a valid cookie")
			g.It("should accept a valid auth key")
			g.It("should reject an invalid auth key")
		})
	})

	g.Describe("Items", func() {
		g.It("Should create an item", func() {

		})
		g.It("Should return a pre-made item")
		g.It("Should return a list of pre-made items")
		g.It("Should update a item")
		g.It("Should delete a item")
	})

}
