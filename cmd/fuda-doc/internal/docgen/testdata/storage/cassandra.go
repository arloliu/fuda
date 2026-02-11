// Package storage provides configuration types for various storage backends.
package storage

import "time"

// CassandraConfig holds Cassandra cluster connection settings.
// It supports multi-datacenter deployments and tunable consistency.
type CassandraConfig struct {
	// Hosts is a comma-separated list of seed nodes.
	// At least one seed node must be reachable for the driver to bootstrap.
	//
	// Example:
	//   hosts: "cass-1.example.com,cass-2.example.com"
	Hosts string `yaml:"hosts" default:"127.0.0.1" env:"CASSANDRA_HOSTS"`

	// Port is the CQL native transport port.
	Port int `yaml:"port" default:"9042" env:"CASSANDRA_PORT"`

	// Keyspace is the default keyspace for queries.
	Keyspace string `yaml:"keyspace" default:"app" env:"CASSANDRA_KEYSPACE" validate:"required"`

	// Consistency is the default consistency level.
	//
	// Supported values:
	//   - ONE: Fastest reads, lowest durability
	//   - QUORUM: Balanced performance/durability
	//   - LOCAL_QUORUM: Datacenter-aware quorum
	//   - ALL: Highest durability, slowest
	Consistency string `yaml:"consistency" default:"QUORUM" validate:"oneof=ONE QUORUM LOCAL_QUORUM ALL"`

	// ConnectTimeout is the initial connection timeout per host.
	ConnectTimeout time.Duration `yaml:"connect_timeout" default:"5s"`

	// Auth holds optional authentication credentials for the cluster.
	Auth *CassandraAuth `yaml:"auth,omitempty"`

	// Retry configures the retry policy for failed queries.
	Retry RetryPolicy `yaml:"retry"`
}

// CassandraAuth stores credentials for Cassandra authentication.
type CassandraAuth struct {
	// Username for plain-text authentication.
	Username string `yaml:"username" env:"CASSANDRA_USER"`

	// Password for plain-text authentication.
	Password string `yaml:"password" env:"CASSANDRA_PASSWORD"`
}

// RetryPolicy configures automatic retry behaviour for failed operations.
type RetryPolicy struct {
	// MaxRetries is the maximum number of retry attempts.
	MaxRetries int `yaml:"max_retries" default:"3"`

	// InitialBackoff is the delay before the first retry.
	InitialBackoff time.Duration `yaml:"initial_backoff" default:"100ms"`

	// MaxBackoff caps the exponential backoff duration.
	MaxBackoff time.Duration `yaml:"max_backoff" default:"5s"`
}

// S3Config configures an S3-compatible object store.
type S3Config struct {
	// Bucket is the storage bucket name.
	Bucket string `yaml:"bucket" env:"S3_BUCKET" validate:"required"`

	// Region is the AWS region or compatible region identifier.
	Region string `yaml:"region" default:"us-east-1" env:"S3_REGION"`

	// Endpoint is a custom S3-compatible endpoint URL.
	// Leave empty to use the default AWS endpoint.
	//
	// Examples:
	//   http://localhost:9000          (MinIO local)
	//   https://s3.us-west-2.amazonaws.com  (explicit AWS)
	Endpoint string `yaml:"endpoint,omitempty" env:"S3_ENDPOINT"`

	// Credentials holds the access key pair for authentication.
	Credentials S3Credentials `yaml:"credentials"`
}

// S3Credentials stores AWS-style access credentials.
type S3Credentials struct {
	// AccessKeyID is the public part of the access key pair.
	AccessKeyID string `yaml:"access_key_id" env:"AWS_ACCESS_KEY_ID" validate:"required"`

	// SecretAccessKey is the secret part of the access key pair.
	// Never log or expose this value.
	SecretAccessKey string `yaml:"secret_access_key" env:"AWS_SECRET_ACCESS_KEY" validate:"required"`
}
