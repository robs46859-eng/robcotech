package settings

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the gateway service
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Arkham   ArkhamConfig
	Auth     AuthConfig
	RateLimit RateLimitConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	PoolSize int
}

// ArkhamConfig holds Arkham security configuration
type ArkhamConfig struct {
	Enabled           bool
	Endpoint          string
	DeceptionEnabled  bool
	CrossTenantShare  bool
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret        string
	APIKeyHeader     string
	TokenExpiry      time.Duration
	RefreshTokenExpiry time.Duration
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled        bool
	RequestsPerMin int64
	BurstSize      int
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnvAsInt("SERVER_PORT", 8080),
			ReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:  getEnvAsDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DATABASE_HOST", "localhost"),
			Port:            getEnvAsInt("DATABASE_PORT", 5432),
			User:            getEnv("DATABASE_USER", "postgres"),
			Password:        getEnv("DATABASE_PASSWORD", "postgres"),
			Database:        getEnv("DATABASE_NAME", "fullstackarkham"),
			SSLMode:         getEnv("DATABASE_SSL_MODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DATABASE_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DATABASE_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDuration("DATABASE_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
			PoolSize: getEnvAsInt("REDIS_POOL_SIZE", 10),
		},
		Arkham: ArkhamConfig{
			Enabled:          getEnvAsBool("ARKHAM_ENABLED", true),
			Endpoint:         getEnv("ARKHAM_ENDPOINT", "http://localhost:8081"),
			DeceptionEnabled: getEnvAsBool("ARKHAM_DECEPTION_ENABLED", false),
			CrossTenantShare: getEnvAsBool("ARKHAM_CROSS_TENANT_SHARE", true),
		},
		Auth: AuthConfig{
			JWTSecret:        getEnv("JWT_SECRET", "change-me-in-production"),
			APIKeyHeader:     getEnv("API_KEY_HEADER", "X-API-Key"),
			TokenExpiry:      getEnvAsDuration("TOKEN_EXPIRY", 24*time.Hour),
			RefreshTokenExpiry: getEnvAsDuration("REFRESH_TOKEN_EXPIRY", 7*24*time.Hour),
		},
		RateLimit: RateLimitConfig{
			Enabled:        getEnvAsBool("RATE_LIMIT_ENABLED", true),
			RequestsPerMin: getEnvAsInt64("RATE_LIMIT_REQUESTS_PER_MIN", 60),
			BurstSize:      getEnvAsInt("RATE_LIMIT_BURST", 10),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if durationVal, err := time.ParseDuration(value); err == nil {
			return durationVal
		}
	}
	return defaultValue
}
