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

func TestAPIClients(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(apiClientsTestSuite))
}

type apiClientsTestSuite struct {
	suite.Suite

	ctx                  context.Context
	builder              *Builder
	exampleAPIClient     *types.APIClient
	exampleInput         *types.APICientCreationInput
	exampleAPIClientList *types.APIClientList
}

var _ suite.SetupTestSuite = (*apiClientsTestSuite)(nil)

func (s *apiClientsTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.builder = buildTestRequestBuilder()
	s.exampleAPIClient = fakes.BuildFakeAPIClient()
	s.exampleAPIClient.ClientSecret = nil
	s.exampleInput = fakes.BuildFakeAPIClientCreationInputFromClient(s.exampleAPIClient)
	s.exampleAPIClientList = fakes.BuildFakeAPIClientList()

	for i := 0; i < len(s.exampleAPIClientList.Clients); i++ {
		s.exampleAPIClientList.Clients[i].ClientSecret = nil
	}
}

func (s *apiClientsTestSuite) TestBuilder_BuildGetAPIClientRequest() {
	const expectedPathFormat = "/api/v1/api_clients/%d"

	s.Run("happy path", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, s.exampleAPIClient.ID)

		actual, err := s.builder.BuildGetAPIClientRequest(s.ctx, s.exampleAPIClient.ID)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func (s *apiClientsTestSuite) TestBuilder_BuildGetAPIClientsRequest() {
	const expectedPath = "/api/v1/api_clients"

	s.Run("happy path", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

		actual, err := s.builder.BuildGetAPIClientsRequest(s.ctx, nil)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func (s *apiClientsTestSuite) TestBuilder_BuildCreateAPIClientRequest() {
	const expectedPath = "/api/v1/api_clients"

	s.Run("happy path", func() {
		t := s.T()

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

		req, err := s.builder.BuildCreateAPIClientRequest(s.ctx, &http.Cookie{}, s.exampleInput)
		assert.NoError(t, err)

		assertRequestQuality(t, req, spec)
	})
}

func (s *apiClientsTestSuite) TestBuilder_BuildArchiveAPIClientRequest() {
	const expectedPathFormat = "/api/v1/api_clients/%d"

	s.Run("happy path", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, s.exampleAPIClient.ID)

		actual, err := s.builder.BuildArchiveAPIClientRequest(s.ctx, s.exampleAPIClient.ID)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func (s *apiClientsTestSuite) TestBuilder_BuildGetAuditLogForAPIClientRequest() {
	const expectedPath = "/api/v1/api_clients/%d/audit"

	s.Run("happy path", func() {
		t := s.T()

		actual, err := s.builder.BuildGetAuditLogForAPIClientRequest(s.ctx, s.exampleAPIClient.ID)
		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")

		spec := newRequestSpec(true, http.MethodGet, "", expectedPath, s.exampleAPIClient.ID)
		assertRequestQuality(t, actual, spec)
	})
}
