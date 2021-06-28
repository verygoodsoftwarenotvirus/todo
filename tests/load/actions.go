package main

import (
	"context"
	"errors"
	"math/rand"
	"net/http"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/client/httpclient"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/client/httpclient/requests"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
)

var (
	// ErrUnavailableYet is a sentinel error value.
	ErrUnavailableYet = errors.New("can't do this yet")
)

type (
	// actionFunc represents a thing you can do.
	actionFunc func() (*http.Request, error)

	// Action is a wrapper struct around some important values.
	Action struct {
		Action actionFunc
		Name   string
		Weight int
	}
)

// RandomAction takes a client and returns a closure which is an action.
func RandomAction(c *httpclient.Client, builder *requests.Builder) *Action {
	allActions := map[string]*Action{
		"GetHealthCheck": {
			Name: "GetHealthCheck",
			Action: func() (*http.Request, error) {
				ctx := context.Background()
				return builder.BuildHealthCheckRequest(ctx)
			},
			Weight: 100,
		},
		"CreateUser": {
			Name: "CreateUser",
			Action: func() (*http.Request, error) {
				ctx := context.Background()
				ui := fakes.BuildFakeUserCreationInput()
				return builder.BuildCreateUserRequest(ctx, ui)
			},
			Weight: 100,
		},
	}

	for k, v := range buildItemActions(c, builder) {
		allActions[k] = v
	}

	for k, v := range buildWebhookActions(c, builder) {
		allActions[k] = v
	}

	totalWeight := 0
	for _, rb := range allActions {
		totalWeight += rb.Weight
	}

	rand.Seed(time.Now().UnixNano())
	r := rand.Intn(totalWeight)

	for _, rb := range allActions {
		r -= rb.Weight
		if r <= 0 {
			return rb
		}
	}

	return nil
}
