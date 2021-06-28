package httpclient

import (
	"context"
	"net/http"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestAPIClients(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(apiClientsTestSuite))
}

type apiClientsTestSuite struct {
	suite.Suite

	ctx                  context.Context
	exampleAPIClient     *types.APIClient
	exampleAPIClientList *types.APIClientList
}

var _ suite.SetupTestSuite = (*apiClientsTestSuite)(nil)

func (s *apiClientsTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.exampleAPIClient = fakes.BuildFakeAPIClient()
	s.exampleAPIClient.ClientSecret = nil
	s.exampleAPIClientList = fakes.BuildFakeAPIClientList()

	for i := 0; i < len(s.exampleAPIClientList.Clients); i++ {
		s.exampleAPIClientList.Clients[i].ClientSecret = nil
	}
}

func (s *apiClientsTestSuite) TestClient_GetAPIClient() {
	const expectedPathFormat = "/api/v1/api_clients/%d"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, s.exampleAPIClient.ID)
		c, _ := buildTestClientWithJSONResponse(t, spec, s.exampleAPIClient)

		actual, err := c.GetAPIClient(s.ctx, s.exampleAPIClient.ID)
		require.NotNil(t, actual)
		assert.NoError(t, err)
		assert.Equal(t, s.exampleAPIClient, actual)
	})

	s.Run("with invalid API client ID", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, s.exampleAPIClient.ID)
		c, _ := buildTestClientWithJSONResponse(t, spec, s.exampleAPIClient)

		actual, err := c.GetAPIClient(s.ctx, 0)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)

		actual, err := c.GetAPIClient(s.ctx, s.exampleAPIClient.ID)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	s.Run("with error executing request", func() {
		t := s.T()

		c, _ := buildTestClientThatWaitsTooLong(t)

		actual, err := c.GetAPIClient(s.ctx, s.exampleAPIClient.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func (s *apiClientsTestSuite) TestClient_GetAPIClients() {
	const expectedPath = "/api/v1/api_clients"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)
		c, _ := buildTestClientWithJSONResponse(t, spec, s.exampleAPIClientList)

		actual, err := c.GetAPIClients(s.ctx, nil)
		require.NotNil(t, actual)
		assert.NoError(t, err)
		assert.Equal(t, s.exampleAPIClientList, actual)
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetAPIClients(s.ctx, nil)

		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	s.Run("with error executing request", func() {
		t := s.T()

		c, _ := buildTestClientThatWaitsTooLong(t)

		actual, err := c.GetAPIClients(s.ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func (s *apiClientsTestSuite) TestClient_CreateAPIClient() {
	const expectedPath = "/api/v1/api_clients"

	s.Run("standard", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeAPIClientCreationInputFromClient(s.exampleAPIClient)
		exampleResponse := fakes.BuildFakeAPIClientCreationResponseFromClient(s.exampleAPIClient)
		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)
		c, _ := buildTestClientWithJSONResponse(t, spec, exampleResponse)

		actual, err := c.CreateAPIClient(s.ctx, &http.Cookie{}, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleResponse, actual)
	})

	s.Run("with nil cookie", func() {
		t := s.T()

		c, _ := buildTestClientWithJSONResponse(t, nil, s.exampleAPIClient)

		_, err := c.CreateAPIClient(s.ctx, nil, nil)
		assert.Error(t, err)
	})

	s.Run("with nil input", func() {
		t := s.T()

		c, _ := buildTestClientWithJSONResponse(t, nil, s.exampleAPIClient)

		_, err := c.CreateAPIClient(s.ctx, &http.Cookie{}, nil)
		assert.Error(t, err)
	})

	s.Run("with error building request", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeAPIClientCreationInputFromClient(s.exampleAPIClient)
		c := buildTestClientWithInvalidURL(t)

		actual, err := c.CreateAPIClient(s.ctx, &http.Cookie{}, exampleInput)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	s.Run("with invalid response from server", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeAPIClientCreationInputFromClient(s.exampleAPIClient)
		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)
		c := buildTestClientWithInvalidResponse(t, spec)

		actual, err := c.CreateAPIClient(s.ctx, &http.Cookie{}, exampleInput)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func (s *apiClientsTestSuite) TestClient_ArchiveAPIClient() {
	const expectedPathFormat = "/api/v1/api_clients/%d"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, s.exampleAPIClient.ID)
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusOK)

		assert.NoError(t, c.ArchiveAPIClient(s.ctx, s.exampleAPIClient.ID), "no error should be returned")
	})

	s.Run("with invalid API client ID", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, s.exampleAPIClient.ID)
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusOK)

		assert.Error(t, c.ArchiveAPIClient(s.ctx, 0), "no error should be returned")
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		assert.Error(t, c.ArchiveAPIClient(s.ctx, s.exampleAPIClient.ID), "error should be returned")
	})

	s.Run("with error executing request", func() {
		t := s.T()

		c, _ := buildTestClientThatWaitsTooLong(t)

		assert.Error(t, c.ArchiveAPIClient(s.ctx, s.exampleAPIClient.ID), "no error should be returned")
	})
}

func (s *apiClientsTestSuite) TestClient_GetAuditLogForAPIClient() {
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
		assert.NoError(t, err)
		assert.Equal(t, exampleAuditLogEntryList, actual)
	})

	s.Run("with invalid API client ID", func() {
		t := s.T()

		spec := newRequestSpec(true, expectedMethod, "", expectedPath, s.exampleAPIClient.ID)
		exampleAuditLogEntryList := fakes.BuildFakeAuditLogEntryList().Entries

		c, _ := buildTestClientWithJSONResponse(t, spec, exampleAuditLogEntryList)

		actual, err := c.GetAuditLogForAPIClient(s.ctx, 0)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)

		actual, err := c.GetAuditLogForAPIClient(s.ctx, s.exampleAPIClient.ID)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	s.Run("with error executing request", func() {
		t := s.T()

		spec := newRequestSpec(true, expectedMethod, "", expectedPath, s.exampleAPIClient.ID)
		c := buildTestClientWithInvalidResponse(t, spec)

		actual, err := c.GetAuditLogForAPIClient(s.ctx, s.exampleAPIClient.ID)
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}
