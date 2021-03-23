package main

import (
	"context"
	"errors"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/client/http"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/client/http/requests"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

type (
	vegetaHelper interface {
		Targeter() func(*vegeta.Target) error
		HandleResult(res *vegeta.Result)
	}

	itemsTargeter struct {
		client         *http.Client
		requestBuilder *requests.Builder
		createdItemIDs map[uint64]struct{}
	}
)

var (
	errInvalidAction = errors.New("invalid action")
)

func newItemsTargeter(client *http.Client, requestBuilder *requests.Builder) *itemsTargeter {
	return &itemsTargeter{
		client:         client,
		requestBuilder: requestBuilder,
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

func (t *itemsTargeter) HandleResult(res *vegeta.Result) {

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

			req, err := t.requestBuilder.BuildGetItemRequest(ctx, randomIDFromMap(t.createdItemIDs))
			if err != nil {
				return err
			}

			return initializeTargetFromRequest(req, target)
		case "createItem":
			dummyInput := fakes.BuildFakeItemCreationInput()

			req, err := t.requestBuilder.BuildCreateItemRequest(ctx, dummyInput)
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

			req, err := t.requestBuilder.BuildUpdateItemRequest(ctx, x)
			if err != nil {
				return err
			}

			return initializeTargetFromRequest(req, target)
		case "archiveItem":
			// how do I figure out which item to delete?
			if len(t.createdItemIDs) == 0 {
				return nil
			}

			req, err := t.requestBuilder.BuildArchiveItemRequest(ctx, randomIDFromMap(t.createdItemIDs))
			if err != nil {
				return err
			}

			return initializeTargetFromRequest(req, target)
		default:
			return fmt.Errorf("%s: %w", n, errInvalidAction)
		}
	}
}
