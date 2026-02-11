// Package messaging provides configuration types for message brokers.
package messaging

import "time"

// KafkaConfig configures a Kafka producer/consumer client.
// Supports SASL authentication and TLS.
type KafkaConfig struct {
	// Brokers is a comma-separated list of broker addresses.
	//
	// Format: host1:port1,host2:port2
	// Example: "kafka-1:9092,kafka-2:9092,kafka-3:9092"
	Brokers string `yaml:"brokers" default:"localhost:9092" env:"KAFKA_BROKERS"`

	// ClientID identifies this client to the Kafka cluster.
	// Used for logging and quota enforcement on the broker side.
	ClientID string `yaml:"client_id" default:"fuda-app" env:"KAFKA_CLIENT_ID"`

	// Consumer holds consumer-group specific settings.
	Consumer ConsumerConfig `yaml:"consumer"`

	// Producer holds producer-specific settings.
	Producer ProducerConfig `yaml:"producer"`

	// SASL configures Simple Authentication and Security Layer.
	// Set to nil to disable authentication.
	SASL *SASLConfig `yaml:"sasl,omitempty"`
}

// ConsumerConfig tunes the Kafka consumer behaviour.
type ConsumerConfig struct {
	// GroupID is the consumer group identifier.
	// All consumers with the same GroupID share partition assignments.
	GroupID string `yaml:"group_id" default:"fuda-group" env:"KAFKA_GROUP_ID" validate:"required"`

	// Topics is the list of topics to subscribe to.
	Topics []string `yaml:"topics" default:"events,commands"`

	// AutoOffsetReset controls where to start consuming when no committed offset exists.
	//
	// Supported values:
	//   - earliest: Start from the beginning of the topic
	//   - latest:   Start from the latest message
	AutoOffsetReset string `yaml:"auto_offset_reset" default:"latest" validate:"oneof=earliest latest"`

	// MaxPollInterval is the maximum delay between poll calls before the
	// consumer is considered failed and its partitions are reassigned.
	MaxPollInterval time.Duration `yaml:"max_poll_interval" default:"5m"`
}

// ProducerConfig tunes the Kafka producer behaviour.
type ProducerConfig struct {
	// Acks controls the durability guarantee for produced messages.
	//
	// Values:
	//   0  - Fire and forget (no ack)
	//   1  - Leader ack only
	//   -1 - All in-sync replicas must ack
	Acks int `yaml:"acks" default:"-1" validate:"oneof=0 1 -1"`

	// BatchSize is the maximum number of bytes to batch before sending.
	BatchSize int `yaml:"batch_size" default:"16384"`

	// LingerMs is the time to wait for additional messages before sending a batch.
	LingerMs int `yaml:"linger_ms" default:"5"`

	// Compression sets the compression codec for produced messages.
	//
	// Supported codecs:
	//   none, gzip, snappy, lz4, zstd
	Compression string `yaml:"compression" default:"snappy" validate:"oneof=none gzip snappy lz4 zstd"`
}

// SASLConfig configures SASL authentication for Kafka.
type SASLConfig struct {
	// Mechanism is the SASL mechanism to use.
	//
	// Supported:
	//   PLAIN, SCRAM-SHA-256, SCRAM-SHA-512
	Mechanism string `yaml:"mechanism" default:"PLAIN" validate:"oneof=PLAIN SCRAM-SHA-256 SCRAM-SHA-512"`

	// Username for SASL authentication.
	Username string `yaml:"username" env:"KAFKA_SASL_USER" validate:"required"`

	// Password for SASL authentication.
	Password string `yaml:"password" env:"KAFKA_SASL_PASSWORD" validate:"required"`
}
