package events

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/nats-io/nats.go"
	"github.com/streadway/amqp"
	"gocloud.dev/gcp"
	"gocloud.dev/pubsub"
	"gocloud.dev/pubsub/awssnssqs"
	"gocloud.dev/pubsub/azuresb"
	"gocloud.dev/pubsub/gcppubsub"
	"gocloud.dev/pubsub/kafkapubsub"
	"gocloud.dev/pubsub/mempubsub"
	_ "gocloud.dev/pubsub/mempubsub"
	"gocloud.dev/pubsub/natspubsub"
	"gocloud.dev/pubsub/rabbitpubsub"
	_ "gocloud.dev/pubsub/rabbitpubsub"
	"golang.org/x/oauth2/google"
)

const (
	// ProviderGoogleCloudPubSub is a pub/sub provider string.
	ProviderGoogleCloudPubSub = "google_cloud_pubsub"
	// ProviderAWSSNS is a pub/sub provider string.
	ProviderAWSSNS = "aws_sns"
	// ProviderRabbitMQ is a pub/sub provider string.
	ProviderRabbitMQ = "rabbit_mq"
	// ProviderAzureServiceBus is a pub/sub provider string.
	ProviderAzureServiceBus = "azure_service_bus"
	// ProviderKafka is a pub/sub provider string.
	ProviderKafka = "kafka"
	// ProviderNATS is a pub/sub provider string.
	ProviderNATS = "nats"
	// ProviderMemory is a pub/sub provider string.
	ProviderMemory = "memory"
)

type (
	// Config configures an EventPublisher
	Config struct {
		Enabled         bool      `json:"scopes" mapstructure:"scopes" toml:"scopes,omitempty"`
		Name            string    `json:"name" mapstructure:"name" toml:"name,omitempty"`
		Provider        string    `json:"provider" mapstructure:"provider" toml:"provider,omitempty"`
		TopicIdentifier string    `json:"topic_identifier" mapstructure:"topic_identifier" toml:"topic_identifier,omitempty"`
		ConnectionURL   string    `json:"connection_url" mapstructure:"connection_url" toml:"connection_url,omitempty"`
		GCPPubSub       GCPPubSub `json:"gcp" mapstructure:"gcp" toml:"gcp,omitempty"`
	}

	GCPPubSub struct {
		Scopes                    []string `json:"scopes" mapstructure:"scopes" toml:"scopes,omitempty"`
		ServiceAccountKeyFilepath string
	}
)

var _ validation.ValidatableWithContext = (*Config)(nil)

func (c *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, c,
		validation.Field(c.Name, validation.Required),
		validation.Field(c.Provider, validation.In(
			ProviderGoogleCloudPubSub,
			ProviderAWSSNS,
			ProviderRabbitMQ,
			ProviderAzureServiceBus,
			ProviderKafka,
			ProviderNATS,
			ProviderMemory,
		)),
	)
}

var errInvalidProvider = errors.New("invalid events topic provider")

func ProvideTopic(ctx context.Context, cfg *Config) (*pubsub.Topic, error) {
	switch cfg.Provider {
	case ProviderGoogleCloudPubSub:
		var creds *google.Credentials

		if cfg.GCPPubSub.ServiceAccountKeyFilepath != "" {
			serviceAccountKeyBytes, err := os.ReadFile(cfg.GCPPubSub.ServiceAccountKeyFilepath)
			if err != nil {
				return nil, fmt.Errorf("reading service account key file: %w", err)
			}

			if creds, err = google.CredentialsFromJSON(ctx, serviceAccountKeyBytes, cfg.GCPPubSub.Scopes...); err != nil {
				return nil, fmt.Errorf("using service account key credentials: %w", err)
			}
		} else {
			var err error
			if creds, err = gcp.DefaultCredentials(ctx); err != nil {
				return nil, fmt.Errorf("constructing pub/sub credentials: %w", err)
			}
		}

		conn, _, err := gcppubsub.Dial(ctx, creds.TokenSource)
		if err != nil {
			return nil, fmt.Errorf("dialing connection to pub/sub %w", err)
		}

		pubClient, err := gcppubsub.PublisherClient(ctx, conn)
		if err != nil {
			return nil, fmt.Errorf("establishing publisher client: %w", err)
		}

		return gcppubsub.OpenTopicByPath(pubClient, cfg.TopicIdentifier, nil)
	case ProviderAWSSNS:
		sess, err := session.NewSession(nil)
		if err != nil {
			return nil, fmt.Errorf("establishing AWS session: %w", err)
		}

		topic := awssnssqs.OpenSNSTopic(ctx, sess, cfg.TopicIdentifier, nil)

		return topic, nil
	case ProviderKafka:
		config := kafkapubsub.MinimalConfig()

		return kafkapubsub.OpenTopic(strings.Split(cfg.ConnectionURL, ","), config, cfg.TopicIdentifier, nil)
	case ProviderRabbitMQ:
		rabbitConn, err := amqp.Dial(cfg.ConnectionURL)
		if err != nil {
			return nil, err
		}

		topic := rabbitpubsub.OpenTopic(rabbitConn, cfg.TopicIdentifier, nil)

		return topic, nil
	case ProviderNATS:
		natsConn, err := nats.Connect(cfg.ConnectionURL)
		if err != nil {
			return nil, err
		}

		return natspubsub.OpenTopic(natsConn, cfg.TopicIdentifier, nil)
	case ProviderAzureServiceBus:
		busNamespace, err := azuresb.NewNamespaceFromConnectionString(cfg.ConnectionURL)
		if err != nil {
			return nil, err
		}

		busTopic, err := azuresb.NewTopic(busNamespace, cfg.TopicIdentifier, nil)
		if err != nil {
			return nil, err
		}

		return azuresb.OpenTopic(ctx, busTopic, nil)
	case ProviderMemory:
		topic := mempubsub.NewTopic()
		return topic, nil
	default:
		return nil, errInvalidProvider
	}
}
