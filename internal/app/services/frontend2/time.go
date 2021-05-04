package frontend2

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/panicking"

	"github.com/nleeper/goment"
)

var gomentPanicker = panicking.NewProductionPanicker()

func mustGoment(ts uint64) *goment.Goment {
	g, err := goment.Unix(int64(ts))
	if err != nil {
		// literally impossible
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
