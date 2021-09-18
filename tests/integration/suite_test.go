package integration

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/client/httpclient"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	cookieAuthType = "cookie"
	pasetoAuthType = "PASETO"
)

var (
	globalClientExceptions []string
)

type testClientWrapper struct {
	main     *httpclient.Client
	admin    *httpclient.Client
	authType string
}

func TestIntegration(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(TestSuite))
}

type TestSuite struct {
	suite.Suite

	ctx    context.Context
	user   *types.User
	cookie *http.Cookie
	cookieClient,
	pasetoClient,
	adminCookieClient,
	adminPASETOClient *httpclient.Client
}

var _ suite.SetupTestSuite = (*TestSuite)(nil)

func (s *TestSuite) SetupTest() {
	t := s.T()
	testName := t.Name()

	ctx, span := tracing.StartCustomSpan(context.Background(), testName)
	defer span.End()

	s.ctx, _ = tracing.StartCustomSpan(ctx, testName)
	s.user, s.cookie, s.cookieClient, s.pasetoClient = createUserAndClientForTest(s.ctx, t)
	s.adminCookieClient, s.adminPASETOClient = buildAdminCookieAndPASETOClients(s.ctx, t)
}

func (s *TestSuite) runForEachClientExcept(name string, subtestBuilder func(*testClientWrapper) func(), exceptions ...string) {
	for a, c := range s.eachClientExcept(exceptions...) {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("%s via %s", name, authType), subtestBuilder(testClients))
	}
}

func (s *TestSuite) eachClientExcept(exceptions ...string) map[string]*testClientWrapper {
	t := s.T()

	clients := map[string]*testClientWrapper{
		cookieAuthType: {authType: cookieAuthType, main: s.cookieClient, admin: s.adminCookieClient},
		pasetoAuthType: {authType: pasetoAuthType, main: s.pasetoClient, admin: s.adminPASETOClient},
	}

	for _, name := range exceptions {
		delete(clients, name)
	}

	for _, name := range globalClientExceptions {
		delete(clients, name)
	}

	require.NotEmpty(t, clients)

	return clients
}

var _ suite.WithStats = (*TestSuite)(nil)

func (s *TestSuite) HandleStats(_ string, stats *suite.SuiteInformation) {
	const totalExpectedTestCount = 57

	if stats.Passed() {
		s.Equal(totalExpectedTestCount, len(stats.TestStats), "expected total number of tests run to equal %d, but it was %d", totalExpectedTestCount, len(stats.TestStats))
	}
}
