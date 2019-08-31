package client

import (
	"net"
	"net/http"
	"time"
)

const (
	userAgentHeader = "User-Agent"
	userAgent       = "TODO Service Client"
)

type defaultRoundTripper struct {
	baseTransport *http.Transport
}

func newDefaultRoundTripper() *defaultRoundTripper {
	return &defaultRoundTripper{
		baseTransport: buildDefaultTransport(),
	}
}

func (t *defaultRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set(userAgentHeader, userAgent)
	return t.baseTransport.RoundTrip(req)
}

func buildDefaultTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}
