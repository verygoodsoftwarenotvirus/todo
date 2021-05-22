package stripe

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	testutil "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/capitalism"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/client"
	"github.com/stripe/stripe-go/webhook"
)

const (
	fakeAPIKey = "fake_api_key"
)

func buildTestPaymentManager(t *testing.T) *stripePaymentManager {
	t.Helper()

	logger := logging.NewNonOperationalLogger()

	pm := NewStripePaymentManager(logger, &capitalism.StripeConfig{})

	return pm.(*stripePaymentManager)
}

func TestNewStripePaymentManager(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNonOperationalLogger()
		pm := NewStripePaymentManager(logger, &capitalism.StripeConfig{})

		assert.NotNil(t, pm)
	})
}

func Test_stripePaymentManager_CreateCheckoutSession(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleAPIKey := fakeAPIKey
		pm := buildTestPaymentManager(t)

		exampleSubscriptionPlanID := "example_subscription_plan_id"
		expected := "example_session_id"

		mockAPIBackend := &mockBackend{}
		mockConnectBackend := &mockBackend{}
		mockUploadsBackend := &mockBackend{}

		mockAPIBackend.AnticipateCall(t, &stripe.CheckoutSession{ID: expected})
		mockAPIBackend.On(
			"Call",
			http.MethodPost,
			"/v1/checkout/sessions",
			exampleAPIKey,
			pm.buildCheckoutSessionParams(exampleSubscriptionPlanID),
			mock.IsType(&stripe.CheckoutSession{}),
		).Return(nil)

		mockedBackends := &stripe.Backends{
			API:     mockAPIBackend,
			Connect: mockConnectBackend,
			Uploads: mockUploadsBackend,
		}
		pm.client = client.New(exampleAPIKey, mockedBackends)

		actual, err := pm.CreateCheckoutSession(ctx, exampleSubscriptionPlanID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	T.Run("with error executing request", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleAPIKey := fakeAPIKey
		pm := buildTestPaymentManager(t)

		exampleSubscriptionPlanID := "example_subscription_plan_id"
		expected := "example_session_id"

		mockAPIBackend := &mockBackend{}
		mockConnectBackend := &mockBackend{}
		mockUploadsBackend := &mockBackend{}

		mockAPIBackend.AnticipateCall(t, &stripe.CheckoutSession{ID: expected})
		mockAPIBackend.On(
			"Call",
			http.MethodPost,
			"/v1/checkout/sessions",
			exampleAPIKey,
			pm.buildCheckoutSessionParams(exampleSubscriptionPlanID),
			mock.IsType(&stripe.CheckoutSession{}),
		).Return(errors.New("blah"))

		mockedBackends := &stripe.Backends{
			API:     mockAPIBackend,
			Connect: mockConnectBackend,
			Uploads: mockUploadsBackend,
		}
		pm.client = client.New(exampleAPIKey, mockedBackends)

		actual, err := pm.CreateCheckoutSession(ctx, exampleSubscriptionPlanID)
		assert.Error(t, err)
		assert.Empty(t, actual)
	})
}

