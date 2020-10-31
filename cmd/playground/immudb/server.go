package main

import (
	"context"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	immuschema "github.com/codenotary/immudb/pkg/api/schema"
	immuauth "github.com/codenotary/immudb/pkg/auth"
	immuserver "github.com/codenotary/immudb/pkg/server"
	"github.com/codenotary/immudb/pkg/signer"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
	"google.golang.org/grpc/metadata"
)

const (
	serverAdminUsername = "auditAdmin"
	serverAdminPassword = "enc:cGxlYXNlIHRlbGwgbWUgeW91IGNoYW5nZWQgdGhpcyE="
)

func init() {
	immuauth.SysAdminUsername = serverAdminUsername
	immuauth.SysAdminPassword = serverAdminPassword
}

func buildServerOptions(cfg AuditLogConfig) immuserver.Options {
	serverOptions := immuserver.DefaultOptions()
	serverOptions = serverOptions.
		WithAdminPassword(serverAdminPassword).
		WithDevMode(cfg.DevMode).
		WithDir(cfg.DirPath).
		WithPort(cfg.DatabasePort).
		WithSigningKey(cfg.SigningKeyFilepath).
		WithAddress(cfg.DatabaseHost).
		WithConfig("").
		WithNetwork("tcp").
		WithAuth(true).
		WithCorruptionCheck(true)

	if cfg.MutualTLSPkey != "" && cfg.MutualTLSCertificate != "" && cfg.MutualTLSClientCAs != "" {
		serverOptions = serverOptions.
			WithMTLs(true).
			WithMTLsOptions(immuserver.MTLsOptions{
				Pkey:        cfg.MutualTLSPkey,
				Certificate: cfg.MutualTLSCertificate,
				ClientCAs:   cfg.MutualTLSClientCAs,
			})
	}

	return serverOptions
}

func establishSigner(cfg AuditLogConfig, server *immuserver.ImmuServer) error {
	if cfg.SigningKeyFilepath == "" && cfg.PEMEncodedSigningKey != "" {
		block, rest := pem.Decode([]byte(cfg.PEMEncodedSigningKey))
		if len(rest) > 0 {
			return fmt.Errorf("some issue parsing signing key")
		}

		privateKey, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("some issue parsing private key: %w", err)
		}

		sn := signer.NewSignerFromPKey(rand.Reader, privateKey)
		server.RootSigner = immuserver.NewRootSigner(sn)
	} else if cfg.SigningKeyFilepath != "" && cfg.PEMEncodedSigningKey == "" {
		sn, err := signer.NewSigner(cfg.SigningKeyFilepath)
		if err != nil {
			return fmt.Errorf("some issue reading signing key file: %w", err)
		}

		server.RootSigner = immuserver.NewRootSigner(sn)
	}

	return nil
}

func ensureDatabases(ctx context.Context, cfg AuditLogConfig, server *immuserver.ImmuServer) error {
	databases, err := server.DatabaseList(ctx, nil)
	if err != nil {
		return fmt.Errorf("error querying users: %w", err)
	}

	var dbAlreadyCreated bool
	for _, db := range databases.Databases {
		if db.Databasename == cfg.DatabaseName {
			dbAlreadyCreated = true
			break
		}
	}

	dbSchema := &immuschema.Database{Databasename: cfg.DatabaseName}
	if !dbAlreadyCreated {
		if _, err := server.CreateDatabase(ctx, dbSchema); err != nil {
			return fmt.Errorf("error creating database: %w", err)
		}
	}

	if _, err := server.UseDatabase(ctx, dbSchema); err != nil {
		return fmt.Errorf("error using database: %w", err)
	}

	return nil
}

func ensureUsers(ctx context.Context, cfg AuditLogConfig, server *immuserver.ImmuServer) error {
	users, err := server.ListUsers(ctx, nil)
	if err != nil {
		return fmt.Errorf("error querying users: %w", err)
	}

	var userAlreadyCreated bool
	for _, u := range users.Users {
		if string(u.User) == cfg.DatabaseUsername {
			userAlreadyCreated = true
			break
		}
	}

	if !userAlreadyCreated {
		if _, err := server.CreateUser(ctx, &immuschema.CreateUserRequest{
			User:       []byte(cfg.DatabaseUsername),
			Password:   []byte(cfg.DatabasePassword),
			Permission: immuauth.PermissionRW,
			Database:   cfg.DatabaseName,
		}); err != nil {
			return fmt.Errorf("error creating user: %w", err)
		}
	}

	return nil
}

func buildAndStartServer(ctx context.Context, l logging.Logger, cfg AuditLogConfig) (*immuserver.ImmuServer, error) {
	server := immuserver.DefaultServer()
	server.Logger = wrapLogger(l)
	server.Options = buildServerOptions(cfg)

	if err := establishSigner(cfg, server); err != nil {
		return nil, fmt.Errorf("error encountered establishing signer: %w", err)
	}

	go func() {
		if err := server.Start(); err != nil {
			l.Fatal(err)
		}
	}()

	time.Sleep(300 * time.Millisecond)

	loginReq := &immuschema.LoginRequest{
		User:     []byte(serverAdminUsername),
		Password: []byte(serverAdminPassword),
	}
	loginResponse, err := server.Login(ctx, loginReq)
	if err != nil {
		return nil, fmt.Errorf("error logging in as admin: %w", err)
	}

	ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("authorization", loginResponse.Token))

	if err := ensureDatabases(ctx, cfg, server); err != nil {
		return nil, fmt.Errorf("error encountered ensuring databases: %w", err)
	}

	if err := ensureUsers(ctx, cfg, server); err != nil {
		return nil, fmt.Errorf("error encountered ensuring users: %w", err)
	}

	return server, nil
}
