package stripe

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/client"
	"github.com/stripe/stripe-go/form"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/capitalism"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
	"net/http"
	"testing"
)

const (
	fakeAPIKey = "fake_api_key"
)

func buildTestPaymentManager(t *testing.T) *stripePaymentManager {
	t.Helper()

	logger := logging.NewNonOperationalLogger()

	pm := NewStripePaymentManager(logger, fakeAPIKey)

	return pm.(*stripePaymentManager)
}

func TestNewStripePaymentManager(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNonOperationalLogger()
		pm := NewStripePaymentManager(logger, fakeAPIKey)

		assert.NotNil(t, pm)
	})
}

func Test_buildCustomerName(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		account := &types.Account{Name: "example", ID: 123}

		expected := "example (123)"
		actual := buildCustomerName(account)

		assert.Equal(t, expected, actual)
	})
}

func Test_stripePaymentManager_GetCustomerID(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleAccount := fakes.BuildFakeAccount()
		exampleAPIKey := fakeAPIKey
		pm := buildTestPaymentManager(t)

		expected := "fake_customer_id"

		mockAPIBackend := &mockBackend{}
		mockConnectBackend := &mockBackend{}
		mockUploadsBackend := &mockBackend{}

		mockAPIBackend.AnticipateCall(t, &stripe.Customer{ID: expected})
		mockAPIBackend.On(
			"Call",
			http.MethodPost,
			"/v1/customers",
			exampleAPIKey,
			buildGetCustomerParams(exampleAccount),
			mock.IsType(&stripe.Customer{}),
		).Return(nil)

		mockedBackends := &stripe.Backends{
			API:     mockAPIBackend,
			Connect: mockConnectBackend,
			Uploads: mockUploadsBackend,
		}
		pm.client = client.New(exampleAPIKey, mockedBackends)

		actual, err := pm.CreateCustomerID(ctx, exampleAccount)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, mockAPIBackend, mockConnectBackend, mockUploadsBackend)
	})
}

func Test_stripePaymentManager_ListPlans(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		expectedPlanID := "expected_plan_id"
		exampleAPIKey := fakeAPIKey
		pm := buildTestPaymentManager(t)

		exampleStripePlans := []*stripe.Plan{
			{
				ID: expectedPlanID,
			},
		}

		expected := []capitalism.SubscriptionPlan{
			{
				ID: expectedPlanID,
			},
		}

		mockAPIBackend := &mockBackend{}
		mockConnectBackend := &mockBackend{}
		mockUploadsBackend := &mockBackend{}

		mockAPIBackend.AnticipateCall(t, &stripe.PlanList{Data: exampleStripePlans})
		mockAPIBackend.On(
			"CallRaw",
			http.MethodGet,
			"/v1/plans",
			exampleAPIKey,
			mock.IsType(&form.Values{}),
			mock.IsType(&stripe.Params{}),
			mock.IsType(&stripe.PlanList{}),
		).Return(nil)

		mockedBackends := &stripe.Backends{
			API:     mockAPIBackend,
			Connect: mockConnectBackend,
			Uploads: mockUploadsBackend,
		}
		pm.client = client.New(fakeAPIKey, mockedBackends)

		actual, err := pm.ListPlans(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, mockAPIBackend, mockConnectBackend, mockUploadsBackend)
	})
}

