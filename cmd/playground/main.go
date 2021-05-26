package main

import (
	"context"
	"fmt"
	"os"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/secrets"
)

func initializeLocalSecretManager(ctx context.Context) secrets.SecretManager {
	logger := logging.NewNoopLogger()

	cfg := &secrets.Config{
		Provider: secrets.ProviderLocal,
		Key:      os.Getenv("TODO_LOCAL_SECRET_STORE_KEY"),
	}

	k, err := secrets.ProvideSecretKeeper(ctx, cfg)
	if err != nil {
		panic(err)
	}

	sm, err := secrets.ProvideSecretManager(logger, k)
	if err != nil {
		panic(err)
	}

	return sm
}

func main() {
	ctx := context.Background()

	content, err := os.ReadFile("environments/local/todo.config")
	if err != nil {
		panic(err)
	}

	sm := initializeLocalSecretManager(ctx)

	var cfg *config.ServiceConfig
	err = sm.Decrypt(ctx, string(content), &cfg)
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg)
}