func Test_stripePaymentManager_HandleSubscriptionEventWebhook(T *testing.T) {
	T.Parallel()

	T.Run("standard with webhookEventTypeCheckoutCompleted", func(t *testing.T) {
		t.Parallel()

		pm := buildTestPaymentManager(t)
		pm.webhookSecret = "example_webhook_secret"

		testEventTypes := []string{
			webhookEventTypeCheckoutCompleted,
			webhookEventTypeInvoicePaid,
			webhookEventTypeInvoicePaymentFailed,
			t.Name(),
		}

		for _, et := range testEventTypes {
			exampleEvent := &stripe.Event{
				Account: "whatever",
				Created: time.Now().Unix(),
				Data: &stripe.EventData{
					Object: map[string]interface{}{
						"things": "stuff",
					},
					PreviousAttributes: map[string]interface{}{
						"things": "stuff",
					},
				},
				ID:   "example",
				Type: et,
			}

			var b bytes.Buffer
			require.NoError(t, json.NewEncoder(&b).Encode(exampleEvent))

			mac := hmac.New(sha256.New, []byte(pm.webhookSecret))
			_, err := mac.Write(b.Bytes())
			require.NoError(t, err)

			now := time.Now()
			exampleSig := webhook.ComputeSignature(now, b.Bytes(), pm.webhookSecret)
			exampleSignature := fmt.Sprintf("t=%d,v1=%s", now.Unix(), hex.EncodeToString(exampleSig))

			req := httptest.NewRequest(http.MethodPost, "/webhook_update", bytes.NewReader(b.Bytes()))
			req.Header.Set(webhookHeaderName, exampleSignature)

			err = pm.HandleSubscriptionEventWebhook(req)
			assert.NoError(t, err)
		}
	})

	T.Run("with invalid body", func(t *testing.T) {
		t.Parallel()

		pm := buildTestPaymentManager(t)
		pm.webhookSecret = "example_webhook_secret"

		exampleEvent := &stripe.Event{
			Account: "whatever",
			Created: time.Now().Unix(),
			Data: &stripe.EventData{
				Object: map[string]interface{}{
					"things": "stuff",
				},
				PreviousAttributes: map[string]interface{}{
					"things": "stuff",
				},
			},
			ID:   "example",
			Type: webhookEventTypeCheckoutCompleted,
		}

		var b bytes.Buffer
		require.NoError(t, json.NewEncoder(&b).Encode(exampleEvent))

		mac := hmac.New(sha256.New, []byte(pm.webhookSecret))
		_, err := mac.Write(b.Bytes())
		require.NoError(t, err)

		mrc := &testutil.MockReadCloser{}
		mrc.On("Read", mock.IsType([]byte(""))).Return(int(0), errors.New("blah"))

		req := httptest.NewRequest(http.MethodPost, "/webhook_update", mrc)
		req.Header.Set(webhookHeaderName, "bad-sig")

		err = pm.HandleSubscriptionEventWebhook(req)
		assert.Error(t, err)
	})

	T.Run("with invalid signature", func(t *testing.T) {
		t.Parallel()

		pm := buildTestPaymentManager(t)

		req := httptest.NewRequest(http.MethodPost, "/webhook_update", nil)
		req.Header.Set(webhookHeaderName, "bad-sig")

		err := pm.HandleSubscriptionEventWebhook(req)
		assert.Error(t, err)
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

	T.Run("with error executing request", func(t *testing.T) {
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
		).Return(errors.New("blah"))

		mockedBackends := &stripe.Backends{
			API:     mockAPIBackend,
			Connect: mockConnectBackend,
			Uploads: mockUploadsBackend,
		}
		pm.client = client.New(exampleAPIKey, mockedBackends)

		actual, err := pm.CreateCustomerID(ctx, exampleAccount)
		assert.Error(t, err)
		assert.Empty(t, actual)

		mock.AssertExpectationsForObjects(t, mockAPIBackend, mockConnectBackend, mockUploadsBackend)
	})
}

func Test_stripePaymentManager_findSubscriptionID(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleAPIKey := fakeAPIKey
		exampleCustomerID := "fake_customer"
		examplePlanID := "fake_plan"
		pm := buildTestPaymentManager(t)

		mockAPIBackend := &mockBackend{}
		mockConnectBackend := &mockBackend{}
		mockUploadsBackend := &mockBackend{}

		expected := "fake_subscription_id"

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

		actual, err := pm.findSubscriptionID(ctx, exampleCustomerID, examplePlanID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, mockAPIBackend, mockConnectBackend, mockUploadsBackend)
	})
}

