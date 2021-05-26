package requests

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	exampleURI       = "https://todo.verygoodsoftwarenotvirus.ru"
	asciiControlChar = string(byte(127))
)

var (
	parsedExampleURL *url.URL
	invalidParsedURL *url.URL
)

func init() {
	var err error

	parsedExampleURL, err = url.Parse(exampleURI)
	if err != nil {
		panic(err)
	}

	u := mustParseURL("https://verygoodsoftwarenotvirus.ru")
	u.Scheme = fmt.Sprintf(`%s://`, asciiControlChar)
	invalidParsedURL = u
}

// mustParseURL parses a url or otherwise panics.
func mustParseURL(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}

	return u
}

type (
	valuer map[string][]string
)

func (v valuer) ToValues() url.Values {
	return url.Values(v)
}

type requestSpec struct {
	path              string
	method            string
	query             string
	pathArgs          []interface{}
	bodyShouldBeEmpty bool
}

func newRequestSpec(bodyShouldBeEmpty bool, method, query, path string, pathArgs ...interface{}) *requestSpec {
	return &requestSpec{
		path:              path,
		pathArgs:          pathArgs,
		method:            method,
		query:             query,
		bodyShouldBeEmpty: bodyShouldBeEmpty,
	}
}

func assertRequestQuality(t *testing.T, req *http.Request, spec *requestSpec) {
	t.Helper()

	expectedPath := fmt.Sprintf(spec.path, spec.pathArgs...)

	require.NotNil(t, req, "provided req must not be nil")
	require.NotNil(t, spec, "provided spec must not be nil")

	bodyBytes, err := httputil.DumpRequest(req, true)
	require.NotEmpty(t, bodyBytes)
	require.NoError(t, err)

	if spec.bodyShouldBeEmpty {
		bodyLines := strings.Split(string(bodyBytes), "\n")
		require.NotEmpty(t, bodyLines)
		assert.Empty(t, bodyLines[len(bodyLines)-1], "body was expected to be empty, and was not empty")
	}

	assert.Equal(t, spec.query, req.URL.Query().Encode(), "expected query to be %q, but was %q instead", spec.query, req.URL.Query().Encode())
	assert.Equal(t, expectedPath, req.URL.Path, "expected path to be %q, but was %q instead", expectedPath, req.URL.Path)
	assert.Equal(t, spec.method, req.Method, "expected method to be %q, but was %q instead", spec.method, req.Method)
}

func buildTestRequestBuilder() *Builder {
	l := logging.NewNoopLogger()

	return &Builder{
		url:     parsedExampleURL,
		logger:  l,
		tracer:  tracing.NewTracer("test"),
		encoder: encoding.ProvideClientEncoder(l, encoding.ContentTypeJSON),
	}
}

type testHelper struct {
	ctx         context.Context
	builder     *Builder
	exampleUser *types.User
}

func buildTestHelper() *testHelper {
	helper := &testHelper{
		ctx:         context.Background(),
		builder:     buildTestRequestBuilder(),
		exampleUser: fakes.BuildFakeUser(),
	}

	// the hashed passwords is never transmitted over the wire.
	helper.exampleUser.HashedPassword = ""
	// the two factor secret is transmitted over the wire only on creation.
	helper.exampleUser.TwoFactorSecret = ""
	// the two factor secret validation is never transmitted over the wire.
	helper.exampleUser.TwoFactorSecretVerifiedOn = nil

	return helper
}

func buildTestRequestBuilderWithInvalidURL() *Builder {
	l := logging.NewNoopLogger()

	return &Builder{
		url:     invalidParsedURL,
		logger:  l,
		tracer:  tracing.NewTracer("test"),
		encoder: encoding.ProvideClientEncoder(l, encoding.ContentTypeJSON),
	}
}
