package frontend

import (
	// import embed for the side effect.
	_ "embed"
	"net/http"
)

//go:embed assets/favicon.svg
var svgFaviconSrc []byte

func (s *Service) favicon(res http.ResponseWriter, _ *http.Request) {
	res.Header().Set("Content-Type", "image/svg+xml")
	renderBytesToResponse(svgFaviconSrc, res)
}
