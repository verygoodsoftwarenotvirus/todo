package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/sqlite"
	"gitlab.com/verygoodsoftwarenotvirus/todo/server/v1"

	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	config "github.com/uber/jaeger-client-go/config"
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

// initJaeger returns an instance of Jaeger Tracer that samples 100% of traces and logs all spans to stdout.
func initJaeger(service string) (opentracing.Tracer, io.Closer) {
	cfg := &config.Configuration{
		Sampler: &config.SamplerConfig{Type: "const", Param: 1},
		Reporter: &config.ReporterConfig{
			LogSpans: true,
		},
	}
	tracer, closer, err := cfg.New(service, config.Logger(jaeger.StdLogger))
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}

	return tracer, closer
}

func main() {
	dbCfg := database.Config{
		Debug:            true,
		ConnectionString: dbFile,
		SchemaDir:        schemaDir,
	}

	tracer, closer := initJaeger("todo-server")
	defer closer.Close()

	cfg := server.Config{
		DebugMode:    true,
		CookieSecret: []byte(cookieSecret),
		CertFile:     certToUse,
		KeyFile:      keyToUse,
		Tracer:       tracer,
		DBBuilder:    sqlite.NewSqlite,
	}

	if server, err := server.NewDebug(cfg, dbCfg); err != nil {
		panic(err)
	} else {
		server.Serve()
	}
}
