package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/client/httpclient"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/client/httpclient/requests"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

type (
	vegetaHelper interface {
		Targeter() func(*vegeta.Target) error
		HandleResult(res *vegeta.Result)
	}

	itemsTargeter struct {
		logger         logging.Logger
		client         *httpclient.Client
		requestBuilder *requests.Builder
		createdItems   map[uint64]struct{}
		createdItemHat sync.RWMutex
	}
)

var (
	errInvalidAction = errors.New("invalid action")
)

func newItemsTargeter(client *httpclient.Client, requestBuilder *requests.Builder, logger logging.Logger) *itemsTargeter {
	return &itemsTargeter{
		client:         client,
		requestBuilder: requestBuilder,
		logger:         logging.EnsureLogger(logger),
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
	u, err := url.ParseRequestURI(res.URL)
	if err != nil {
		// this should never occur
		panic(err)
	}

	switch u.Path {
	case "/api/v1/items":
		if res.Method == http.MethodPost {
			var x *types.Item

			if err = json.NewDecoder(bytes.NewReader(res.Body)).Decode(&x); err != nil {
				t.logger.Error(err, "decoding item response")
			}

			t.createdItemHat.Lock()
			defer t.createdItemHat.Unlock()
			t.createdItems[x.ID] = struct{}{}
		}
	default:
		t.logger.WithValue("path", u.Path).Info("request made and skipped")
	}
}

func (t *itemsTargeter) Targeter() func(*vegeta.Target) error {
	return func(target *vegeta.Target) error {
		n := t.selectAction().name
		ctx := context.Background()

		switch n {
		case "getItem":
			// how do I figure out which item to get?
			if len(t.createdItems) == 0 {
				return nil
			}

			req, err := t.requestBuilder.BuildGetItemRequest(ctx, randomIDFromMap(&t.createdItemHat, t.createdItems))
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
			if len(t.createdItems) == 0 {
				return nil
			}

			id := randomIDFromMap(&t.createdItemHat, t.createdItems)

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
			if len(t.createdItems) == 0 {
				return nil
			}

			id := randomIDFromMap(&t.createdItemHat, t.createdItems)

			req, err := t.requestBuilder.BuildArchiveItemRequest(ctx, id)
			if err != nil {
				return err
			}

			t.createdItemHat.Lock()
			delete(t.createdItems, id)

			return initializeTargetFromRequest(req, target)
		default:
			return fmt.Errorf("%s: %w", n, errInvalidAction)
		}
	}
}
