package encoding

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"io"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
)

type (
	// ClientEncoder is an encoder for a service client.
	ClientEncoder interface {
		ContentType() string
		Unmarshal(ctx context.Context, data []byte, v interface{}) error
		Encode(ctx context.Context, dest io.Writer, v interface{}) error
		EncodeReader(ctx context.Context, data interface{}) (io.Reader, error)
	}

	// clientEncoder is our concrete implementation of ClientEncoder.
	clientEncoder struct {
		logger      logging.Logger
		tracer      tracing.Tracer
		contentType *contentType
	}
)

func (e *clientEncoder) Unmarshal(ctx context.Context, data []byte, v interface{}) error {
	_, span := e.tracer.StartSpan(ctx)
	defer span.End()

	switch e.contentType {
	case ContentTypeXML:
		return xml.Unmarshal(data, v)
	default:
		return json.Unmarshal(data, v)
	}
}

func (e *clientEncoder) Encode(ctx context.Context, dest io.Writer, data interface{}) error {
	_, span := e.tracer.StartSpan(ctx)
	defer span.End()

	switch e.contentType {
	case ContentTypeXML:
		return xml.NewEncoder(dest).Encode(data)
	default:
		return json.NewEncoder(dest).Encode(data)
	}
}

func (e *clientEncoder) EncodeReader(ctx context.Context, data interface{}) (io.Reader, error) {
	_, span := e.tracer.StartSpan(ctx)
	defer span.End()

	switch e.contentType {
	case ContentTypeXML:
		out, err := xml.Marshal(data)
		if err != nil {
			return nil, observability.PrepareError(err, e.logger, span, "marshaling to XML")
		}

		return bytes.NewReader(out), nil
	default:
		out, err := json.Marshal(data)
		if err != nil {
			return nil, observability.PrepareError(err, e.logger, span, "marshaling to JSON")
		}

		return bytes.NewReader(out), nil
	}
}

// ProvideClientEncoder provides a ClientEncoder.
func ProvideClientEncoder(logger logging.Logger, encoding *contentType) ClientEncoder {
	return &clientEncoder{
		logger:      logging.EnsureLogger(logger).WithName("client_encoder"),
		tracer:      tracing.NewTracer("client_encoder"),
		contentType: encoding,
	}
}
