package errors_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	apperrors "screenshot-tweets/internal/errors"

	"github.com/stretchr/testify/assert"
)

func TestNewScreenshotError(t *testing.T) {
	originalErr := fmt.Errorf("connection timeout")
	screenshotErr := apperrors.NewScreenshotError("https://example.com", 5, originalErr)

	assert.Equal(t, "https://example.com", screenshotErr.URL)
	assert.Equal(t, 5, screenshotErr.Day)
	assert.Equal(t, "timeout", screenshotErr.ErrorType)
	assert.Equal(t, originalErr.Error(), screenshotErr.Message)
	assert.WithinDuration(t, time.Now(), screenshotErr.Timestamp, time.Second)
}

func TestScreenshotError_Error(t *testing.T) {
	err := &apperrors.ScreenshotError{
		URL:       "https://example.com",
		Day:       3,
		ErrorType: "timeout",
		Message:   "request timeout",
		Timestamp: time.Now(),
	}

	expected := "screenshot error (Day 3, timeout): https://example.com - request timeout"
	assert.Equal(t, expected, err.Error())
}

func TestCategorizeError(t *testing.T) {
	for _, test := range []struct {
		errorMessage string
		expectedType string
	}{
		{"connection timeout", "timeout"},
		{"context deadline exceeded", "timeout"},
		{"404 not found", "not_found"},
		{"403 forbidden", "forbidden"},
		{"500 internal server error", "server_error"},
		{"502 bad gateway", "server_error"},
		{"503 service unavailable", "server_error"},
		{"dns lookup failed", "dns_error"},
		{"no such host", "dns_error"},
		{"connection refused", "connection_error"},
		{"connection reset by peer", "connection_error"},
		{"failed to launch browser", "browser_error"},
		{"some unknown error", "unknown"},
	} {
		t.Run(test.expectedType, func(t *testing.T) {
			originalErr := fmt.Errorf(test.errorMessage)
			screenshotErr := apperrors.NewScreenshotError("https://example.com", 1, originalErr)
			assert.Equal(t, test.expectedType, screenshotErr.ErrorType)
		})
	}
}

func TestCategorizeNetworkError(t *testing.T) {
	netErr := &net.DNSError{
		Err:    "no such host",
		Name:   "example.com",
		Server: "8.8.8.8",
	}

	screenshotErr := apperrors.NewScreenshotError("https://example.com", 1, netErr)
	assert.Equal(t, "dns_error", screenshotErr.ErrorType)
}

func TestIsRetryableError(t *testing.T) {
	for _, test := range []struct {
		name      string
		errorType string
		expected  bool
	}{
		{"timeout error", "timeout", true},
		{"server error", "server_error", true},
		{"network error", "network_error", true},
		{"connection error", "connection_error", true},
		{"not found error", "not_found", false},
		{"forbidden error", "forbidden", false},
		{"dns error", "dns_error", false},
		{"browser error", "browser_error", false},
		{"unknown error", "unknown", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := &apperrors.ScreenshotError{
				ErrorType: test.errorType,
			}
			assert.Equal(t, test.expected, apperrors.IsRetryableError(err))
		})
	}
}

func TestIsRetryableErrorGeneric(t *testing.T) {
	for _, test := range []struct {
		errorMessage string
		expected     bool
	}{
		{"connection timeout", true},
		{"context deadline exceeded", true},
		{"connection refused", true},
		{"502 bad gateway", true},
		{"503 service unavailable", true},
		{"504 gateway timeout", true},
		{"404 not found", false},
		{"some random error", false},
	} {
		t.Run(test.errorMessage, func(t *testing.T) {
			err := fmt.Errorf(test.errorMessage)
			assert.Equal(t, test.expected, apperrors.IsRetryableError(err))
		})
	}
}

func TestNewDefaultRetryConfig(t *testing.T) {
	config := apperrors.NewDefaultRetryConfig()

	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 1*time.Second, config.InitialDelay)
	assert.Equal(t, 2.0, config.BackoffFactor)
	assert.Equal(t, 30*time.Second, config.MaxDelay)
	assert.Contains(t, config.RetryableErrors, "timeout")
	assert.Contains(t, config.RetryableErrors, "server_error")
}

func TestRetryConfig_ShouldRetry(t *testing.T) {
	config := apperrors.NewDefaultRetryConfig()

	timeoutErr := &apperrors.ScreenshotError{ErrorType: "timeout"}
	assert.True(t, config.ShouldRetry(timeoutErr, 0))
	assert.True(t, config.ShouldRetry(timeoutErr, 1))
	assert.True(t, config.ShouldRetry(timeoutErr, 2))
	assert.False(t, config.ShouldRetry(timeoutErr, 3))

	notFoundErr := &apperrors.ScreenshotError{ErrorType: "not_found"}
	assert.False(t, config.ShouldRetry(notFoundErr, 0))
}

func TestRetryConfig_GetDelay(t *testing.T) {
	config := apperrors.RetryConfig{
		InitialDelay:  1 * time.Second,
		BackoffFactor: 2.0,
		MaxDelay:      10 * time.Second,
	}

	assert.Equal(t, 2*time.Second, config.GetDelay(1))
	assert.Equal(t, 4*time.Second, config.GetDelay(2))
	assert.Equal(t, 8*time.Second, config.GetDelay(3))
	assert.Equal(t, 10*time.Second, config.GetDelay(4)) // Should be capped at MaxDelay
}

func TestIsRetryableContextError(t *testing.T) {
	ctxErr := context.DeadlineExceeded
	assert.True(t, apperrors.IsRetryableError(ctxErr))
}

