package testdata

import (
	"time"

	oa "github.com/arloliu/fuda/cmd/fuda-doc/internal/docgen/testdata/auth"
	msg "github.com/arloliu/fuda/cmd/fuda-doc/internal/docgen/testdata/messaging"
	"github.com/arloliu/fuda/cmd/fuda-doc/internal/docgen/testdata/storage"
)

// Config is the root configuration struct for the application.
// It contains all settings needed to run the service including
// server, database, and secrets configuration.
type Config struct {
	// AppName is the name of the application.
	// This is used for logging, metrics, and service discovery.
	// Should be lowercase with hyphens (e.g., "my-app").
	AppName string `yaml:"app_name" default:"my-app" env:"APP_NAME"`

	// Environment specifies the deployment environment.
	// This affects logging levels, feature toggles, and external service endpoints.
	//
	// Valid values:
	//   - dev: Development mode with debug logging
	//   - staging: Pre-production testing environment
	//   - prod: Production mode with optimized settings
	Environment string `yaml:"environment" default:"dev" env:"APP_ENV" validate:"oneof=dev prod staging"`

	// Server configures the HTTP server settings.
	// See ServerConfig for detailed options.
	Server ServerConfig `yaml:"server"`

	// Database holds configurations for all database connections.
	// Supports primary SQL database, Redis cache, and optional analytics DB.
	Database DatabaseConfig `yaml:"database"`

	// Storage configures external storage backends.
	// Cassandra is used for primary data, S3 for object storage.
	Storage StorageConfig `yaml:"storage"`

	// Messaging configures message broker connections.
	Messaging msg.KafkaConfig `yaml:"messaging"`

	// Auth configures OAuth 2.0 / OIDC authentication.
	// Set to nil to disable external authentication.
	Auth *oa.OAuthConfig `yaml:"auth,omitempty"`

	// Secrets configuration for loading sensitive values from external sources.
	// Supports file://, vault://, and environment-based secret paths.
	Secrets SecretsConfig `yaml:"secrets"`

	// Features contains feature flags for enabling/disabling functionality.
	Features map[string]bool `yaml:"features" default:"beta:false,v2:false"`

	// Metadata stores arbitrary key-value pairs for the application.
	Metadata map[string]string `yaml:"metadata" default:"owner:sre,version:v1"`

	// Tags is a list of labels for categorizing the application.
	Tags []string `yaml:"tags" default:"web,api"`
}

// StorageConfig groups external storage backend settings.
// Each backend can be independently enabled or disabled.
type StorageConfig struct {
	// Cassandra is the primary NoSQL data store.
	// Used for high-throughput, low-latency data access.
	Cassandra storage.CassandraConfig `yaml:"cassandra"`

	// ObjectStore configures S3-compatible object storage.
	// Used for file uploads, backups, and static assets.
	ObjectStore *storage.S3Config `yaml:"object_store,omitempty"`
}

// ServerConfig configures the HTTP server settings.
// All timeouts use Go duration format (e.g., "30s", "5m", "1h").
type ServerConfig struct {
	// Host is the network interface to bind to.
	Host string `yaml:"host" default:"0.0.0.0" env:"SERVER_HOST"`

	// Port is the TCP port number to listen on.
	Port int `yaml:"port" default:"8080" env:"SERVER_PORT" validate:"min=1024,max=65535"`

	// Timeout is the maximum duration for reading/writing requests.
	Timeout time.Duration `yaml:"timeout" default:"30s" env:"SERVER_TIMEOUT"`

	// TLS contains TLS/HTTPS configuration.
	// Set to nil or omit to disable TLS.
	TLS *TLSConfig `yaml:"tls,omitempty"`
}

