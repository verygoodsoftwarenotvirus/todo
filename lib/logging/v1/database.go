package logging

import (
	"context"
	"database/sql"
	"database/sql/driver"

	"github.com/ExpansiveWorlds/instrumentedsql"
)

// ProvideDatabaseConnection provides a logged database connection
func ProvideDatabaseConnection(logger Logger, driver driver.Driver, connStr string) (*sql.DB, error) {
	sql.Register(
		"instrumented",
		instrumentedsql.WrapDriver(
			driver,
			// instrumentedsql.WithTracer(instrumentedsql.Tracer),
			instrumentedsql.WithLogger(instrumentedsql.LoggerFunc(
				func(ctx context.Context, msg string, keyvals ...interface{}) {
					var (
						currentKey string
					)

					values := map[string]interface{}{}
					for i, x := range keyvals {
						if i%2 == 0 {
							if y, ok := x.(string); ok {
								currentKey = y
							}
						} else if currentKey != "" && x != nil {
							values[currentKey] = x
							currentKey = ""
						}
					}
					values["msg"] = msg

					logger.WithValues(values).Debug("")
				},
			)),
		),
	)

	return sql.Open("instrumented", connStr)
}
