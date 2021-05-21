package stripe

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/capitalism"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/client"
	"github.com/stripe/stripe-go/webhook"
)

const (
	implementationName = "stripe_payment_manager"
)

type (
	WebhookSecret string
	APIKey        string

	stripePaymentManager struct {
		client        *client.API
		successURL    string
		cancelURL     string
		webhookSecret string
		logger        logging.Logger
		tracer        tracing.Tracer
	}
)

// NewStripePaymentManager builds a Stripe-backed stripePaymentManager
func NewStripePaymentManager(logger logging.Logger, cfg *capitalism.StripeConfig) capitalism.PaymentManager {
	return &stripePaymentManager{
		client:        client.New(cfg.APIKey, nil),
		webhookSecret: cfg.WebhookSecret,
		successURL:    cfg.SuccessURL,
		cancelURL:     cfg.CancelURL,
		logger:        logging.EnsureLogger(logger),
		tracer:        tracing.NewTracer(implementationName),
	}
}

func buildCustomerName(account *types.Account) string {
	return fmt.Sprintf("%s (%d)", account.Name, account.ID)
}

func buildGetCustomerParams(a *types.Account) *stripe.CustomerParams {
	p := &stripe.CustomerParams{
		Name:    stripe.String(buildCustomerName(a)),
		Email:   stripe.String(a.ContactEmail),
		Phone:   stripe.String(a.ContactPhone),
		Address: &stripe.AddressParams{},
	}
	p.AddMetadata(keys.AccountIDKey, a.ExternalID)

	return p
}

func (s *stripePaymentManager) CreateCheckoutSession(ctx context.Context, subscriptionPlanID string) (string, error) {
	_, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue(keys.AccountSubscriptionPlanIDKey, subscriptionPlanID)

	params := &stripe.CheckoutSessionParams{
		SuccessURL:         stripe.String(s.successURL),
		CancelURL:          stripe.String(s.cancelURL),
		Mode:               stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Items: []*stripe.CheckoutSessionSubscriptionDataItemsParams{
				{
					Plan:     stripe.String(subscriptionPlanID),
					Quantity: stripe.Int64(1), // For metered billing, do not pass quantity
				},
			},
		},
	}

	sess, err := s.client.CheckoutSessions.New(params)
	if err != nil {
		return "", observability.PrepareError(err, logger, span, "creating checkout session")
	}

	return sess.ID, nil
}

const (
	webhookEventTypeCheckoutCompleted    = "checkout.session.completed"
	webhookEventTypeInvoicePaid          = "invoice.paid"
	webhookEventTypeInvoicePaymentFailed = "invoice.payment_failed"
)

func (s *stripePaymentManager) handleWebhook(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	if req.Method != "POST" {
		http.Error(res, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	logger := s.logger.WithRequest(req)

	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		observability.AcknowledgeError(err, logger, span, "parsing received webhook content")
		return
	}

	event, err := webhook.ConstructEvent(b, req.Header.Get("Stripe-Signature"), s.webhookSecret)

	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		log.Printf("webhook.ConstructEvent: %v", err)
		return
	}

	switch event.Type {
	case webhookEventTypeCheckoutCompleted:
		// Payment is successful, and the subscription is created.
		// You should provision the subscription and save the customer ID to your database.
	case webhookEventTypeInvoicePaid:
		// Continue to provision the subscription as payments continue to be made.
		// Store the status in your database and check when a user accesses your service.
		// This approach helps you avoid hitting rate limits.
	case webhookEventTypeInvoicePaymentFailed:
		// The payment failed, or the customer does not have a valid payment method.
		// The subscription becomes past_due. Notify your customer and send them to the
		// customer portal to update their payment information.
	default:
		// unhandled event type
	}
}

func (s *stripePaymentManager) CreateCustomerID(ctx context.Context, account *types.Account) (string, error) {
	_, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue(keys.AccountIDKey, account.ID)

	params := buildGetCustomerParams(account)

	c, err := s.client.Customers.New(params)
	if err != nil {
		return "", observability.PrepareError(err, logger, span, "creating customer")
	}

	return c.ID, nil
}

func (s *stripePaymentManager) ListPlans(ctx context.Context) ([]capitalism.SubscriptionPlan, error) {
	_, span := s.tracer.StartSpan(ctx)
	defer span.End()

	params := &stripe.PlanListParams{}
	out := []capitalism.SubscriptionPlan{}

	plans := s.client.Plans.List(params)
	if err := plans.Err(); err != nil {
		return nil, observability.PrepareError(err, s.logger, span, "listing <>")
	}

	for plans.Next() {
		p := plans.Plan()
		out = append(out, capitalism.SubscriptionPlan{
			ID:    p.ID,
			Name:  p.Nickname,
			Price: uint32(p.Amount),
		})
	}

	return out, nil
}

var errSubscriptionNotFound = errors.New("subscription not found")

func (s *stripePaymentManager) findSubscriptionID(ctx context.Context, customerID, planID string) (string, error) {
	_, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue(keys.AccountSubscriptionPlanIDKey, planID)

	cus, err := s.client.Customers.Get(customerID, nil)
	if err != nil {
		return "", observability.PrepareError(err, logger, span, "fetching customer")
	}

	for _, sub := range cus.Subscriptions.Data {
		if sub.Plan.ID == planID {
			return sub.ID, nil
		}
	}

	return "", errSubscriptionNotFound
}

func (s *stripePaymentManager) SubscribeToPlan(ctx context.Context, customerID, paymentMethodToken, planID string) (string, error) {
	_, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue(keys.AccountSubscriptionPlanIDKey, planID)

	// check first if the plan is already implemented.
	subscriptionID, err := s.findSubscriptionID(ctx, customerID, planID)
	if err != nil && !errors.Is(err, errSubscriptionNotFound) {
		return "", observability.PrepareError(err, logger, span, "checking subscription status")
	} else if subscriptionID != "" {
		return subscriptionID, nil
	}

	params := &stripe.SubscriptionParams{
		Customer:      stripe.String(customerID),
		Plan:          stripe.String(planID),
		DefaultSource: stripe.String(paymentMethodToken),
	}

	subscription, err := s.client.Subscriptions.New(params)
	if err != nil {
		return "", observability.PrepareError(err, logger, span, "subscribing to plan")
	}

	return subscription.ID, nil
}

func (s *stripePaymentManager) UnsubscribeFromPlan(ctx context.Context, subscriptionID string) error {
	_, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue("subscription_id", subscriptionID)

	params := &stripe.SubscriptionCancelParams{
		InvoiceNow: stripe.Bool(true),
		Prorate:    stripe.Bool(true),
	}

	if _, err := s.client.Subscriptions.Cancel(subscriptionID, params); err != nil {
		return observability.PrepareError(err, logger, span, "unsubscribing from plan")
	}

	return nil
}
