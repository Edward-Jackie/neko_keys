package config

import (
	"fmt"
	"os"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	JWTSecret     string
	Port          string
	AdminUsername string
	AdminPassword string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (*Config, error) {
	cfg := &Config{
		DBHost:        getEnvOrDefault("DB_HOST", "localhost"),
		DBPort:        getEnvOrDefault("DB_PORT", "3306"),
		DBUser:        getEnvOrDefault("DB_USER", "root"),
		DBPassword:    getEnvOrDefault("DB_PASSWORD", ""),
		DBName:        getEnvOrDefault("DB_NAME", "neko_keys"),
		JWTSecret:     getEnvOrDefault("JWT_SECRET", "neko-secret-key"),
		Port:          getEnvOrDefault("PORT", "9002"),
		AdminUsername: getEnvOrDefault("ADMIN_USERNAME", "admin"),
		AdminPassword: getEnvOrDefault("ADMIN_PASSWORD", "admin123"),
	}
	return cfg, nil
}

// DSN returns the MySQL data source name.
func (c *Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
