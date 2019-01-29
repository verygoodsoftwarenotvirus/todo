package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/server/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/services/v1/users"

	opentracing "github.com/opentracing/opentracing-go"
	config "github.com/uber/jaeger-client-go/config"
	jexpvar "github.com/uber/jaeger-lib/metrics/expvar"
)

const (
	secure        = false
	dbFile        = "example.db"
	schemaDir     = "database/v1/sqlite/schema"
	certFile      = "certs/cert.pem"
	keyFile       = "certs/key.pem"
	localCertFile = "dev_files/certs/server/cert.pem"
	localKeyFile  = "dev_files/certs/server/key.pem"
	cookieSecret  = "HEREISA32CHARSECRETWHICHISMADEUP"
)

var (
	certToUse, keyToUse string
	debug               bool
)

func init() {
	debug = strings.ToLower(os.Getenv("DOCKER")) == "true"
	if debug {
		log.Println("running in a docker environment")
		certToUse, keyToUse = certFile, keyFile
	} else {
		certToUse, keyToUse = localCertFile, localKeyFile
	}
	log.Printf("debug: %v\n", debug)
	log.Printf("using this cert: %q\n", certToUse)
	log.Printf("using this key: %q\n", keyToUse)
}

// provideJaeger returns an instance of Jaeger Tracer that samples 100% of traces and logs all spans to stdout.
func provideJaeger() (tracer opentracing.Tracer) {
	cfg, err := config.FromEnv()
	if err != nil {
		log.Fatal("cannot parse Jaeger env vars", err)
	}
	cfg.ServiceName = "todo-server"
	cfg.Sampler.Type = "const"
	cfg.Sampler.Param = 1

	// TODO(ys) a quick hack to ensure random generators get different seeds, which are based on current time.
	metricsFactory := jexpvar.NewFactory(10).Namespace(cfg.ServiceName, nil)
	if tracer, _, err = cfg.NewTracer(
		// config.Logger(jaegerLogger),
		config.Metrics(metricsFactory),
		// config.Observer(rpcmetrics.NewObserver(metricsFactory, rpcmetrics.DefaultNameNormalizer)),
	); err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}

	return tracer
}

func main() {
	server, err := BuildServer(
		database.ConnectionDetails(dbFile),
		schemaDir,
		server.CertPair{
			CertFile: certToUse,
			KeyFile:  keyToUse,
		},
		users.CookieName("todo"),
		[]byte(cookieSecret),
		true,
	)

	if err != nil {
		log.Fatal(err)
	} else {
		server.Serve()
	}
}
