package main

import (
	immuclient "github.com/codenotary/immudb/pkg/client"
	"google.golang.org/grpc"
)

func buildClient(cfg AuditLogClientConfig) (immuclient.ImmuClient, error) {
	clientOptions := &immuclient.Options{
		Dir:                cfg.DirPath,
		Address:            cfg.Hostname,
		Port:               cfg.Port,
		HealthCheckRetries: 5,
		MTLs:               false,
		Auth:               true,
		DialOptions:        &[]grpc.DialOption{},
		Tkns:               immuclient.NewTokenService().WithHds(immuclient.NewHomedirService()),
		Metrics:            true,
	}

	if cfg.MutualTLSCertificate != "" && cfg.MutualTLSPkey != "" && cfg.MutualTLSClientCAs != "" {
		clientOptions.MTLs = true
		clientOptions.MTLsOptions = immuclient.MTLsOptions{
			Servername:  cfg.Hostname,
			Pkey:        cfg.MutualTLSPkey,
			Certificate: cfg.MutualTLSCertificate,
			ClientCAs:   cfg.MutualTLSClientCAs,
		}
	}

	client, err := immuclient.NewImmuClient(clientOptions)
	if err != nil {
		return nil, err
	}

	return client, nil
}
