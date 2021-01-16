package fakes

import (
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	fake "github.com/brianvoe/gofakeit/v5"
)

// BuildFakePlan builds a faked plan.
func BuildFakePlan() *types.AccountSubscriptionPlan {
	return &types.AccountSubscriptionPlan{
		ID:          uint64(fake.Uint32()),
		Name:        fake.Password(true, true, false, false, false, 32),
		Description: fake.Word(),
		Price:       uint32(fake.Price(10, 20) * 100),
		Period:      24 * time.Hour * 30,
		CreatedOn:   uint64(uint32(fake.Date().Unix())),
	}
}

// BuildFakePlanList builds a faked AccountSubscriptionPlanList.
func BuildFakePlanList() *types.AccountSubscriptionPlanList {
	var examples []*types.AccountSubscriptionPlan
	for i := 0; i < exampleQuantity; i++ {
		examples = append(examples, BuildFakePlan())
	}

	return &types.AccountSubscriptionPlanList{
		Pagination: types.Pagination{
			Page:          1,
			Limit:         20,
			FilteredCount: exampleQuantity / 2,
			TotalCount:    exampleQuantity,
		},
		Plans: examples,
	}
}

// BuildFakePlanUpdateInputFromPlan builds a faked AccountSubscriptionPlanUpdateInput from an plan.
func BuildFakePlanUpdateInputFromPlan(plan *types.AccountSubscriptionPlan) *types.AccountSubscriptionPlanUpdateInput {
	return &types.AccountSubscriptionPlanUpdateInput{
		Name:        plan.Name,
		Description: plan.Description,
		Price:       plan.Price,
		Period:      plan.Period,
	}
}

// BuildFakePlanCreationInput builds a faked AccountSubscriptionPlanCreationInput.
func BuildFakePlanCreationInput() *types.AccountSubscriptionPlanCreationInput {
	plan := BuildFakePlan()
	return BuildFakePlanCreationInputFromPlan(plan)
}

// BuildFakePlanCreationInputFromPlan builds a faked AccountSubscriptionPlanCreationInput from an plan.
func BuildFakePlanCreationInputFromPlan(plan *types.AccountSubscriptionPlan) *types.AccountSubscriptionPlanCreationInput {
	return &types.AccountSubscriptionPlanCreationInput{
		Name:        plan.Name,
		Description: plan.Description,
		Price:       plan.Price,
		Period:      plan.Period,
	}
}
