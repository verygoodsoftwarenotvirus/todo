package requests

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhooks(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(webhooksTestSuite))
}

type webhooksTestSuite struct {
	suite.Suite

	ctx                context.Context
	builder            *Builder
	exampleWebhook     *types.Webhook
	exampleInput       *types.WebhookCreationInput
	exampleWebhookList *types.WebhookList
}

var _ suite.SetupTestSuite = (*webhooksTestSuite)(nil)

func (s *webhooksTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.builder = buildTestRequestBuilder()
	s.exampleWebhook = fakes.BuildFakeWebhook()
	s.exampleInput = fakes.BuildFakeWebhookCreationInputFromWebhook(s.exampleWebhook)
	s.exampleWebhookList = fakes.BuildFakeWebhookList()
}

func (s *webhooksTestSuite) TestBuilder_BuildGetWebhookRequest() {
	const expectedPathFormat = "/api/v1/webhooks/%d"

	s.Run("happy path", func() {
		t := s.T()

		spec := newRequestSpec(false, http.MethodGet, "", expectedPathFormat, s.exampleWebhook.ID)

		actual, err := s.builder.BuildGetWebhookRequest(s.ctx, s.exampleWebhook.ID)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func (s *webhooksTestSuite) TestBuilder_BuildGetWebhooksRequest() {
	const expectedPath = "/api/v1/webhooks"

	s.Run("happy path", func() {
		t := s.T()

		spec := newRequestSpec(false, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

		actual, err := s.builder.BuildGetWebhooksRequest(s.ctx, nil)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func (s *webhooksTestSuite) TestBuilder_BuildCreateWebhookRequest() {
	const expectedPath = "/api/v1/webhooks"

	s.Run("happy path", func() {
		t := s.T()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

		actual, err := s.builder.BuildCreateWebhookRequest(s.ctx, s.exampleInput)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func (s *webhooksTestSuite) TestBuilder_BuildUpdateWebhookRequest() {
	const expectedPathFormat = "/api/v1/webhooks/%d"

	s.Run("happy path", func() {
		t := s.T()

		spec := newRequestSpec(false, http.MethodPut, "", expectedPathFormat, s.exampleWebhook.ID)

		actual, err := s.builder.BuildUpdateWebhookRequest(s.ctx, s.exampleWebhook)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func (s *webhooksTestSuite) TestBuilder_BuildArchiveWebhookRequest() {
	const expectedPathFormat = "/api/v1/webhooks/%d"

	s.Run("happy path", func() {
		t := s.T()

		spec := newRequestSpec(false, http.MethodDelete, "", expectedPathFormat, s.exampleWebhook.ID)

		actual, err := s.builder.BuildArchiveWebhookRequest(s.ctx, s.exampleWebhook.ID)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func (s *webhooksTestSuite) TestBuilder_BuildGetAuditLogForWebhookRequest() {
	const expectedPath = "/api/v1/webhooks/%d/audit"

	s.Run("happy path", func() {
		t := s.T()

		actual, err := s.builder.BuildGetAuditLogForWebhookRequest(s.ctx, s.exampleWebhook.ID)
		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")

		spec := newRequestSpec(true, http.MethodGet, "", expectedPath, s.exampleWebhook.ID)
		assertRequestQuality(t, actual, spec)
	})
}
