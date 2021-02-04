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

// fetchRandomAccountSubscriptionPlan retrieves a random plan from the list of available accountsubscriptionplans.
func fetchRandomAccountSubscriptionPlan(ctx context.Context, c *httpclient.Client) *types.AccountSubscriptionPlan {
	plansRes, err := c.GetAccountSubscriptionPlans(ctx, nil)
	if err != nil || plansRes == nil || len(plansRes.AccountSubscriptionPlans) == 0 {
		return nil
	}

	randIndex := rand.Intn(len(plansRes.AccountSubscriptionPlans))

	return plansRes.AccountSubscriptionPlans[randIndex]
}

func buildAccountSubscriptionPlanActions(shouldUse bool, c *httpclient.Client) map[string]*Action {
	if shouldUse {
		return nil
	}

	return map[string]*Action{
		"CreateAccountSubscriptionPlan": {
			Name: "CreateAccountSubscriptionPlan",
			Action: func() (*http.Request, error) {
				ctx := context.Background()

				planInput := fakes.BuildFakeAccountSubscriptionPlanCreationInput()

				return c.BuildCreateAccountSubscriptionPlanRequest(ctx, planInput)
			},
			Weight: 100,
		},
		"GetAccountSubscriptionPlan": {
			Name: "GetAccountSubscriptionPlan",
			Action: func() (*http.Request, error) {
				ctx := context.Background()

				randomAccountSubscriptionPlan := fetchRandomAccountSubscriptionPlan(ctx, c)
				if randomAccountSubscriptionPlan == nil {
					return nil, fmt.Errorf("retrieving random plan: %w", ErrUnavailableYet)
				}

				return c.BuildGetAccountSubscriptionPlanRequest(ctx, randomAccountSubscriptionPlan.ID)
			},
			Weight: 100,
		},
		"GetAccountSubscriptionPlans": {
			Name: "GetAccountSubscriptionPlans",
			Action: func() (*http.Request, error) {
				ctx := context.Background()

				return c.BuildGetAccountSubscriptionPlansRequest(ctx, nil)
			},
			Weight: 100,
		},
		"UpdateAccountSubscriptionPlan": {
			Name: "UpdateAccountSubscriptionPlan",
			Action: func() (*http.Request, error) {
				ctx := context.Background()

				if randomAccountSubscriptionPlan := fetchRandomAccountSubscriptionPlan(ctx, c); randomAccountSubscriptionPlan != nil {
					newAccountSubscriptionPlan := fakes.BuildFakeAccountSubscriptionPlanCreationInput()
					randomAccountSubscriptionPlan.Name = newAccountSubscriptionPlan.Name
					randomAccountSubscriptionPlan.Description = newAccountSubscriptionPlan.Description
					randomAccountSubscriptionPlan.Price = newAccountSubscriptionPlan.Price
					randomAccountSubscriptionPlan.Period = newAccountSubscriptionPlan.Period
					return c.BuildUpdateAccountSubscriptionPlanRequest(ctx, randomAccountSubscriptionPlan)
				}

				return nil, ErrUnavailableYet
			},
			Weight: 100,
		},
		"ArchiveAccountSubscriptionPlan": {
			Name: "ArchiveAccountSubscriptionPlan",
			Action: func() (*http.Request, error) {
				ctx := context.Background()

				randomAccountSubscriptionPlan := fetchRandomAccountSubscriptionPlan(ctx, c)
				if randomAccountSubscriptionPlan == nil {
					return nil, fmt.Errorf("retrieving random plan: %w", ErrUnavailableYet)
				}

				return c.BuildArchiveAccountSubscriptionPlanRequest(ctx, randomAccountSubscriptionPlan.ID)
			},
			Weight: 85,
		},
	}
}
