package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/httpclient"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
)

// fetchRandomPlan retrieves a random plan from the list of available plans.
func fetchRandomPlan(ctx context.Context, c *httpclient.V1Client) *types.Plan {
	plansRes, err := c.GetPlans(ctx, nil)
	if err != nil || plansRes == nil || len(plansRes.Plans) == 0 {
		return nil
	}

	randIndex := rand.Intn(len(plansRes.Plans))

	return &plansRes.Plans[randIndex]
}

func buildPlanActions(shouldUse bool, c *httpclient.V1Client) map[string]*Action {
	if shouldUse {
		return nil
	}

	return map[string]*Action{
		"CreatePlan": {
			Name: "CreatePlan",
			Action: func() (*http.Request, error) {
				ctx := context.Background()

				planInput := fakes.BuildFakePlanCreationInput()

				return c.BuildCreatePlanRequest(ctx, planInput)
			},
			Weight: 100,
		},
		"GetPlan": {
			Name: "GetPlan",
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
		"GetPlans": {
			Name: "GetPlans",
			Action: func() (*http.Request, error) {
				ctx := context.Background()

				return c.BuildGetPlansRequest(ctx, nil)
			},
			Weight: 100,
		},
		"UpdatePlan": {
			Name: "UpdatePlan",
			Action: func() (*http.Request, error) {
				ctx := context.Background()

				if randomPlan := fetchRandomPlan(ctx, c); randomPlan != nil {
					newPlan := fakes.BuildFakePlanCreationInput()
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
		"ArchivePlan": {
			Name: "ArchivePlan",
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
