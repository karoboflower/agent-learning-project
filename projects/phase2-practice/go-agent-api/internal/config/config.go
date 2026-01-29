package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	OpenAI   OpenAIConfig
	Redis    RedisConfig
	Postgres PostgresConfig
	Agent    AgentConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port    string
	GinMode string
}

// OpenAIConfig holds OpenAI API configuration
type OpenAIConfig struct {
	APIKey string
	Model  string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// PostgresConfig holds PostgreSQL configuration
type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// AgentConfig holds agent-specific configuration
type AgentConfig struct {
	MaxConcurrent int
	TaskTimeout   int
	MaxRetries    int
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if exists
	_ = godotenv.Load()

	config := &Config{
		Server: ServerConfig{
			Port:    getEnv("SERVER_PORT", "8080"),
			GinMode: getEnv("GIN_MODE", "debug"),
		},
		OpenAI: OpenAIConfig{
			APIKey: getEnv("OPENAI_API_KEY", ""),
			Model:  getEnv("OPENAI_MODEL", "gpt-4"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Postgres: PostgresConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnv("POSTGRES_PORT", "5432"),
			User:     getEnv("POSTGRES_USER", "postgres"),
			Password: getEnv("POSTGRES_PASSWORD", ""),
			DBName:   getEnv("POSTGRES_DB", "agent_api"),
			SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
		},
		Agent: AgentConfig{
			MaxConcurrent: getEnvAsInt("MAX_CONCURRENT_AGENTS", 10),
			TaskTimeout:   getEnvAsInt("TASK_TIMEOUT", 300),
			MaxRetries:    getEnvAsInt("MAX_RETRIES", 3),
		},
	}

	// Validate required fields
	if config.OpenAI.APIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is required")
	}

	return config, nil
}

// getEnv gets environment variable or returns default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets environment variable as int or returns default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetDSN returns PostgreSQL connection string
func (c *PostgresConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// GetRedisAddr returns Redis connection address
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}
