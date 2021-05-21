package capitalism

import (
	"context"
	"net/http"

	"github.com/stretchr/testify/mock"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

var _ PaymentManager = (*mockPaymentManager)(nil)

type mockPaymentManager struct {
	mock.Mock
}

func (m *mockPaymentManager) HandleSubscriptionEventWebhook(req *http.Request) error {
	return m.Called(req).Error(0)
}

// NewMockPaymentManager returns a mockable capitalism.PaymentManager.
func NewMockPaymentManager() *mockPaymentManager {
	return &mockPaymentManager{}
}

func (m *mockPaymentManager) CreateCustomerID(ctx context.Context, account *types.Account) (string, error) {
	returnValues := m.Called(ctx, account)

	return returnValues.String(0), returnValues.Error(1)
}

func (m *mockPaymentManager) ListPlans(ctx context.Context) ([]SubscriptionPlan, error) {
	returnValues := m.Called(ctx)

	return returnValues.Get(0).([]SubscriptionPlan), returnValues.Error(1)
}

func (m *mockPaymentManager) SubscribeToPlan(ctx context.Context, customerID, paymentMethodToken, planID string) (string, error) {
	returnValues := m.Called(ctx, customerID, paymentMethodToken, planID)

	return returnValues.String(0), returnValues.Error(1)
}

func (m *mockPaymentManager) CreateCheckoutSession(ctx context.Context, subscriptionPlanID string) (string, error) {
	returnValues := m.Called(ctx, subscriptionPlanID)

	return returnValues.String(0), returnValues.Error(1)
}

func (m *mockPaymentManager) UnsubscribeFromPlan(ctx context.Context, subscriptionID string) error {
	return m.Called(ctx, subscriptionID).Error(0)
}
