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
	serverAdminUsername = "admin"
	serverAdminPassword = "enc:cGxlYXNlIHRlbGwgbWUgeW91IGNoYW5nZWQgdGhpcyE="
)

func buildServer(ctx context.Context, l logging.Logger, cfg AuditLogConfig) (*immuserver.ImmuServer, error) {
	immuauth.SysAdminUsername = serverAdminUsername
	immuauth.SysAdminPassword = serverAdminPassword

	serverOptions := immuserver.DefaultOptions()
	serverOptions = serverOptions.
		WithAdminPassword(serverAdminPassword).
		WithDevMode(cfg.DevMode).
		WithDir(cfg.DirPath).
		WithPort(cfg.ServerPort).
		WithSigningKey(cfg.SigningKeyFilepath).
		WithConfig("").
		WithNetwork("tcp").
		WithAuth(true).
		WithCorruptionCheck(true).
		WithAddress("0.0.0.0")

	if cfg.MutualTLSPkey != "" && cfg.MutualTLSCertificate != "" && cfg.MutualTLSClientCAs != "" {
		serverOptions = serverOptions.
			WithMTLs(true).
			WithMTLsOptions(immuserver.MTLsOptions{
				Pkey:        cfg.MutualTLSPkey,
				Certificate: cfg.MutualTLSCertificate,
				ClientCAs:   cfg.MutualTLSClientCAs,
			})
	}

	server := immuserver.DefaultServer()

	if cfg.SigningKeyFilepath == "" && cfg.PEMEncodedSigningKey != "" {
		block, rest := pem.Decode([]byte(cfg.PEMEncodedSigningKey))
		if len(rest) > 0 {
			return nil, fmt.Errorf("some issue parsing signing key")
		}

		privateKey, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("some issue parsing private key: %w", err)
		}

		sn := signer.NewSignerFromPKey(rand.Reader, privateKey)
		server.RootSigner = immuserver.NewRootSigner(sn)
	} else if cfg.SigningKeyFilepath != "" && cfg.PEMEncodedSigningKey == "" {
		sn, err := signer.NewSigner(cfg.SigningKeyFilepath)
		if err != nil {
			return nil, fmt.Errorf("some issue reading signing key file: %w", err)
		}

		server.RootSigner = immuserver.NewRootSigner(sn)
	}
	server.Logger = wrapLogger(l)
	//server.Options = serverOptions

	go func() {
		if err := server.Start(); err != nil {
			l.Fatal(err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	loginReq := &immuschema.LoginRequest{
		User:     []byte(serverAdminUsername),
		Password: []byte(serverAdminPassword),
	}
	loginResponse, err := server.Login(ctx, loginReq)
	if err != nil {
		return nil, fmt.Errorf("error logging in as admin: %w", err)
	}

	ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("authorization", loginResponse.Token))

	// ensure database

	databases, err := server.DatabaseList(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error querying users: %w", err)
	}

	var dbAlreadyCreated bool
	for _, db := range databases.Databases {
		if string(db.Databasename) == cfg.DatabaseName {
			dbAlreadyCreated = true
			break
		}
	}

	if !dbAlreadyCreated {
		if _, err := server.CreateDatabase(ctx, &immuschema.Database{Databasename: cfg.DatabaseName}); err != nil {
			return nil, fmt.Errorf("error creating database: %w", err)
		}
	}

	if _, err := server.UseDatabase(ctx, &immuschema.Database{Databasename: cfg.DatabaseName}); err != nil {
		return nil, fmt.Errorf("error using database: %w", err)
	}

	// ensure users

	users, err := server.ListUsers(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error querying users: %w", err)
	}

	var userAlreadyCreated bool
	for _, u := range users.Users {
		if string(u.User) == cfg.Username {
			userAlreadyCreated = true
			break
		}
	}

	if !userAlreadyCreated {
		if _, err := server.CreateUser(ctx, &immuschema.CreateUserRequest{
			User:       []byte(cfg.Username),
			Password:   []byte(cfg.Password),
			Permission: immuauth.PermissionRW,
			Database:   cfg.DatabaseName,
		}); err != nil {
			return nil, fmt.Errorf("error creating user: %w", err)
		}
	}

	return server, nil
}
