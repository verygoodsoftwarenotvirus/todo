package main

import (
	"context"
	"fmt"
	immuschema "github.com/codenotary/immudb/pkg/api/schema"
	"google.golang.org/grpc"
	"log"
	"os"

	immuclient "github.com/codenotary/immudb/pkg/client"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/zerolog"
	"google.golang.org/grpc/metadata"
)

const (
	databaseName = "auditlog"
)

type AuditLogConfig struct {
	DevMode          bool
	DirPath          string
	DatabaseUsername string
	DatabasePassword string
	DatabasePort     int
	DatabaseHost     string
	DatabaseName     string

	PEMEncodedSigningKey string
	SigningKeyFilepath   string

	MutualTLSPkey        string
	MutualTLSCertificate string
	MutualTLSClientCAs   string
}

func main() {
	ctx := context.Background()
	logger := zerolog.NewLogger()

	const (
		username = "username"
		password = "FuckFuckFuck123!"

		dirPath = "./artifacts/audit_log"
	)

	cfg := AuditLogConfig{
		DirPath:          dirPath,
		DatabasePort:     3322,
		DatabaseName:     "auditlog",
		DatabasePassword: password,
		DatabaseUsername: username,
		DatabaseHost:     "0.0.0.0",
		PEMEncodedSigningKey: `-----BEGIN PRIVATE KEY-----
MIGkAgEBBDB7kh4WsXnskAGMZ/ATdYB0/TxCdpgj1dNhKbgK4k7rGvyaMd6xE4/L
bwbiFO5WXaagBwYFK4EEACKhZANiAAQs+yyAbBguvXURexlmc8aCeoBacYWuag3C
ORUoaMVfHJ4YYW8vdZmX0MJf11ZZJv3YAiSXD8CMLKPBGJog/4yPv2ijk8pqS/Em
+fEMvQRplDdLTdX9711puES+248mUe0=
-----END PRIVATE KEY-----
`,
	}

	logger.Info("removing directory")
	if err := os.RemoveAll(dirPath); err != nil {
		logger.Fatal(err)
	}

	_, err := buildAndStartServer(ctx, logger, cfg)
	if err != nil {
		log.Fatal("error building server: ", err)
	}

	////////////////////////////////////////////////////////////////////////////////////

	key := []byte("key")

	clientOptions := &immuclient.Options{
		Dir:                ".",
		Address:            "127.0.0.1",
		Port:               3322,
		HealthCheckRetries: 5,
		MTLs:               false,
		Auth:               true,
		Config:             "configs/immuclient.toml",
		TokenFileName:      "token",
		DialOptions:        &[]grpc.DialOption{},
		Tkns:               immuclient.NewTokenService().WithTokenFileName("token").WithHds(immuclient.NewHomedirService()),
		Metrics:            true,
		PidPath:            "",
		PrometheusHost:     "",
		PrometheusPort:     "",
		LogFileName:        "",
	}
	client, err := immuclient.NewImmuClient(clientOptions)
	if err != nil {
		log.Fatal("error building client: ", err)
	}

	// login with default username and password
	lr, err := client.Login(ctx, []byte(username), []byte(password))
	ctx = metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", lr.Token))

	udr, err := client.UseDatabase(ctx, &immuschema.Database{Databasename: cfg.DatabaseName})
	if err != nil {
		log.Fatal("error using database: ", err)
	}
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", udr.Token))

	if vi, err := client.SafeSet(ctx, key, []byte(`hello world`)); err != nil {
		log.Fatal("error setting key: ", err)
		log.Fatal(err)
	} else {
		fmt.Printf("Item inclusion verified %t\n", vi.Verified)
	}

	if item, err := client.SafeGet(ctx, key); err != nil {
		log.Fatal("error getting key: ", err)
	} else {
		fmt.Printf("Database consistency verified %t\n", item.Verified)
		fmt.Printf("%s\n", item.Value)
	}
}
