package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	BrowserPath    string        `json:"browser_path"`
	DefaultTimeout time.Duration `json:"default_timeout"`
	MaxRetries     int           `json:"max_retries"`
	UserAgent      string        `json:"user_agent"`
	OutputFormats  []string      `json:"output_formats"`
}

func LoadConfig() (*Config, error) {
	config := &Config{
		UserAgent:      getEnvWithDefault("SCREENSHOT_USER_AGENT", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		DefaultTimeout: getTimeoutFromEnv("SCREENSHOT_DEFAULT_TIMEOUT", 30*time.Second),
		OutputFormats:  []string{"original", "twitter", "linkedin"},
		BrowserPath:    getEnvWithDefault("SCREENSHOT_BROWSER_PATH", ""),
		MaxRetries:     getIntFromEnv("SCREENSHOT_MAX_RETRIES", 3),
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

func (c *Config) Validate() error {
	if c.DefaultTimeout <= 0 {
		return fmt.Errorf("default timeout must be positive")
	}

	if c.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}

	if c.UserAgent == "" {
		return fmt.Errorf("user agent cannot be empty")
	}

	if c.BrowserPath != "" {
		if _, err := os.Stat(c.BrowserPath); os.IsNotExist(err) {
			return fmt.Errorf("specified browser path does not exist: %s", c.BrowserPath)
		}
	}

	return nil
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntFromEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getTimeoutFromEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func DefaultConfig() *Config {
	return &Config{
		BrowserPath:    "",
		DefaultTimeout: 30 * time.Second,
		MaxRetries:     3,
		UserAgent:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		OutputFormats:  []string{"original", "twitter", "linkedin"},
	}
}

