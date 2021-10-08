package httpclient

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
)

func TestWebhooks(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(webhooksTestSuite))
}

type webhooksTestSuite struct {
	suite.Suite

	ctx                context.Context
	exampleWebhook     *types.Webhook
	exampleWebhookList *types.WebhookList
}

var _ suite.SetupTestSuite = (*webhooksTestSuite)(nil)

func (s *webhooksTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.exampleWebhook = fakes.BuildFakeWebhook()
	s.exampleWebhookList = fakes.BuildFakeWebhookList()
}

func (s *webhooksTestSuite) TestClient_GetWebhook() {
	const expectedPathFormat = "/api/v1/webhooks/%s"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(false, http.MethodGet, "", expectedPathFormat, s.exampleWebhook.ID)
		c, _ := buildTestClientWithJSONResponse(t, spec, s.exampleWebhook)

		actual, err := c.GetWebhook(s.ctx, s.exampleWebhook.ID)
		assert.NoError(t, err)
		assert.Equal(t, s.exampleWebhook, actual)
	})

	s.Run("with invalid webhook ID", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)

		actual, err := c.GetWebhook(s.ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)

		actual, err := c.GetWebhook(s.ctx, s.exampleWebhook.ID)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	s.Run("with error executing request", func() {
		t := s.T()

		c, _ := buildTestClientThatWaitsTooLong(t)

		actual, err := c.GetWebhook(s.ctx, s.exampleWebhook.ID)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func (s *webhooksTestSuite) TestClient_GetWebhooks() {
	const expectedPath = "/api/v1/webhooks"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(false, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)
		c, _ := buildTestClientWithJSONResponse(t, spec, s.exampleWebhookList)

		actual, err := c.GetWebhooks(s.ctx, nil)
		assert.NoError(t, err)
		assert.Equal(t, s.exampleWebhookList, actual)
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)

		actual, err := c.GetWebhooks(s.ctx, nil)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	s.Run("with error executing request", func() {
		t := s.T()

		c, _ := buildTestClientThatWaitsTooLong(t)

		actual, err := c.GetWebhooks(s.ctx, nil)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func (s *webhooksTestSuite) TestClient_CreateWebhook() {
	const expectedPath = "/api/v1/webhooks"

	s.Run("standard", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeWebhookCreationInputFromWebhook(s.exampleWebhook)
		exampleInput.ID = ""
		exampleInput.BelongsToAccount = ""

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)
		c, _ := buildTestClientWithJSONResponse(t, spec, &types.PreWriteResponse{ID: s.exampleWebhook.ID})

		actual, err := c.CreateWebhook(s.ctx, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, s.exampleWebhook.ID, actual)
	})

	s.Run("with nil input", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)

		actual, err := c.CreateWebhook(s.ctx, nil)
		assert.Error(t, err)
		assert.Empty(t, actual)
	})

	s.Run("with invalid input", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)

		actual, err := c.CreateWebhook(s.ctx, &types.WebhookCreationRequestInput{})
		assert.Error(t, err)
		assert.Empty(t, actual)
	})

	s.Run("with error building request", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeWebhookCreationInputFromWebhook(s.exampleWebhook)
		c := buildTestClientWithInvalidURL(t)

		actual, err := c.CreateWebhook(s.ctx, exampleInput)
		assert.Error(t, err)
		assert.Empty(t, actual)
	})

	s.Run("with error executing request", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeWebhookCreationInputFromWebhook(s.exampleWebhook)
		c, _ := buildTestClientThatWaitsTooLong(t)

		actual, err := c.CreateWebhook(s.ctx, exampleInput)
		assert.Error(t, err)
		assert.Empty(t, actual)
	})
}

func (s *webhooksTestSuite) TestClient_ArchiveWebhook() {
	const expectedPathFormat = "/api/v1/webhooks/%s"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, s.exampleWebhook.ID)
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusOK)

		err := c.ArchiveWebhook(s.ctx, s.exampleWebhook.ID)
		assert.NoError(t, err)
	})

	s.Run("with invalid webhook ID", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)

		err := c.ArchiveWebhook(s.ctx, "")
		assert.Error(t, err)
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)

		err := c.ArchiveWebhook(s.ctx, s.exampleWebhook.ID)
		assert.Error(t, err)
	})

	s.Run("with error executing request", func() {
		t := s.T()

		c, _ := buildTestClientThatWaitsTooLong(t)

		err := c.ArchiveWebhook(s.ctx, s.exampleWebhook.ID)
		assert.Error(t, err)
	})
}
