package frontend

import (
	"fmt"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/panicking"
)

var pricePanicker = panicking.NewProductionPanicker()

const (
	priceDivisor      = 0.01
	arbitraryPriceMax = 100000
)

// NOTE: this function panics when it receives a number > 100,000, as it is not meant to handle those kinds of prices.
func renderPrice(p uint32) string {
	if p >= arbitraryPriceMax {
		pricePanicker.Panic("price to be rendered is too large!")
	}

	return fmt.Sprintf("$%.2f", float64(p)*priceDivisor)
}

func (s *service) handleCheckoutSessionStart(res http.ResponseWriter, req *http.Request) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	logger.Debug("checkout session route called")

	selectedPlan := req.URL.Query().Get("plan")
	if selectedPlan == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	sessionID, err := s.paymentManager.CreateCheckoutSession(ctx, selectedPlan)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "error creating checkout session token")
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	// implementation goes here

	logger.WithValue("sessionID", sessionID).Debug("session id fetched")
}

func (s *service) handleCheckoutSuccess(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	// implementation goes here

	logger.Debug("checkout session success route called")

	res.WriteHeader(http.StatusTooEarly)
}

func (s *service) handleCheckoutCancel(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	// implementation goes here

	logger.Debug("checkout session cancellation route called")

	res.WriteHeader(http.StatusTooEarly)
}

func (s *service) handleCheckoutFailure(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithRequest(req)
	tracing.AttachRequestToSpan(span, req)

	// implementation goes here

	logger.Debug("checkout session failure route called")

	res.WriteHeader(http.StatusTooEarly)
}
