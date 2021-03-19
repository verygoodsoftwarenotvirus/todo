package httpclient

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
	exampleAPIClient     *types.APIClient
	exampleInput         *types.APICientCreationInput
	exampleAPIClientList *types.APIClientList
}

var _ suite.SetupTestSuite = (*apiClientsTestSuite)(nil)

func (s *apiClientsTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.exampleAPIClient = fakes.BuildFakeAPIClient()
	s.exampleAPIClient.ClientSecret = nil
	s.exampleInput = fakes.BuildFakeAPIClientCreationInputFromClient(s.exampleAPIClient)
	s.exampleAPIClientList = fakes.BuildFakeAPIClientList()

	for i := 0; i < len(s.exampleAPIClientList.Clients); i++ {
		s.exampleAPIClientList.Clients[i].ClientSecret = nil
	}
}

func (s *apiClientsTestSuite) TestV1Client_GetAPIClient() {
	const expectedPathFormat = "/api/v1/api_clients/%d"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, s.exampleAPIClient.ID)
		c, _ := buildTestClientWithJSONResponse(t, spec, s.exampleAPIClient)

		actual, err := c.GetAPIClient(s.ctx, s.exampleAPIClient.ID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, s.exampleAPIClient, actual)
	})

	s.Run("with invalid client url", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetAPIClient(s.ctx, s.exampleAPIClient.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func (s *apiClientsTestSuite) TestV1Client_GetAPIClients() {
	const expectedPath = "/api/v1/api_clients"

	spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

	s.Run("standard", func() {
		t := s.T()

		c, _ := buildTestClientWithJSONResponse(t, spec, s.exampleAPIClientList)

		actual, err := c.GetAPIClients(s.ctx, nil)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, s.exampleAPIClientList, actual)
	})

	s.Run("with invalid client url", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetAPIClients(s.ctx, nil)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func (s *apiClientsTestSuite) TestV1Client_CreateAPIClient() {
	const expectedPath = "/api/v1/api_clients"

	spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

	s.Run("standard", func() {
		t := s.T()

		exampleResponse := fakes.BuildFakeAPIClientCreationResponseFromClient(s.exampleAPIClient)
		c, _ := buildTestClientWithJSONResponse(t, spec, exampleResponse)

		actual, err := c.CreateAPIClient(s.ctx, &http.Cookie{}, s.exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse, actual)
	})

	s.Run("with invalid client url", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)

		actual, err := c.CreateAPIClient(s.ctx, &http.Cookie{}, s.exampleInput)
		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})

	s.Run("with invalid response from server", func() {
		t := s.T()

		c := buildTestClientWithInvalidResponse(t, spec)

		actual, err := c.CreateAPIClient(s.ctx, &http.Cookie{}, s.exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	s.Run("without cookie", func() {
		t := s.T()

		c, _ := buildTestClientWithJSONResponse(t, nil, s.exampleAPIClient)

		_, err := c.CreateAPIClient(s.ctx, nil, nil)
		assert.Error(t, err)
	})
}

func (s *apiClientsTestSuite) TestV1Client_ArchiveAPIClient() {
	const expectedPathFormat = "/api/v1/api_clients/%d"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, s.exampleAPIClient.ID)
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusOK)

		err := c.ArchiveAPIClient(s.ctx, s.exampleAPIClient.ID)
		assert.NoError(t, err, "no error should be returned")
	})

	s.Run("with invalid client url", func() {
		t := s.T()

		err := buildTestClientWithInvalidURL(t).ArchiveAPIClient(s.ctx, s.exampleAPIClient.ID)
		assert.Error(t, err, "error should be returned")
	})
}

func (s *apiClientsTestSuite) TestV1Client_GetAuditLogForAPIClient() {
	const (
		expectedPath   = "/api/v1/api_clients/%d/audit"
		expectedMethod = http.MethodGet
	)

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, expectedMethod, "", expectedPath, s.exampleAPIClient.ID)
		exampleAuditLogEntryList := fakes.BuildFakeAuditLogEntryList().Entries

		c, _ := buildTestClientWithJSONResponse(t, spec, exampleAuditLogEntryList)
		actual, err := c.GetAuditLogForAPIClient(s.ctx, s.exampleAPIClient.ID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, exampleAuditLogEntryList, actual)
	})

	s.Run("with invalid client url", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetAuditLogForAPIClient(s.ctx, s.exampleAPIClient.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})

	s.Run("with invalid response", func() {
		t := s.T()

		spec := newRequestSpec(true, expectedMethod, "", expectedPath, s.exampleAPIClient.ID)

		c := buildTestClientWithInvalidResponse(t, spec)
		actual, err := c.GetAuditLogForAPIClient(s.ctx, s.exampleAPIClient.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}
