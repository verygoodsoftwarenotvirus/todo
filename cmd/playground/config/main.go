package main

import (
	"context"
	"fmt"
	"os"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/config/viper"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
)

const exampleConfig = ``

func mustnt(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	f, fileCreateErr := os.CreateTemp("", "*.toml")
	mustnt(fileCreateErr)

	_, writeErr := f.WriteString(exampleConfig)
	mustnt(writeErr)

	cfg, configParseErr := viper.ParseConfigFile(context.Background(), logging.NewNonOperationalLogger(), f.Name())
	mustnt(configParseErr)

	fmt.Println(cfg.Database.CreateTestUser)
}
