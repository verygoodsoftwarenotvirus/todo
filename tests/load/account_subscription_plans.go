package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/httpclient"
)

// fetchRandomPlan retrieves a random plan from the list of available plans.
func fetchRandomPlan(ctx context.Context, c *httpclient.Client) *types.AccountSubscriptionPlan {
	plansRes, err := c.GetPlans(ctx, nil)
	if err != nil || plansRes == nil || len(plansRes.AccountSubscriptionPlans) == 0 {
		return nil
	}

	randIndex := rand.Intn(len(plansRes.AccountSubscriptionPlans))

	return plansRes.AccountSubscriptionPlans[randIndex]
}

func buildPlanActions(shouldUse bool, c *httpclient.Client) map[string]*Action {
	if shouldUse {
		return nil
	}

	return map[string]*Action{
		"CreateAccountSubscriptionPlan": {
			Name: "CreateAccountSubscriptionPlan",
			Action: func() (*http.Request, error) {
				ctx := context.Background()

				planInput := fakes.BuildFakeAccountSubscriptionPlanCreationInput()

				return c.BuildCreatePlanRequest(ctx, planInput)
			},
			Weight: 100,
		},
		"GetAccountSubscriptionPlan": {
			Name: "GetAccountSubscriptionPlan",
			Action: func() (*http.Request, error) {
				ctx := context.Background()

				randomPlan := fetchRandomPlan(ctx, c)
				if randomPlan == nil {
					return nil, fmt.Errorf("retrieving random plan: %w", ErrUnavailableYet)
				}

				return c.BuildGetPlanRequest(ctx, randomPlan.ID)
			},
			Weight: 100,
		},
		"GetAccountSubscriptionPlans": {
			Name: "GetAccountSubscriptionPlans",
			Action: func() (*http.Request, error) {
				ctx := context.Background()

				return c.BuildGetPlansRequest(ctx, nil)
			},
			Weight: 100,
		},
		"UpdateAccountSubscriptionPlan": {
			Name: "UpdateAccountSubscriptionPlan",
			Action: func() (*http.Request, error) {
				ctx := context.Background()

				if randomPlan := fetchRandomPlan(ctx, c); randomPlan != nil {
					newPlan := fakes.BuildFakeAccountSubscriptionPlanCreationInput()
					randomPlan.Name = newPlan.Name
					randomPlan.Description = newPlan.Description
					randomPlan.Price = newPlan.Price
					randomPlan.Period = newPlan.Period
					return c.BuildUpdatePlanRequest(ctx, randomPlan)
				}

				return nil, ErrUnavailableYet
			},
			Weight: 100,
		},
		"ArchiveAccountSubscriptionPlan": {
			Name: "ArchiveAccountSubscriptionPlan",
			Action: func() (*http.Request, error) {
				ctx := context.Background()

				randomPlan := fetchRandomPlan(ctx, c)
				if randomPlan == nil {
					return nil, fmt.Errorf("retrieving random plan: %w", ErrUnavailableYet)
				}

				return c.BuildArchivePlanRequest(ctx, randomPlan.ID)
			},
			Weight: 85,
		},
	}
}
