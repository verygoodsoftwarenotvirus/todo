package integration

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	httpclient "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/client/http"
)

const (
	cookieAuthType = "cookie"
	pasetoAuthType = "PASETO"
)

var (
	globalClientExceptions []string
)

type testClientWrapper struct {
	main, admin *httpclient.Client
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

func (s *TestSuite) ensure() {
	t := s.T()
	t.Helper()

	require.NotNil(t, s.ctx)
	require.NotNil(t, s.user)
	require.NotNil(t, s.cookie)
	require.NotNil(t, s.cookieClient)
	require.NotNil(t, s.pasetoClient)
	require.NotNil(t, s.adminCookieClient)
	require.NotNil(t, s.adminCookieClient)
}

var _ suite.SetupTestSuite = (*TestSuite)(nil)

func (s *TestSuite) SetupTest() {
	t := s.T()
	testName := t.Name()

	ctx, span := tracing.StartCustomSpan(context.Background(), testName)
	defer span.End()

	s.user, s.cookie, s.cookieClient, s.pasetoClient = createUserAndClientForTest(ctx, t)
	s.adminCookieClient, s.adminPASETOClient = buildAdminCookieAndPASETOClients(ctx, t)
	s.ctx, _ = tracing.StartCustomSpan(ctx, testName)

	s.ensure()
}

/*
func (s *TestSuite) runForEachClientExcept(t *testing.T, subtestBuilder func(*testClientWrapper) func(), exceptions ...string) {
	t.Helper()

	for a, c := range s.eachClientExcept(exceptions...) {
		authType, testClients := a, c
		s.Run(fmt.Sprintf("%s via %s", t.Name(), authType), subtestBuilder(testClients))
	}
}
*/

func (s *TestSuite) eachClientExcept(exceptions ...string) map[string]*testClientWrapper {
	s.ensure()
	t := s.T()

	clients := map[string]*testClientWrapper{
		cookieAuthType: {main: s.cookieClient, admin: s.adminCookieClient},
		pasetoAuthType: {main: s.pasetoClient, admin: s.adminPASETOClient},
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

const minimumTestThreshold = 1 * time.Millisecond

func (s *TestSuite) checkTestRunsForPositiveResultsThatOccurredTooQuickly(stats *suite.SuiteInformation) {
	t := s.T()

	for testName, stat := range stats.TestStats {
		if stat.End.Sub(stat.Start) < minimumTestThreshold && stat.Passed {
			t.Fatalf("suspiciously quick test execution time: %q", testName)
		}
	}
}

func (s *TestSuite) HandleStats(_ string, stats *suite.SuiteInformation) {
	s.checkTestRunsForPositiveResultsThatOccurredTooQuickly(stats)
}
