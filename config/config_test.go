package config_test

import (
	"os"
	"testing"
	"time"

	"screenshot-tweets/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig()

	assert.Empty(t, cfg.BrowserPath)
	assert.Equal(t, 30*time.Second, cfg.DefaultTimeout)
	assert.Equal(t, 3, cfg.MaxRetries)
	assert.Contains(t, cfg.UserAgent, "Chrome")
	assert.Equal(t, []string{"original", "twitter", "linkedin"}, cfg.OutputFormats)
}

func TestLoadConfigDefaults(t *testing.T) {
	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Empty(t, cfg.BrowserPath)
	assert.Equal(t, 30*time.Second, cfg.DefaultTimeout)
	assert.Equal(t, 3, cfg.MaxRetries)
	assert.Contains(t, cfg.UserAgent, "Chrome")
}

func TestLoadConfigFromEnvironment(t *testing.T) {
	originalValues := make(map[string]string)
	envVars := map[string]string{
		"SCREENSHOT_DEFAULT_TIMEOUT": "45s",
		"SCREENSHOT_MAX_RETRIES":     "5",
		"SCREENSHOT_USER_AGENT":      "Test-Agent/1.0",
	}

	for key, value := range envVars {
		originalValues[key] = os.Getenv(key)
		os.Setenv(key, value)
	}

	defer func() {
		for key, originalValue := range originalValues {
			if originalValue == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, originalValue)
			}
		}
	}()

	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, 45*time.Second, cfg.DefaultTimeout)
	assert.Equal(t, 5, cfg.MaxRetries)
	assert.Equal(t, "Test-Agent/1.0", cfg.UserAgent)
}

func TestLoadConfigInvalidEnvironment(t *testing.T) {
	originalTimeout := os.Getenv("SCREENSHOT_DEFAULT_TIMEOUT")
	originalRetries := os.Getenv("SCREENSHOT_MAX_RETRIES")

	os.Setenv("SCREENSHOT_DEFAULT_TIMEOUT", "invalid-duration")
	os.Setenv("SCREENSHOT_MAX_RETRIES", "invalid-number")

	defer func() {
		if originalTimeout == "" {
			os.Unsetenv("SCREENSHOT_DEFAULT_TIMEOUT")
		} else {
			os.Setenv("SCREENSHOT_DEFAULT_TIMEOUT", originalTimeout)
		}
		if originalRetries == "" {
			os.Unsetenv("SCREENSHOT_MAX_RETRIES")
		} else {
			os.Setenv("SCREENSHOT_MAX_RETRIES", originalRetries)
		}
	}()

	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, 30*time.Second, cfg.DefaultTimeout)
	assert.Equal(t, 3, cfg.MaxRetries)
}

func TestConfigValidation(t *testing.T) {
	for _, test := range []struct {
		name          string
		config        *config.Config
		expectError   bool
		errorContains string
	}{
		{
			name:   "valid config",
			config: config.DefaultConfig(),
		},
		{
			name: "invalid timeout",
			config: &config.Config{
				DefaultTimeout: 0,
				MaxRetries:     3,
				UserAgent:      "test",
			},
			expectError:   true,
			errorContains: "timeout must be positive",
		},
		{
			name: "negative retries",
			config: &config.Config{
				DefaultTimeout: 30 * time.Second,
				MaxRetries:     -1,
				UserAgent:      "test",
			},
			expectError:   true,
			errorContains: "retries cannot be negative",
		},
		{
			name: "empty user agent",
			config: &config.Config{
				DefaultTimeout: 30 * time.Second,
				MaxRetries:     3,
				UserAgent:      "",
			},
			expectError:   true,
			errorContains: "user agent cannot be empty",
		},
		{
			name: "nonexistent browser path",
			config: &config.Config{
				DefaultTimeout: 30 * time.Second,
				MaxRetries:     3,
				UserAgent:      "test",
				BrowserPath:    "/non/existent/browser",
			},
			expectError:   true,
			errorContains: "browser path does not exist",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := test.config.Validate()
			if test.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.errorContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

