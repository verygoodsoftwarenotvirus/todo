package capitalism

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

var _ PaymentManager = (*NoopPaymentManager)(nil)

// NoopPaymentManager is a no-op payment manager.
type NoopPaymentManager struct{}

// CreateCustomerID satisfies our interface.
func (n *NoopPaymentManager) CreateCustomerID(_ context.Context, _ *types.Account) (string, error) {
	return "", nil
}

// HandleSubscriptionEventWebhook satisfies our interface.
func (n *NoopPaymentManager) HandleSubscriptionEventWebhook(_ *http.Request) error {
	return nil
}

// SubscribeToPlan satisfies our interface.
func (n *NoopPaymentManager) SubscribeToPlan(_ context.Context, _, _, _ string) (string, error) {
	return "", nil
}

// CreateCheckoutSession satisfies our interface.
func (n *NoopPaymentManager) CreateCheckoutSession(_ context.Context, _ string) (string, error) {
	return "", nil
}

// UnsubscribeFromPlan satisfies our interface.
func (n *NoopPaymentManager) UnsubscribeFromPlan(_ context.Context, _ string) error {
	return nil
}
