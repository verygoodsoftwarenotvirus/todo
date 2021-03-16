package main

import (
	"context"
	"log"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

type attackTargetPair struct {
	attacker *vegeta.Attacker
	targeter vegeta.Targeter
	name     string
}

func buildAttackTargetPairs(ctx context.Context) []*attackTargetPair {
	itemAttacker, itemsClient, err := createAttacker(ctx, "items")
	if err != nil {
		log.Fatal(err)
	}

	itemTargeter := newItemsTargeter(itemsClient).Targeter()

	return []*attackTargetPair{
		{
			attacker: itemAttacker,
			targeter: itemTargeter,
		},
	}
}
