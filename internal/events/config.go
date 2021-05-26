package events

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

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
	"gocloud.dev/pubsub/natspubsub"
	"gocloud.dev/pubsub/rabbitpubsub"
	"golang.org/x/oauth2/google"
)

const (
	// ProviderGoogleCloudPubSub is a pub/sub provider string.
	ProviderGoogleCloudPubSub = "google_cloud_pubsub"
	// ProviderAWSSQS is a pub/sub provider string.
	ProviderAWSSQS = "aws_sqs"
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

var (
	errNilConfig       = errors.New("nil config provided")
	errNilSubscription = errors.New("nil subscription provided")
)

type (
	// Config configures a publisher or a subscriber.
	Config struct {
		Provider               string        `json:"provider" mapstructure:"provider" toml:"provider,omitempty"`
		Topic                  string        `json:"topic" mapstructure:"topic" toml:"topic,omitempty"`
		SubscriptionIdentifier string        `json:"subscription_identifier" mapstructure:"subscription_identifier" toml:"subscription_identifier,omitempty"`
		ConnectionURL          string        `json:"connection_url" mapstructure:"connection_url" toml:"connection_url,omitempty"`
		GCPPubSub              GCPPubSub     `json:"gcp" mapstructure:"gcp" toml:"gcp,omitempty"`
		AckDeadline            time.Duration `json:"ack_deadline" mapstructure:"ack_deadline" toml:"ack_deadline,omitempty"`
		Enabled                bool          `json:"scopes" mapstructure:"scopes" toml:"scopes,omitempty"`
	}

	// GCPPubSub is a struct to segment the GCP-specific portions of configuration.
	GCPPubSub struct {
		ServiceAccountKeyFilepath string
	}
)

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates the Config struct.
func (c *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, c,
		validation.Field(&c.Provider, validation.In(
			ProviderGoogleCloudPubSub,
			ProviderAWSSQS,
			ProviderRabbitMQ,
			ProviderAzureServiceBus,
			ProviderKafka,
			ProviderNATS,
			ProviderMemory,
		)),
	)
}

var errInvalidProvider = errors.New("invalid events subscription provider")

