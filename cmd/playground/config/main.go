package main

import (
	"context"
	"fmt"
	"os"

	viper "gitlab.com/verygoodsoftwarenotvirus/todo/internal/config/viper"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
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
