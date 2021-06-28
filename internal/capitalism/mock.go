package capitalism

import (
	"context"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/stretchr/testify/mock"
)

var _ PaymentManager = (*MockPaymentManager)(nil)

// MockPaymentManager is a mockable capitalism.PaymentManager.
type MockPaymentManager struct {
	mock.Mock
}

// NewMockPaymentManager returns a mockable capitalism.PaymentManager.
func NewMockPaymentManager() *MockPaymentManager {
	return &MockPaymentManager{}
}

// HandleSubscriptionEventWebhook satisfies our interface contract.
func (m *MockPaymentManager) HandleSubscriptionEventWebhook(req *http.Request) error {
	return m.Called(req).Error(0)
}

// CreateCustomerID satisfies our interface contract.
func (m *MockPaymentManager) CreateCustomerID(ctx context.Context, account *types.Account) (string, error) {
	returnValues := m.Called(ctx, account)

	return returnValues.String(0), returnValues.Error(1)
}

// ListPlans satisfies our interface contract.
func (m *MockPaymentManager) ListPlans(ctx context.Context) ([]SubscriptionPlan, error) {
	returnValues := m.Called(ctx)

	return returnValues.Get(0).([]SubscriptionPlan), returnValues.Error(1)
}

// SubscribeToPlan satisfies our interface contract.
func (m *MockPaymentManager) SubscribeToPlan(ctx context.Context, customerID, paymentMethodToken, planID string) (string, error) {
	returnValues := m.Called(ctx, customerID, paymentMethodToken, planID)

	return returnValues.String(0), returnValues.Error(1)
}

// CreateCheckoutSession satisfies our interface contract.
func (m *MockPaymentManager) CreateCheckoutSession(ctx context.Context, subscriptionPlanID string) (string, error) {
	returnValues := m.Called(ctx, subscriptionPlanID)

	return returnValues.String(0), returnValues.Error(1)
}

// UnsubscribeFromPlan satisfies our interface contract.
func (m *MockPaymentManager) UnsubscribeFromPlan(ctx context.Context, subscriptionID string) error {
	return m.Called(ctx, subscriptionID).Error(0)
}
