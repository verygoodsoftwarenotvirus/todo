package frontend

import (
	"github.com/nleeper/goment"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/panicking"
)

var gomentPanicker = panicking.NewProductionPanicker()

func mustGoment(ts uint64) *goment.Goment {
	g, err := goment.Unix(int64(ts))
	if err != nil {
		// literally impossible to get here, and I hate it lol
		gomentPanicker.Panic(err)
	}

	return g
}

func relativeTime(ts uint64) string {
	return mustGoment(ts).FromNow()
}

func relativeTimeFromPtr(ts *uint64) string {
	if ts == nil {
		return "never"
	}

	return relativeTime(*ts)
}