// TLSConfig holds TLS/HTTPS settings for secure connections.
// Both CertFile and KeyFile are required when Enabled is true.
type TLSConfig struct {
	// Enabled toggles TLS on/off.
	Enabled bool `yaml:"enabled" default:"false" env:"TLS_ENABLED"`

	// CertFile is the path to the TLS certificate file.
	CertFile string `yaml:"cert_file" env:"TLS_CERT_FILE" validate:"required_if=Enabled true"`

	// KeyFile is the path to the TLS private key file.
	KeyFile string `yaml:"key_file" env:"TLS_KEY_FILE" validate:"required_if=Enabled true"`
}

// DatabaseConfig holds configurations for different databases.
type DatabaseConfig struct {
	// Primary is the main SQL database connection.
	Primary SQLConfig `yaml:"primary"`

	// Redis is the Redis cache configuration.
	Redis RedisConfig `yaml:"redis"`

	// Analytics is an optional secondary SQL database.
	Analytics *SQLConfig `yaml:"analytics,omitempty"`
}

// SQLConfig configures a SQL database connection.
type SQLConfig struct {
	// Driver is the database driver name.
	Driver string `yaml:"driver" default:"postgres" validate:"required"`

	// DSN is the computed connection string.
	DSN string `yaml:"dsn,omitempty" dsn:"{{.User}}:{{.Password}}@tcp({{.Host}}:{{.Port}})/{{.Database}}"`

	// Host is the database server hostname.
	Host string `yaml:"host" default:"localhost" env:"DB_HOST"`

	// Port is the database server port number.
	Port int `yaml:"port" default:"5432" env:"DB_PORT"`

	// User is the database username.
	User string `yaml:"user" default:"postgres" env:"DB_USER"`

	// Password is the database password.
	Password string `yaml:"password" env:"DB_PASSWORD" validate:"required"`

	// Database is the name of the database to connect to.
	Database string `yaml:"database" default:"app_db" env:"DB_NAME"`
}

// RedisConfig configures a Redis connection.
type RedisConfig struct {
	// Address is the Redis server address in host:port format.
	Address string `yaml:"address" default:"localhost:6379" env:"REDIS_ADDR"`

	// Password is the Redis authentication password.
	Password string `yaml:"password,omitempty" refFrom:"PasswordFile"`

	// PasswordFile is the path to a file containing the Redis password.
	PasswordFile string `yaml:"password_file" default:"/run/secrets/redis_password" env:"REDIS_PASSWORD_FILE"`

	// DB is the Redis database index (0-15).
	DB int `yaml:"db" default:"0"`
}

// SecretsConfig demonstrates loading secrets from external sources.
type SecretsConfig struct {
	// APIKey is the authentication key for external API calls.
	APIKey string `yaml:"api_key" ref:"file:///run/secrets/api_key" validate:"required"`

	// JWTSecret is loaded from the path specified in JWTSecretPath.
	JWTSecret string `yaml:"jwt_secret,omitempty" refFrom:"JWTSecretPath"`

	// JWTSecretPath is the file path where the JWT secret is stored.
	JWTSecretPath string `yaml:"jwt_secret_path" default:"/run/secrets/jwt_secret" env:"JWT_SECRET_PATH"`

	// VaultToken is loaded from HashiCorp Vault secret storage.
	VaultToken string `yaml:"vault_token,omitempty" ref:"vault:///secret/data/app#token"`

	// Certificate is the TLS certificate loaded as raw bytes.
	Certificate []byte `yaml:"certificate,omitempty" ref:"file:///etc/ssl/certs/app.crt"`

	// PrivateKey is the TLS private key loaded from a dynamic path.
	PrivateKey []byte `yaml:"private_key,omitempty" refFrom:"PrivateKeyPath"`

	// PrivateKeyPath is the file path to the private key.
	PrivateKeyPath string `yaml:"private_key_path" default:"/etc/ssl/private/app.key" env:"PRIVATE_KEY_PATH"`
}

// Duration is a non-struct type alias — the parser should skip it.
type Duration int64

// unexportedConfig should NOT appear in FindAllStructs results.
type unexportedConfig struct {
	Secret string `yaml:"secret"`
}
