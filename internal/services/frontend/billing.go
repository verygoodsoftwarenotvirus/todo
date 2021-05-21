package frontend

import (
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
)

func (s *service) handleCheckoutSessionStart(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	logger.Debug("checkout session route called")

	selectedPlan := validSubscriptionPlans[0].ID

	sessionID, err := s.paymentManager.CreateCheckoutSession(ctx, selectedPlan)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "error creating checkout session token")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger.WithValue("sessionID", sessionID).Debug("session id fetched")

	if _, err = res.Write([]byte(fmt.Sprintf(`{ "sessionID": %q }`, sessionID))); err != nil {
		return
	}
}

func (s *service) handleCheckoutSuccess(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	// implementation goes here

	logger.Debug("checkout session success route called")
}

func (s *service) handleCheckoutCancel(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	// implementation goes here

	logger.Debug("checkout session cancellation route called")
}

func (s *service) handleCheckoutFailure(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	// implementation goes here

	logger.Debug("checkout session failure route called")
}
