package httpserver

import (
	"crypto/tls"
	"net/http"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/metrics"

	"github.com/google/wire"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/newsman"
)

var (
	// Providers is our wire superset of providers this package offers
	Providers = wire.NewSet(
		paramFetcherProviders,
		ProvideServer,
		ProvideNamespace,
		ProvideNewsmanTypeNameManipulationFunc,
	)
)

// ProvideNamespace provides a namespace
func ProvideNamespace() metrics.Namespace {
	return "todo-service"
}

// ProvideNewsmanTypeNameManipulationFunc provides an WebhookIDFetcher
func ProvideNewsmanTypeNameManipulationFunc(logger logging.Logger) newsman.TypeNameManipulationFunc {
	return func(s string) string {
		logger.WithName("events").WithValue("type_name", s).Info("event occurred")
		return s
	}
}

// provideHTTPServer provides an HTTP httpServer
func provideHTTPServer() *http.Server {
	// heavily inspired by https://blog.cloudflare.com/exposing-go-on-the-internet/
	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		TLSConfig: &tls.Config{
			PreferServerCipherSuites: true,
			// "Only use curves which have assembly implementations"
			CurvePreferences: []tls.CurveID{
				tls.CurveP256,
				tls.X25519,
			},
			MinVersion: tls.VersionTLS12,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
		},
	}
	return srv
}
