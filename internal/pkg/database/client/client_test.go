package dbclient

import (
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
)

const (
	defaultLimit = uint8(20)
)

func buildTestClient() (*Client, *database.MockDatabase) {
	db := database.BuildMockDatabase()

	c := &Client{
		logger:  noop.NewLogger(),
		querier: db,
		tracer:  tracing.NewTracer("test"),
	}

	return c, db
}