func Test_stripePaymentManager_SubscribeToPlan(T *testing.T) {
	T.Parallel()

	T.Run("with pre-existening subscription", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleAPIKey := fakeAPIKey
		exampleCustomerID := "fake_customer"
		examplePlanID := "fake_plan"
		examplePaymentMethodToken := "fake_payment_token"
		expected := "fake_subscription"
		pm := buildTestPaymentManager(t)

		expectedCustomer := &stripe.Customer{
			ID: exampleCustomerID,
			Subscriptions: &stripe.SubscriptionList{
				Data: []*stripe.Subscription{
					{
						ID: expected,
						Plan: &stripe.Plan{
							ID: examplePlanID,
						},
					},
				},
			},
		}

		mockAPIBackend := &mockBackend{}
		mockConnectBackend := &mockBackend{}
		mockUploadsBackend := &mockBackend{}

		mockAPIBackend.AnticipateCall(t, expectedCustomer)
		mockAPIBackend.On(
			"Call",
			http.MethodGet,
			fmt.Sprintf("/v1/customers/%s", exampleCustomerID),
			exampleAPIKey,
			mock.IsType(&stripe.CustomerParams{}),
			mock.IsType(&stripe.Customer{}),
		).Return(nil)

		mockedBackends := &stripe.Backends{
			API:     mockAPIBackend,
			Connect: mockConnectBackend,
			Uploads: mockUploadsBackend,
		}
		pm.client = client.New(fakeAPIKey, mockedBackends)

		actual, err := pm.SubscribeToPlan(ctx, exampleCustomerID, examplePaymentMethodToken, examplePlanID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, mockAPIBackend, mockConnectBackend, mockUploadsBackend)
	})

	T.Run("without pre-existening subscription", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleAPIKey := fakeAPIKey
		exampleCustomerID := "fake_customer"
		examplePlanID := "fake_plan"
		examplePaymentMethodToken := "fake_payment_token"
		expected := "fake_subscription"
		pm := buildTestPaymentManager(t)

		expectedCustomer := &stripe.Customer{
			ID: exampleCustomerID,
			Subscriptions: &stripe.SubscriptionList{
				Data: []*stripe.Subscription{
					{
						ID: expected,
						Plan: &stripe.Plan{
							ID: examplePlanID,
						},
					},
				},
			},
		}

		mockAPIBackend := &mockBackend{}
		mockConnectBackend := &mockBackend{}
		mockUploadsBackend := &mockBackend{}

		mockAPIBackend.AnticipateCall(t, expectedCustomer)
		mockAPIBackend.On(
			"Call",
			http.MethodGet,
			fmt.Sprintf("/v1/customers/%s", exampleCustomerID),
			exampleAPIKey,
			mock.IsType(&stripe.CustomerParams{}),
			mock.IsType(&stripe.Customer{}),
		).Return(nil)

		mockedBackends := &stripe.Backends{
			API:     mockAPIBackend,
			Connect: mockConnectBackend,
			Uploads: mockUploadsBackend,
		}
		pm.client = client.New(fakeAPIKey, mockedBackends)

		actual, err := pm.SubscribeToPlan(ctx, exampleCustomerID, examplePaymentMethodToken, examplePlanID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, mockAPIBackend, mockConnectBackend, mockUploadsBackend)
	})
}

func Test_stripePaymentManager_UnsubscribeFromPlan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleSubscriptionID := "fake_subscription_id"
		pm := buildTestPaymentManager(t)

		mockAPIBackend := &mockBackend{}
		mockConnectBackend := &mockBackend{}
		mockUploadsBackend := &mockBackend{}

		mockedBackends := &stripe.Backends{
			API:     mockAPIBackend,
			Connect: mockConnectBackend,
			Uploads: mockUploadsBackend,
		}
		pm.client = client.New(fakeAPIKey, mockedBackends)

		err := pm.UnsubscribeFromPlan(ctx, exampleSubscriptionID)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockAPIBackend, mockConnectBackend, mockUploadsBackend)

	})
}

func Test_stripePaymentManager_findSubscriptionID(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleCustomerID := "fake_customer"
		examplePlanID := "fake_plan"
		pm := buildTestPaymentManager(t)

		mockAPIBackend := &mockBackend{}
		mockConnectBackend := &mockBackend{}
		mockUploadsBackend := &mockBackend{}

		mockedBackends := &stripe.Backends{
			API:     mockAPIBackend,
			Connect: mockConnectBackend,
			Uploads: mockUploadsBackend,
		}
		pm.client = client.New(fakeAPIKey, mockedBackends)

		expected := "fake_subscription_id"

		actual, err := pm.findSubscriptionID(ctx, exampleCustomerID, examplePlanID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, mockAPIBackend, mockConnectBackend, mockUploadsBackend)
	})
}
