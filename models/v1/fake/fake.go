package fake

import (
	"time"

	fake "github.com/brianvoe/gofakeit"
)

func init() {
	fake.Seed(time.Now().UnixNano())
}
