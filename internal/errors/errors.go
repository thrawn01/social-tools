package errors

import (
	"fmt"
	"net"
	"strings"
	"time"
)

type ScreenshotError struct {
	URL       string    `json:"url"`
	Day       int       `json:"day"`
	ErrorType string    `json:"error_type"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

func (e ScreenshotError) Error() string {
	return fmt.Sprintf("screenshot error (Day %d, %s): %s - %s", e.Day, e.ErrorType, e.URL, e.Message)
}

func NewScreenshotError(url string, day int, err error) *ScreenshotError {
	errorType := categorizeError(err)
	return &ScreenshotError{
		URL:       url,
		Day:       day,
		ErrorType: errorType,
		Message:   err.Error(),
		Timestamp: time.Now(),
	}
}

func categorizeError(err error) string {
	errStr := strings.ToLower(err.Error())

	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "context deadline exceeded") {
		return "timeout"
	}

	if strings.Contains(errStr, "404") || strings.Contains(errStr, "not found") {
		return "not_found"
	}

	if strings.Contains(errStr, "403") || strings.Contains(errStr, "forbidden") {
		return "forbidden"
	}

	if strings.Contains(errStr, "500") || strings.Contains(errStr, "502") || strings.Contains(errStr, "503") {
		return "server_error"
	}

	if strings.Contains(errStr, "dns") || strings.Contains(errStr, "no such host") {
		return "dns_error"
	}

	if _, ok := err.(net.Error); ok {
		return "network_error"
	}

	if strings.Contains(errStr, "connection refused") || strings.Contains(errStr, "connection reset") {
		return "connection_error"
	}

	if strings.Contains(errStr, "browser") || strings.Contains(errStr, "launch") {
		return "browser_error"
	}

	return "unknown"
}

func IsRetryableError(err error) bool {
	if screenshotErr, ok := err.(*ScreenshotError); ok {
		switch screenshotErr.ErrorType {
		case "timeout", "server_error", "network_error", "connection_error":
			return true
		case "not_found", "forbidden", "dns_error", "browser_error":
			return false
		default:
			return false
		}
	}

	errStr := strings.ToLower(err.Error())
	retryablePatterns := []string{
		"timeout",
		"context deadline exceeded",
		"connection refused",
		"connection reset",
		"502",
		"503",
		"504",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}

type RetryConfig struct {
	MaxRetries      int           `json:"max_retries"`
	InitialDelay    time.Duration `json:"initial_delay"`
	BackoffFactor   float64       `json:"backoff_factor"`
	MaxDelay        time.Duration `json:"max_delay"`
	RetryableErrors []string      `json:"retryable_errors"`
}

func NewDefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:    3,
		InitialDelay:  1 * time.Second,
		BackoffFactor: 2.0,
		MaxDelay:      30 * time.Second,
		RetryableErrors: []string{
			"timeout",
			"server_error",
			"network_error",
			"connection_error",
		},
	}
}

func (rc RetryConfig) ShouldRetry(err error, attempt int) bool {
	if attempt >= rc.MaxRetries {
		return false
	}
	return IsRetryableError(err)
}

func (rc RetryConfig) GetDelay(attempt int) time.Duration {
	multiplier := 1.0
	for i := 0; i < attempt; i++ {
		multiplier *= rc.BackoffFactor
	}
	delay := time.Duration(float64(rc.InitialDelay) * multiplier)
	if delay > rc.MaxDelay {
		return rc.MaxDelay
	}
	return delay
}
