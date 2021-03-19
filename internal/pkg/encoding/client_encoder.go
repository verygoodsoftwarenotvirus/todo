package encoding

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"io"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
)

func buildContentType(s string) *contentType {
	ct := contentType(&s)

	return &ct
}

var (
	// ContentTypeJSON is what we use to indicate we want JSON for some reason.
	ContentTypeJSON ContentType = buildContentType(contentTypeJSON)
	// ContentTypeXML is what we use to indicate we want XML for some reason.
	ContentTypeXML ContentType = buildContentType(contentTypeXML)
)

type (
	// ContentType is the publicly accessible version of contentType.
	ContentType *contentType

	contentType *string

	// ClientEncoder is an encoder for a service client.
	ClientEncoder interface {
		ContentType() string
		EncodeReader(ctx context.Context, data interface{}) (io.Reader, error)
	}

	// clientEncoder is our concrete implementation of ClientEncoder.
	clientEncoder struct {
		logger   logging.Logger
		tracer   tracing.Tracer
		encoding *contentType
	}
)

func (e *clientEncoder) ContentType() string {
	switch e.encoding {
	case ContentTypeJSON:
		return contentTypeJSON
	case ContentTypeXML:
		return contentTypeXML
	default:
		return ""
	}
}

func (e *clientEncoder) EncodeReader(ctx context.Context, data interface{}) (io.Reader, error) {
	_, span := e.tracer.StartSpan(ctx)
	defer span.End()

	switch e.encoding {
	case ContentTypeXML:
		out, err := xml.Marshal(data)
		if err != nil {
			tracing.AttachErrorToSpan(span, err)
			return nil, err
		}

		return bytes.NewReader(out), nil
	default:
		out, err := json.Marshal(data)
		if err != nil {
			tracing.AttachErrorToSpan(span, err)
			return nil, err
		}

		return bytes.NewReader(out), nil
	}
}

// ProvideClientEncoder provides a ClientEncoder.
func ProvideClientEncoder(logger logging.Logger, encoding *contentType) ClientEncoder {
	return &clientEncoder{
		logger:   logging.EnsureLogger(logger).WithName("client_encoder"),
		tracer:   tracing.NewTracer("client_encoder"),
		encoding: encoding,
	}
}
