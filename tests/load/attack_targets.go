package main

import (
	"context"
	"log"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

type attackTargetPair struct {
	attacker *vegeta.Attacker
	helper   vegetaHelper
	name     string
}

func buildAttackTargetPairs(ctx context.Context, logger logging.Logger) []*attackTargetPair {
	itemAttacker, itemsClient, requestBuilder, err := createAttacker(ctx, "items")
	if err != nil {
		log.Fatal(err)
	}

	logger = logging.EnsureLogger(logger)

	targeter := newItemsTargeter(itemsClient, requestBuilder, logger)

	return []*attackTargetPair{
		{
			name:     "items",
			attacker: itemAttacker,
			helper:   targeter,
		},
	}
}
