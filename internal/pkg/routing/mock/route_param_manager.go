package mock

import (
	"net/http"

	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

// NewRouteParamManager returns a new RouteParamManager.
func NewRouteParamManager() *RouteParamManager {
	return &RouteParamManager{}
}

// RouteParamManager is a mock routing.RouteParamManager.
type RouteParamManager struct {
	mock.Mock
}

// UserIDFetcherFromRequestContext satisfies our interface contract.
func (m *RouteParamManager) UserIDFetcherFromRequestContext(req *http.Request) uint64 {
	return m.Called(req).Get(0).(uint64)
}

// SessionInfoFetcherFromRequestContext satisfies our interface contract.
func (m *RouteParamManager) SessionInfoFetcherFromRequestContext(req *http.Request) (*types.RequestContext, error) {
	args := m.Called(req)

	return args.Get(0).(*types.RequestContext), args.Error(1)
}

// BuildRouteParamIDFetcher satisfies our interface contract.
func (m *RouteParamManager) BuildRouteParamIDFetcher(logger logging.Logger, key, logDescription string) func(*http.Request) uint64 {
	return m.Called(logger, key, logDescription).Get(0).(func(*http.Request) uint64)
}

// BuildRouteParamStringIDFetcher satisfies our interface contract.
func (m *RouteParamManager) BuildRouteParamStringIDFetcher(key string) func(req *http.Request) string {
	return m.Called(key).Get(0).(func(*http.Request) string)
}
