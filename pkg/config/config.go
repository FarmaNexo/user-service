// pkg/config/config.go
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config representa toda la configuración del servicio
type Config struct {
	Environment string         `mapstructure:"environment"`
	Server      ServerConfig   `mapstructure:"server"`
	Database    DatabaseConfig `mapstructure:"database"`
	JWT         JWTConfig      `mapstructure:"jwt"`
	AWS         AWSConfig      `mapstructure:"aws"`
	SQS         SQSConfig      `mapstructure:"sqs"`
	S3          S3Config       `mapstructure:"s3"`
	Log         LogConfig      `mapstructure:"log"`
}

// ServerConfig configuración del servidor HTTP
type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	Host         string        `mapstructure:"host"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// DatabaseConfig configuración de PostgreSQL
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"db_name"`
	Schema          string        `mapstructure:"schema"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// JWTConfig configuración de JWT (solo validación, no generación)
type JWTConfig struct {
	Secret              string        `mapstructure:"secret"`
	AccessTokenDuration time.Duration `mapstructure:"access_token_duration"`
	Issuer              string        `mapstructure:"issuer"`
}

// AWSConfig configuración de AWS
type AWSConfig struct {
	Region   string `mapstructure:"region"`
	Endpoint string `mapstructure:"endpoint"`
}

// SQSConfig configuración de SQS
type SQSConfig struct {
	UserEventsQueueURL string `mapstructure:"user_events_queue_url"`
	AuthEventsQueueURL string `mapstructure:"auth_events_queue_url"`
}

// S3Config configuración de S3
type S3Config struct {
	AvatarsBucket string `mapstructure:"avatars_bucket"`
}

// LogConfig configuración de logging
type LogConfig struct {
	Level    string `mapstructure:"level"`
	Encoding string `mapstructure:"encoding"`
}

// ========================================
// LOAD CONFIG
// ========================================

// LoadConfig carga la configuración basada en el environment
func LoadConfig(environment string) (*Config, error) {
	v := viper.New()

	v.SetConfigName(fmt.Sprintf("config.%s", environment))
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath("../configs")
	v.AddConfigPath("../../configs")

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	config.Environment = environment

	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

// ========================================
// VALIDATION
// ========================================

func validateConfig(config *Config) error {
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if config.Database.DBName == "" {
		return fmt.Errorf("database name is required")
	}

	if config.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required")
	}
	if len(config.JWT.Secret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters")
	}

	return nil
}

// ========================================
// HELPERS
// ========================================

// GetDSN retorna el Data Source Name para PostgreSQL
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.DBName,
		c.SSLMode,
	)
}

func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

func (c *Config) IsLocal() bool {
	return c.Environment == "local"
}

func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

func (c *Config) IsQA() bool {
	return c.Environment == "qa"
}

func (c *Config) IsUAT() bool {
	return c.Environment == "uat"
}
