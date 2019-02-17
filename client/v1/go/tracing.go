package client

import (
	"crypto/tls"
	"net/http/httptrace"
	"net/textproto"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

// NewClientTrace returns a client trace from a span that logs the various events in the trace
func NewClientTrace(span opentracing.Span) *httptrace.ClientTrace {
	return &httptrace.ClientTrace{
		GetConn:              buildGetConn(span),              // func(hostPort string)
		GotConn:              buildGotConn(span),              // func(GotConnInfo)
		PutIdleConn:          buildPutIdleConn(span),          // func(err error)
		GotFirstResponseByte: buildGotFirstResponseByte(span), // func()
		Got100Continue:       buildGot100Continue(span),       // func()
		Got1xxResponse:       buildGot1xxResponse(span),       // func(code int, header textproto.MIMEHeader) error
		DNSStart:             buildDNSStart(span),             // func(DNSStartInfo)
		DNSDone:              buildDNSDone(span),              // func(DNSDoneInfo)
		ConnectStart:         buildConnectStart(span),         // func(network, addr string)
		ConnectDone:          buildConnectDone(span),          // func(network, addr string, err error)
		TLSHandshakeStart:    buildTLSHandshakeStart(span),    // func()
		TLSHandshakeDone:     buildTLSHandshakeDone(span),     // func(tls.ConnectionState, error)
		WroteHeaderField:     buildWroteHeaderField(span),     // func(key string, value []string)
		WroteHeaders:         buildWroteHeaders(span),         // func()
		Wait100Continue:      buildWait100Continue(span),      // func()
		WroteRequest:         buildWroteRequest(span),         // func(WroteRequestInfo)
	}
}

func buildGetConn(span opentracing.Span) func(string) {
	return func(hostPort string) {
		span.LogFields(
			log.String("event", "GetConn"),
		)
	}
}

func buildGotConn(span opentracing.Span) func(httptrace.GotConnInfo) {
	return func(connInfo httptrace.GotConnInfo) {
		span.LogFields(
			log.String("event", "GotConn"),
		)
	}
}

func buildPutIdleConn(span opentracing.Span) func(err error) {
	return func(err error) {
		span.LogFields(
			log.String("event", "PutIdleConn"),
		)
	}
}

func buildGotFirstResponseByte(span opentracing.Span) func() {
	return func() {
		span.LogFields(
			log.String("event", "GotFirstResponseByte"),
		)
	}
}

func buildGot100Continue(span opentracing.Span) func() {
	return func() {
		span.LogFields(
			log.String("event", "Got100Continue"),
		)
	}
}

func buildGot1xxResponse(span opentracing.Span) func(int, textproto.MIMEHeader) error {
	return func(code int, header textproto.MIMEHeader) error {
		span.LogFields(
			log.String("event", "Got1xxResponse"),
		)
		return nil
	}
}

func buildDNSStart(span opentracing.Span) func(httptrace.DNSStartInfo) {
	return func(info httptrace.DNSStartInfo) {
		span.LogFields(
			log.String("event", "DNSStart"),
		)
	}
}

func buildDNSDone(span opentracing.Span) func(httptrace.DNSDoneInfo) {
	return func(info httptrace.DNSDoneInfo) {
		span.LogFields(
			log.String("event", "DNSDone"),
		)
	}
}

func buildConnectStart(span opentracing.Span) func(network, addr string) {
	return func(network, addr string) {
		span.LogFields(
			log.String("event", "ConnectStart"),
		)
	}
}

func buildConnectDone(span opentracing.Span) func(network, addr string, err error) {
	return func(network, addr string, err error) {
		span.LogFields(
			log.String("event", "ConnectDone"),
		)
	}
}

func buildTLSHandshakeStart(span opentracing.Span) func() {
	return func() {
		span.LogFields(
			log.String("event", "TLSHandshakeStart"),
		)
	}
}

func buildTLSHandshakeDone(span opentracing.Span) func(tls.ConnectionState, error) {
	return func(cs tls.ConnectionState, err error) {
		span.LogFields(
			log.String("event", "TLSHandshakeDone"),
		)
	}
}

func buildWroteHeaderField(span opentracing.Span) func(string, []string) {
	return func(key string, value []string) {
		span.LogFields(
			log.String("event", "WroteHeaderField"),
		)
	}
}

func buildWroteHeaders(span opentracing.Span) func() {
	return func() {
		span.LogFields(
			log.String("event", "WroteHeaders"),
		)
	}
}

func buildWait100Continue(span opentracing.Span) func() {
	return func() {
		span.LogFields(
			log.String("event", "Wait100Continue"),
		)
	}
}

func buildWroteRequest(span opentracing.Span) func(httptrace.WroteRequestInfo) {
	return func(wri httptrace.WroteRequestInfo) {
		span.LogFields(
			log.String("event", "WroteRequest"),
		)
	}
}
