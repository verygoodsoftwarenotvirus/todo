package httpclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestWebhooks(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(webhooksTestSuite))
}

type webhooksTestSuite struct {
	suite.Suite

	ctx                context.Context
	exampleWebhook     *types.Webhook
	exampleInput       *types.WebhookCreationInput
	exampleWebhookList *types.WebhookList
}

var _ suite.SetupTestSuite = (*webhooksTestSuite)(nil)

func (s *webhooksTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.exampleWebhook = fakes.BuildFakeWebhook()
	s.exampleInput = fakes.BuildFakeWebhookCreationInputFromWebhook(s.exampleWebhook)
	s.exampleWebhookList = fakes.BuildFakeWebhookList()
}

func (s *webhooksTestSuite) TestV1Client_GetWebhook() {
	const expectedPathFormat = "/api/v1/webhooks/%d"

	s.Run("happy path", func() {
		t := s.T()

		spec := newRequestSpec(false, http.MethodGet, "", expectedPathFormat, s.exampleWebhook.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode(s.exampleWebhook))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetWebhook(s.ctx, s.exampleWebhook.ID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, s.exampleWebhook, actual)
	})

	s.Run("with invalid client url", func() {
		t := s.T()

		actual, err := buildTestClientWithInvalidURL(t).GetWebhook(s.ctx, s.exampleWebhook.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func (s *webhooksTestSuite) TestV1Client_GetWebhooks() {
	const expectedPath = "/api/v1/webhooks"

	spec := newRequestSpec(false, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

	s.Run("happy path", func() {
		t := s.T()

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode(s.exampleWebhookList))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetWebhooks(s.ctx, nil)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, s.exampleWebhookList, actual)
	})

	s.Run("with invalid client url", func() {
		t := s.T()

		actual, err := buildTestClientWithInvalidURL(t).GetWebhooks(s.ctx, nil)
		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func (s *webhooksTestSuite) TestV1Client_CreateWebhook() {
	const expectedPath = "/api/v1/webhooks"

	spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

	s.Run("happy path", func() {
		t := s.T()

		s.exampleInput.BelongsToAccount = 0

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					var x *types.WebhookCreationInput
					require.NoError(t, json.NewDecoder(req.Body).Decode(&x))
					assert.Equal(t, s.exampleInput, x)

					require.NoError(t, json.NewEncoder(res).Encode(s.exampleWebhook))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.CreateWebhook(s.ctx, s.exampleInput)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, s.exampleWebhook, actual)
	})

	s.Run("with invalid client url", func() {
		t := s.T()

		actual, err := buildTestClientWithInvalidURL(t).CreateWebhook(s.ctx, s.exampleInput)
		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func (s *webhooksTestSuite) TestV1Client_UpdateWebhook() {
	const expectedPathFormat = "/api/v1/webhooks/%d"

	s.Run("happy path", func() {
		t := s.T()

		spec := newRequestSpec(false, http.MethodPut, "", expectedPathFormat, s.exampleWebhook.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					assert.NoError(t, json.NewEncoder(res).Encode(s.exampleWebhook))
				},
			),
		)

		err := buildTestClient(t, ts).UpdateWebhook(s.ctx, s.exampleWebhook)
		assert.NoError(t, err, "no error should be returned")
	})

	s.Run("with invalid client url", func() {
		t := s.T()

		err := buildTestClientWithInvalidURL(t).UpdateWebhook(s.ctx, s.exampleWebhook)
		assert.Error(t, err, "error should be returned")
	})
}

func (s *webhooksTestSuite) TestV1Client_ArchiveWebhook() {
	const expectedPathFormat = "/api/v1/webhooks/%d"

	s.Run("happy path", func() {
		t := s.T()

		spec := newRequestSpec(false, http.MethodDelete, "", expectedPathFormat, s.exampleWebhook.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)
				},
			),
		)

		err := buildTestClient(t, ts).ArchiveWebhook(s.ctx, s.exampleWebhook.ID)
		assert.NoError(t, err, "no error should be returned")
	})

	s.Run("with invalid client url", func() {
		t := s.T()

		err := buildTestClientWithInvalidURL(t).ArchiveWebhook(s.ctx, s.exampleWebhook.ID)
		assert.Error(t, err, "error should be returned")
	})
}

func (s *webhooksTestSuite) TestV1Client_GetAuditLogForWebhook() {
	const (
		expectedPath   = "/api/v1/webhooks/%d/audit"
		expectedMethod = http.MethodGet
	)

	s.Run("happy path", func() {
		t := s.T()

		spec := newRequestSpec(true, expectedMethod, "", expectedPath, s.exampleWebhook.ID)
		exampleAuditLogEntryList := fakes.BuildFakeAuditLogEntryList().Entries

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode(exampleAuditLogEntryList))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetAuditLogForWebhook(s.ctx, s.exampleWebhook.ID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, exampleAuditLogEntryList, actual)
	})

	s.Run("with invalid client url", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetAuditLogForWebhook(s.ctx, s.exampleWebhook.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})

	s.Run("with invalid response", func() {
		t := s.T()

		spec := newRequestSpec(true, expectedMethod, "", expectedPath, s.exampleWebhook.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode("BLAH"))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetAuditLogForWebhook(s.ctx, s.exampleWebhook.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}
