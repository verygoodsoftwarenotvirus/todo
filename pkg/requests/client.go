package requests

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/panicking"
)

const (
	clientName = "todo_client_v1"
)

// Builder is a client for interacting with v1 of our HTTP API.
type Builder struct {
	logger logging.Logger
	tracer tracing.Tracer
	url    *url.URL

	panicker    panicking.Panicker
	contentType string
	debug       bool
}

// URL provides the client's URL.
func (b *Builder) URL() *url.URL {
	return b.url
}

// SetURL provides the client's URL.
func (b *Builder) SetURL(u *url.URL) error {
	if u == nil {
		return ErrNoURLProvided
	}

	b.url = u

	return nil
}

// NewBuilder builds a new API client for us.
func NewBuilder(u *url.URL) (*Builder, error) {
	l := logging.NewNonOperationalLogger()

	if u == nil {
		return nil, ErrNoURLProvided
	}

	c := &Builder{
		url:         u,
		debug:       false,
		contentType: "application/json",
		panicker:    panicking.NewProductionPanicker(),
		logger:      l,
		tracer:      tracing.NewTracer(clientName),
	}

	return c, nil
}

// BuildURL builds standard service URLs.
func (b *Builder) BuildURL(ctx context.Context, qp url.Values, parts ...string) string {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if u := b.buildRawURL(ctx, qp, parts...); u != nil {
		return u.String()
	}

	return ""
}

func buildRawURL(u *url.URL, qp url.Values, includeVersionPrefix bool, parts ...string) (*url.URL, error) {
	tu := *u

	if includeVersionPrefix {
		parts = append([]string{"api", "v1"}, parts...)
	}

	u, err := url.Parse(path.Join(parts...))
	if err != nil {
		return nil, err
	}

	if qp != nil {
		u.RawQuery = qp.Encode()
	}

	return tu.ResolveReference(u), nil
}

// buildRawURL takes a given set of query parameters and url parts, and returns.
// a parsed url object from them.
func (b *Builder) buildRawURL(ctx context.Context, queryParams url.Values, parts ...string) *url.URL {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	tu := *b.url

	parts = append([]string{"api", "v1"}, parts...)

	u, err := url.Parse(path.Join(parts...))
	if err != nil {
		b.logger.Error(err, "building url")
		return nil
	}

	if queryParams != nil {
		u.RawQuery = queryParams.Encode()
	}

	out := tu.ResolveReference(u)

	tracing.AttachURLToSpan(span, out)

	return out
}

// buildVersionlessURL builds a url without the `/api/v1/` prefix. It should otherwise be identical to buildRawURL.
func (b *Builder) buildVersionlessURL(ctx context.Context, qp url.Values, parts ...string) string {
	_, span := b.tracer.StartSpan(ctx)
	defer span.End()

	u, err := buildRawURL(b.url, qp, false, parts...)
	if err != nil {
		b.logger.Error(err, "building versionless url")
		return ""
	}

	return u.String()
}

// BuildWebsocketURL builds a standard url and then converts its scheme to the websocket protocol.
func (b *Builder) BuildWebsocketURL(ctx context.Context, parts ...string) string {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	u := b.buildRawURL(ctx, nil, parts...)
	u.Scheme = "ws"

	return u.String()
}

// BuildHealthCheckRequest builds a health check HTTP request.
func (b *Builder) BuildHealthCheckRequest(ctx context.Context) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	u := *b.url
	uri := fmt.Sprintf("%s://%s/_meta_/ready", u.Scheme, u.Host)

	return http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
}

// buildDataRequest builds an HTTP request for a given method, url, and body data.
func (b *Builder) buildDataRequest(ctx context.Context, method, uri string, in interface{}) (*http.Request, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	body, err := createBodyFromStruct(in)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, uri, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-type", "application/json")

	return req, nil
}
