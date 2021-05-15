package frontend

import (
	// import embed for the side effect.
	_ "embed"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
)

//go:embed assets/favicon.svg
var svgFaviconSrc []byte

func (s *service) favicon(res http.ResponseWriter, req *http.Request) {
	_, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	tracing.AttachRequestToSpan(span, req)

	res.Header().Set("Content-Type", "image/svg+xml")
	s.renderBytesToResponse(svgFaviconSrc, res)
}
