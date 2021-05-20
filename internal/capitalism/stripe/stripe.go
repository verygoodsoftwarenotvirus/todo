package stripe

import (
	"context"
	"errors"
	"fmt"
	"github.com/stripe/stripe-go/plan"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/capitalism"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
	"github.com/stripe/stripe-go/sub"
)

const (
	implementationName = "stripe_payment_manager"
)

type (
	PaymentManager struct {
		apiKey string
		logger logging.Logger
		tracer tracing.Tracer
	}
)

// NewStripePaymentManager builds a Stripe-backed PaymentManager
func NewStripePaymentManager(logger logging.Logger, apiKey string) *PaymentManager {
	stripe.Key = apiKey

	return &PaymentManager{
		apiKey: apiKey,
		logger: logging.EnsureLogger(logger),
		tracer: tracing.NewTracer(implementationName),
	}
}

func buildCustomerName(account *types.Account) string {
	return fmt.Sprintf("%s (%d)", account.Name, account.ID)
}

var errCustomerNotFound = errors.New("customer not found")

func (s *PaymentManager) findCustomerIDForAccount(ctx context.Context, account *types.Account) (string, error) {
	_, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue(keys.AccountIDKey, account.ID)

	params := &stripe.CustomerListParams{}
	//params.Filters.AddFilter("created", "gte", strconv.FormatUint(account.CreatedOn, 10))

	customers := customer.List(params)
	if err := customers.Err(); err != nil {
		return "", observability.PrepareError(err, logger, span, "listing <>")
	}

	for customers.Next() {
		c := customers.Customer()
		if c.Metadata[keys.AccountIDKey] == account.ExternalID {
			return c.ID, nil
		}
	}

	return "", errCustomerNotFound
}

func (s *PaymentManager) createCustomer(ctx context.Context, account *types.Account) (string, error) {
	_, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue(keys.AccountIDKey, account.ID)

	params := &stripe.CustomerParams{
		Name:    stripe.String(buildCustomerName(account)),
		Email:   stripe.String(account.ContactEmail),
		Phone:   stripe.String(account.ContactPhone),
		Address: &stripe.AddressParams{},
	}
	params.AddMetadata(keys.AccountIDKey, account.ExternalID)

	c, err := customer.New(params)
	if err != nil {
		return "", observability.PrepareError(err, logger, span, "creating customer")
	}

	return c.ID, nil
}

func (s *PaymentManager) GetCustomerID(ctx context.Context, account *types.Account) (string, error) {
	_, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue(keys.AccountIDKey, account.ID)

	customerID, err := s.findCustomerIDForAccount(ctx, account)
	if err != nil && !errors.Is(err, errCustomerNotFound) {
		observability.AcknowledgeError(err, logger, span, "ensuring customer")
		return "", err
	}

	if customerID != "" {
		return customerID, nil
	}

	customerID, err = s.createCustomer(ctx, account)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "ensuring customer")
		return "", err
	}

	return customerID, nil
}

func (s *PaymentManager) ListPlans(ctx context.Context) ([]capitalism.SubscriptionPlan, error) {
	_, span := s.tracer.StartSpan(ctx)
	defer span.End()

	params := &stripe.PlanListParams{}
	out := []capitalism.SubscriptionPlan{}

	plans := plan.List(params)
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

func (s *PaymentManager) findSubscriptionID(ctx context.Context, customerID, planID string) (string, error) {
	_, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue(keys.AccountSubscriptionPlanIDKey, planID)

	cus, err := customer.Get(customerID, nil)
	if err != nil {
		return "", observability.PrepareError(err, logger, span, "fetching customer")
	}

	var subscriptionID string
	if len(cus.Subscriptions.Data) > 0 && cus.Subscriptions.Data[0].Plan.ID == planID {
		subscriptionID = cus.Subscriptions.Data[0].ID
	}

	return subscriptionID, nil
}

func (s *PaymentManager) SubscribeToPlan(ctx context.Context, customerID, paymentMethodToken, planID string) (string, error) {
	_, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue(keys.AccountSubscriptionPlanIDKey, planID)

	if subscriptionID, err := s.findSubscriptionID(ctx, customerID, planID); err != nil {
		return "", observability.PrepareError(err, logger, span, "checking subscription status")
	} else if subscriptionID != "" {
		return subscriptionID, nil
	}

	params := &stripe.SubscriptionParams{
		Customer:      stripe.String(customerID),
		Plan:          stripe.String(planID),
		DefaultSource: stripe.String(paymentMethodToken),
	}

	subscription, err := sub.New(params)
	if err != nil {
		return "", observability.PrepareError(err, logger, span, "subscribing to plan")
	}

	return subscription.ID, nil
}

func (s *PaymentManager) UnsubscribeFromPlan(ctx context.Context, subscriptionID string) error {
	_, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue("subscription_id", subscriptionID)

	params := &stripe.SubscriptionCancelParams{
		InvoiceNow: stripe.Bool(true),
		Prorate:    stripe.Bool(true),
	}

	_, err := sub.Cancel(subscriptionID, params)
	if err != nil {
		return observability.PrepareError(err, logger, span, "unsubscribing from plan")
	}

	return nil
}
