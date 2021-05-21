package capitalism

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

type (
	SubscriptionPlan struct {
		ID    string
		Name  string
		Price uint32
	}

	// PaymentManager handles payments via 3rd-party providers.
	PaymentManager interface {
		CreateCustomerID(ctx context.Context, account *types.Account) (string, error)
		ListPlans(ctx context.Context) ([]SubscriptionPlan, error)
		SubscribeToPlan(ctx context.Context, customerID, paymentMethodToken, planID string) (string, error)
		CreateCheckoutSession(ctx context.Context, subscriptionPlanID string) (string, error)
		UnsubscribeFromPlan(ctx context.Context, subscriptionID string) error
	}
)