// ProvidePublishTopic uses a configuration to provide a pubsub subscription.
func ProvidePublishTopic(ctx context.Context, cfg *Config) (*pubsub.Topic, error) {
	switch cfg.Provider {
	case ProviderGoogleCloudPubSub:
		var creds *google.Credentials

		if cfg.GCPPubSub.ServiceAccountKeyFilepath != "" {
			serviceAccountKeyBytes, err := os.ReadFile(cfg.GCPPubSub.ServiceAccountKeyFilepath)
			if err != nil {
				return nil, fmt.Errorf("reading service account key file: %w", err)
			}

			if creds, err = google.CredentialsFromJSON(ctx, serviceAccountKeyBytes); err != nil {
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

		return gcppubsub.OpenTopicByPath(pubClient, cfg.Topic, nil)
	case ProviderAWSSQS:
		sess, err := session.NewSession(nil)
		if err != nil {
			return nil, fmt.Errorf("establishing AWS session: %w", err)
		}

		topic := awssnssqs.OpenSQSTopic(ctx, sess, cfg.Topic, nil)

		return topic, nil
	case ProviderKafka:
		config := kafkapubsub.MinimalConfig()

		return kafkapubsub.OpenTopic(strings.Split(cfg.ConnectionURL, ","), config, cfg.Topic, nil)
	case ProviderRabbitMQ:
		rabbitConn, err := amqp.Dial(cfg.ConnectionURL)
		if err != nil {
			return nil, fmt.Errorf("connecting to rabbitmq: %w", err)
		}

		topic := rabbitpubsub.OpenTopic(rabbitConn, cfg.Topic, nil)

		return topic, nil
	case ProviderNATS:
		natsConn, err := nats.Connect(cfg.ConnectionURL)
		if err != nil {
			return nil, fmt.Errorf("connecting to nats: %w", err)
		}

		return natspubsub.OpenTopic(natsConn, cfg.Topic, nil)
	case ProviderAzureServiceBus:
		busNamespace, err := azuresb.NewNamespaceFromConnectionString(cfg.ConnectionURL)
		if err != nil {
			return nil, fmt.Errorf("establishing namespace for Azure Service Bus: %w", err)
		}

		busTopic, err := azuresb.NewTopic(busNamespace, cfg.Topic, nil)
		if err != nil {
			return nil, fmt.Errorf("establishing subscription for Azure Service Bus: %w", err)
		}

		return azuresb.OpenTopic(ctx, busTopic, nil)
	case ProviderMemory:
		topic := mempubsub.NewTopic()
		return topic, nil
	default:
		return nil, errInvalidProvider
	}
}

// ProvideSubscription uses a configuration to provide a pub/sub subscription.
func ProvideSubscription(ctx context.Context, cfg *Config) (*pubsub.Subscription, error) {
	switch cfg.Provider {
	case ProviderGoogleCloudPubSub:
		var creds *google.Credentials

		if cfg.GCPPubSub.ServiceAccountKeyFilepath != "" {
			serviceAccountKeyBytes, err := os.ReadFile(cfg.GCPPubSub.ServiceAccountKeyFilepath)
			if err != nil {
				return nil, fmt.Errorf("reading service account key file: %w", err)
			}

			if creds, err = google.CredentialsFromJSON(ctx, serviceAccountKeyBytes); err != nil {
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

		// Construct a SubscriberClient using the connection.
		subClient, err := gcppubsub.SubscriberClient(ctx, conn)
		if err != nil {
			return nil, fmt.Errorf("dialing connection to pub/sub %w", err)
		}

		// Construct a *pubsub.Subscription.
		return gcppubsub.OpenSubscriptionByPath(subClient, cfg.SubscriptionIdentifier, nil)
	case ProviderAWSSQS:
		sess, err := session.NewSession(nil)
		if err != nil {
			return nil, fmt.Errorf("establishing AWS session: %w", err)
		}

		subscription := awssnssqs.OpenSubscription(ctx, sess, cfg.SubscriptionIdentifier, nil)

		return subscription, nil
	case ProviderKafka:
		addrs := strings.Split(cfg.ConnectionURL, ",")
		config := kafkapubsub.MinimalConfig()

		return kafkapubsub.OpenSubscription(addrs, config, cfg.Topic, []string{cfg.SubscriptionIdentifier}, nil)
	case ProviderRabbitMQ:
		rabbitConn, err := amqp.Dial(cfg.ConnectionURL)
		if err != nil {
			return nil, fmt.Errorf("connecting to rabbitmq: %w", err)
		}

		subscription := rabbitpubsub.OpenSubscription(rabbitConn, cfg.SubscriptionIdentifier, nil)

		return subscription, nil
	case ProviderNATS:
		natsConn, err := nats.Connect(cfg.ConnectionURL)
		if err != nil {
			return nil, fmt.Errorf("connecting to NATS: %w", err)
		}

		return natspubsub.OpenSubscription(natsConn, cfg.SubscriptionIdentifier, nil)
	case ProviderAzureServiceBus:
		busNamespace, err := azuresb.NewNamespaceFromConnectionString(cfg.ConnectionURL)
		if err != nil {
			return nil, fmt.Errorf("connecting to Azure Service Bus: %w", err)
		}

		busTopic, err := azuresb.NewTopic(busNamespace, cfg.Topic, nil)
		if err != nil {
			return nil, fmt.Errorf("establishing service bus subscription: %w", err)
		}

		busSub, err := azuresb.NewSubscription(busTopic, cfg.SubscriptionIdentifier, nil)
		if err != nil {
			return nil, fmt.Errorf("establishing service bus subscription: %w", err)
		}

		return azuresb.OpenSubscription(ctx, busNamespace, busTopic, busSub, nil)
	default:
		// ProviderMemory not supported because it requires a topic to use properly
		return nil, errInvalidProvider
	}
}
