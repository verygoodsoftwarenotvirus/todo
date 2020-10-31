package main

import (
	"context"
	"fmt"

	immuclient "github.com/codenotary/immudb/pkg/client"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
)

type ImmutableAuditor struct {
	logger logging.Logger
	token  string
	client immuclient.ImmuClient
}

func ProvideImmutableAuditor(ctx context.Context, logger logging.Logger, username, password string) (*ImmutableAuditor, error) {
	options := immuclient.DefaultOptions()

	client, err := immuclient.NewImmuClient(options)
	if err != nil {
		return nil, fmt.Errorf("error initializing immudb client: %w", err)
	}

	// login with default username and password
	lr, err := client.Login(ctx, []byte(username), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("error logging into immudb server: %w", err)
	}

	a := &ImmutableAuditor{
		token:  lr.Token,
		client: client,
		logger: logger.WithName("immutable_audit_log"),
	}

	// don't read this quickly
	return a, nil
}
