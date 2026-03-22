package configs

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server    ServerConfig
	Sonyflake SonyflakeConfig
	Database  DatabaseConfig
	Postgres  PostgresConfig
}

type ServerConfig struct {
	Port         string
	Host         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type SonyflakeConfig struct {
	StartTime    time.Time
	SequenceBits uint64
}

type DatabaseConfig struct {
	UseInMemory bool // if true PostgresConfig fields are ignored
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 10*time.Second),
		},
		Sonyflake: SonyflakeConfig{
			StartTime:    getTimeEnv("SONYFLAKE_START_TIME", time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)),
			SequenceBits: getUintEnv("SONYFLAKE_SEQUENCE_BITS", 10),
		},
		Database: DatabaseConfig{
			UseInMemory: getBoolEnv("DB_USE_IN_MEMORY", false),
		},
		Postgres: PostgresConfig{
			Host:     getEnv("POSTGRES_HOST", "postgres"),
			Port:     getEnv("POSTGRES_PORT", "5432"),
			User:     getEnv("POSTGRES_USER", "postgres"),
			Password: getEnv("POSTGRES_PASSWORD", "postgres"),
			Name:     getEnv("POSTGRES_NAME", "cloud_storage"),
			SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getUintEnv(key string, defaultValue uint64) uint64 {
	if value := os.Getenv(key); value != "" {
		if uintVal, err := strconv.ParseUint(value, 10, 64); err == nil {
			return uintVal
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getTimeEnv(key string, defaultValue time.Time) time.Time {
	if value := os.Getenv(key); value != "" {
		if timeVal, err := time.Parse(time.RFC3339, value); err == nil {
			return timeVal
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if durationVal, err := time.ParseDuration(value); err == nil {
			return durationVal
		}
	}
	return defaultValue
}
