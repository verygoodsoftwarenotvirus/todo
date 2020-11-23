package main

import (
	"context"
	"errors"
	"math/rand"
	"net/http"
	"time"

	client "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/httpclient"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
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
		Weight int
		Name   string
	}
)

// RandomAction takes a client and returns a closure which is an action.
func RandomAction(c *client.V1Client) *Action {
	allActions := map[string]*Action{
		"GetHealthCheck": {
			Name: "GetHealthCheck",
			Action: func() (*http.Request, error) {
				ctx := context.Background()
				return c.BuildHealthCheckRequest(ctx)
			},
			Weight: 100,
		},
		"CreateUser": {
			Name: "CreateUser",
			Action: func() (*http.Request, error) {
				ctx := context.Background()
				ui := fakes.BuildFakeUserCreationInput()
				return c.BuildCreateUserRequest(ctx, ui)
			},
			Weight: 100,
		},
	}

	for k, v := range buildItemActions(c) {
		allActions[k] = v
	}

	for k, v := range buildWebhookActions(c) {
		allActions[k] = v
	}

	for k, v := range buildOAuth2ClientActions(c) {
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
