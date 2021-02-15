package httpserver

import (
	"fmt"
	"net/http"
	"regexp"
)

var idReplacementRegex = regexp.MustCompile(`[^(v)]\\d+`)

func formatSpanNameForRequest(operation string, req *http.Request) string {
	return fmt.Sprintf(
		"%s %s: %s",
		req.Method,
		idReplacementRegex.ReplaceAllString(req.URL.Path, "/{id}"),
		operation,
	)
}
