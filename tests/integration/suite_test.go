package integration

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/httpclient"
)

const (
	cookieAuthType = "cookie"
	pasetoAuthType = "PASETO"
)

func TestIntegration(t *testing.T) {
	t.Parallel()

	suite.Run(t, &TestSuite{})
}

type TestSuite struct {
	suite.Suite

	ctx          context.Context
	user         *types.User
	cookie       *http.Cookie
	cookieClient *httpclient.Client
	pasetoClient *httpclient.Client

	adminCookieClient *httpclient.Client
	adminPASETOClient *httpclient.Client
}

func (s *TestSuite) eachClient(exceptions ...string) map[string]testClientWrapper {
	eachClient := map[string]testClientWrapper{
		cookieAuthType: {main: s.cookieClient, admin: s.adminCookieClient},
		pasetoAuthType: {main: s.pasetoClient, admin: s.adminPASETOClient},
	}
	output := map[string]testClientWrapper{}

	for x, c := range eachClient {
		for _, name := range exceptions {
			if strings.TrimSpace(x) != strings.TrimSpace(name) {
				output[x] = c
			}
		}
	}

	return output
}

func (s *TestSuite) SetupTest(t *testing.T) {
	t.Helper()
	testName := t.Name()

	ctx, span := tracing.StartCustomSpan(context.Background(), testName)
	defer span.End()

	s.user, s.cookie, s.cookieClient, s.pasetoClient = createUserAndClientForTest(ctx, t)
	s.adminCookieClient, s.adminPASETOClient = buildAdminCookieAndPASETOClients(ctx)
	s.ctx, _ = tracing.StartCustomSpan(ctx, testName)

	t.Logf("created user %d for test %q", s.user.ID, testName)
}

func (s *TestSuite) HandleStats(_ string, stats *suite.SuiteInformation) {
	t := s.T()

	for testName, stat := range stats.TestStats {
		var state = "failed"
		if stat.Passed {
			state = "passed"
		}

		t.Logf("%s %s in %v", testName, state, stat.End.Sub(stat.Start))
	}
}