func Test_stripePaymentManager_SubscribeToPlan(T *testing.T) {
	T.Parallel()

	T.Run("with pre-existing subscription", func(t *testing.T) {
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

	T.Run("with error checking pre-existing subscription", func(t *testing.T) {
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
		).Return(errors.New("blah"))

		mockedBackends := &stripe.Backends{
			API:     mockAPIBackend,
			Connect: mockConnectBackend,
			Uploads: mockUploadsBackend,
		}
		pm.client = client.New(fakeAPIKey, mockedBackends)

		actual, err := pm.SubscribeToPlan(ctx, exampleCustomerID, examplePaymentMethodToken, examplePlanID)
		assert.Error(t, err)
		assert.Empty(t, actual)

		mock.AssertExpectationsForObjects(t, mockAPIBackend, mockConnectBackend, mockUploadsBackend)
	})

	T.Run("without pre-existing subscription", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleAPIKey := fakeAPIKey
		exampleCustomerID := "fake_customer"
		exampleSubscriptionID := "fake_subscription"
		examplePlanID := "fake_plan"
		examplePaymentMethodToken := "fake_payment_token"
		expected := "fake_subscription"
		pm := buildTestPaymentManager(t)

		expectedCustomer := &stripe.Customer{
			ID: exampleCustomerID,
			Subscriptions: &stripe.SubscriptionList{
				Data: []*stripe.Subscription{},
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

		expectedSubscription := &stripe.Subscription{
			ID: exampleSubscriptionID,
		}

		mockAPIBackend.AnticipateCall(t, expectedSubscription)
		mockAPIBackend.On(
			"Call",
			http.MethodPost,
			"/v1/subscriptions",
			exampleAPIKey,
			mock.IsType(&stripe.SubscriptionParams{}),
			mock.IsType(&stripe.Subscription{}),
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

	T.Run("without pre-existing subscription and with error creating subscription", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleAPIKey := fakeAPIKey
		exampleCustomerID := "fake_customer"
		exampleSubscriptionID := "fake_subscription"
		examplePlanID := "fake_plan"
		examplePaymentMethodToken := "fake_payment_token"
		pm := buildTestPaymentManager(t)

		expectedCustomer := &stripe.Customer{
			ID: exampleCustomerID,
			Subscriptions: &stripe.SubscriptionList{
				Data: []*stripe.Subscription{},
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

		expectedSubscription := &stripe.Subscription{
			ID: exampleSubscriptionID,
		}

		mockAPIBackend.AnticipateCall(t, expectedSubscription)
		mockAPIBackend.On(
			"Call",
			http.MethodPost,
			"/v1/subscriptions",
			exampleAPIKey,
			mock.IsType(&stripe.SubscriptionParams{}),
			mock.IsType(&stripe.Subscription{}),
		).Return(errors.New("blah"))

		mockedBackends := &stripe.Backends{
			API:     mockAPIBackend,
			Connect: mockConnectBackend,
			Uploads: mockUploadsBackend,
		}
		pm.client = client.New(fakeAPIKey, mockedBackends)

		actual, err := pm.SubscribeToPlan(ctx, exampleCustomerID, examplePaymentMethodToken, examplePlanID)
		assert.Error(t, err)
		assert.Empty(t, actual)

		mock.AssertExpectationsForObjects(t, mockAPIBackend, mockConnectBackend, mockUploadsBackend)
	})
}

func Test_stripePaymentManager_UnsubscribeFromPlan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleAPIKey := fakeAPIKey
		exampleSubscriptionID := "fake_subscription_id"
		pm := buildTestPaymentManager(t)

		mockAPIBackend := &mockBackend{}
		mockConnectBackend := &mockBackend{}
		mockUploadsBackend := &mockBackend{}

		expectedCustomer := &stripe.Customer{
			Subscriptions: &stripe.SubscriptionList{
				Data: []*stripe.Subscription{
					{
						ID: exampleSubscriptionID,
					},
				},
			},
		}

		mockAPIBackend.AnticipateCall(t, expectedCustomer)
		mockAPIBackend.On(
			"Call",
			http.MethodDelete,
			fmt.Sprintf("/v1/subscriptions/%s", exampleSubscriptionID),
			exampleAPIKey,
			buildCancellationParams(),
			mock.IsType(&stripe.Subscription{}),
		).Return(nil)

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

	T.Run("with error executing request", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleAPIKey := fakeAPIKey
		exampleSubscriptionID := "fake_subscription_id"
		pm := buildTestPaymentManager(t)

		mockAPIBackend := &mockBackend{}
		mockConnectBackend := &mockBackend{}
		mockUploadsBackend := &mockBackend{}

		expectedCustomer := &stripe.Customer{
			Subscriptions: &stripe.SubscriptionList{
				Data: []*stripe.Subscription{
					{
						ID: exampleSubscriptionID,
					},
				},
			},
		}

		mockAPIBackend.AnticipateCall(t, expectedCustomer)
		mockAPIBackend.On(
			"Call",
			http.MethodDelete,
			fmt.Sprintf("/v1/subscriptions/%s", exampleSubscriptionID),
			exampleAPIKey,
			buildCancellationParams(),
			mock.IsType(&stripe.Subscription{}),
		).Return(errors.New("blah"))

		mockedBackends := &stripe.Backends{
			API:     mockAPIBackend,
			Connect: mockConnectBackend,
			Uploads: mockUploadsBackend,
		}
		pm.client = client.New(fakeAPIKey, mockedBackends)

		err := pm.UnsubscribeFromPlan(ctx, exampleSubscriptionID)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockAPIBackend, mockConnectBackend, mockUploadsBackend)
	})
}
