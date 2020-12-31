package fakes

import (
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	fake "github.com/brianvoe/gofakeit/v5"
)

// BuildFakePlan builds a faked plan.
func BuildFakePlan() *types.Plan {
	return &types.Plan{
		ID:          uint64(fake.Uint32()),
		Name:        fake.Password(true, true, false, false, false, 32),
		Description: fake.Word(),
		Price:       uint32(fake.Price(10, 20) * 100),
		Period:      24 * time.Hour * 30,
		CreatedOn:   uint64(uint32(fake.Date().Unix())),
	}
}

// BuildFakePlanList builds a faked PlanList.
func BuildFakePlanList() *types.PlanList {
	var examples []types.Plan
	for i := 0; i < exampleQuantity; i++ {
		examples = append(examples, *BuildFakePlan())
	}

	return &types.PlanList{
		Pagination: types.Pagination{
			Page:       1,
			Limit:      20,
			TotalCount: exampleQuantity,
		},
		Plans: examples,
	}
}

// BuildFakePlanUpdateInputFromPlan builds a faked PlanUpdateInput from an plan.
func BuildFakePlanUpdateInputFromPlan(plan *types.Plan) *types.PlanUpdateInput {
	return &types.PlanUpdateInput{
		Name:        plan.Name,
		Description: plan.Description,
		Price:       plan.Price,
		Period:      plan.Period,
	}
}

// BuildFakePlanCreationInput builds a faked PlanCreationInput.
func BuildFakePlanCreationInput() *types.PlanCreationInput {
	plan := BuildFakePlan()
	return BuildFakePlanCreationInputFromPlan(plan)
}

// BuildFakePlanCreationInputFromPlan builds a faked PlanCreationInput from an plan.
func BuildFakePlanCreationInputFromPlan(plan *types.Plan) *types.PlanCreationInput {
	return &types.PlanCreationInput{
		Name:        plan.Name,
		Description: plan.Description,
		Price:       plan.Price,
		Period:      plan.Period,
	}
}
