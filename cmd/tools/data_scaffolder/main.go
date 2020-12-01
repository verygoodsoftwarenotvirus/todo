package main

import (
	"errors"
	"os"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/zerolog"

	flag "github.com/spf13/pflag"
)

var (
	uri   string
	count int
)

func init() {
	flag.StringVarP(&uri, "url", "u", "", "where the target instance is hosted")
	flag.IntVarP(&count, "count", "c", -1, "how many users/items per user to create")
}

func main() {
	flag.Parse()
	logger := zerolog.NewLogger()

	if uri == "" {
		logger.Fatal(errors.New("uri must not be empty"))
	}

	if count <= 0 {
		os.Exit(0)
	}

	for i := 0; i < count; i++ {
		_ = i
		// create user

		for j := 0; j < count; j++ {
			_ = j
			// create items for user
			// create oauth2 clients for user
			// create webhooks for user
		}
	}
}
