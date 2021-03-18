package main

import (
	"context"
	"log"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

type attackTargetPair struct {
	attacker *vegeta.Attacker
	helper   vegetaHelper
	name     string
}

func buildAttackTargetPairs(ctx context.Context) []*attackTargetPair {
	itemAttacker, itemsClient, requestBuilder, err := createAttacker(ctx, "items")
	if err != nil {
		log.Fatal(err)
	}

	targeter := newItemsTargeter(itemsClient, requestBuilder)

	return []*attackTargetPair{
		{
			name:     "items",
			attacker: itemAttacker,
			helper:   targeter,
		},
	}
}
