package main

import (
	"context"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/httpclient"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

type itemsTargeter struct {
	client         *httpclient.Client
	createdItemIDs map[uint64]struct{}
}

func newItemsTargeter(client *httpclient.Client) *itemsTargeter {
	return &itemsTargeter{
		client:         client,
		createdItemIDs: make(map[uint64]struct{}),
	}
}

func (t *itemsTargeter) selectAction() *action {
	return selectAction(
		action{
			name:   "getItem",
			weight: 100,
		},
	)
}

func (t *itemsTargeter) Targeter() func(*vegeta.Target) error {
	return func(target *vegeta.Target) error {
		n := t.selectAction().name
		ctx := context.Background()

		switch n {
		case "getItem":
			// how do I figure out which item to get?
			if len(t.createdItemIDs) == 0 {
				return nil
			}

			req, err := t.client.BuildGetItemRequest(ctx, randomIDFromMap(t.createdItemIDs))
			if err != nil {
				return err
			}

			return initializeTargetFromRequest(req, target)
		case "createItem":
			dummyInput := fakes.BuildFakeItemCreationInput()

			req, err := t.client.BuildCreateItemRequest(ctx, dummyInput)
			if err != nil {
				return err
			}

			return initializeTargetFromRequest(req, target)
		case "updateItem":
			// how do I figure out which item to update?
			if len(t.createdItemIDs) == 0 {
				return nil
			}

			id := randomIDFromMap(t.createdItemIDs)

			x, err := t.client.GetItem(ctx, id)
			if err != nil {
				return err
			}

			newInput := fakes.BuildFakeItem()
			x.Name = newInput.Name

			req, err := t.client.BuildUpdateItemRequest(ctx, x)
			if err != nil {
				return err
			}

			return initializeTargetFromRequest(req, target)
		case "archiveItem":
			// how do I figure out which item to delete?
			if len(t.createdItemIDs) == 0 {
				return nil
			}

			req, err := t.client.BuildArchiveItemRequest(ctx, randomIDFromMap(t.createdItemIDs))
			if err != nil {
				return err
			}

			return initializeTargetFromRequest(req, target)
		default:
			return fmt.Errorf("invalid action: %q", n)
		}
	}
}
