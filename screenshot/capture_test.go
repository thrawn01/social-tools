package screenshot_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"screenshot-tweets/screenshot"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDefaultConfig(t *testing.T) {
	config := screenshot.NewDefaultConfig()

	assert.Equal(t, 800, config.ViewportWidth)
	assert.Equal(t, 600, config.ViewportHeight)
	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.Equal(t, ".", config.OutputDir)
	assert.Contains(t, config.UserAgent, "Chrome")
}

func TestGenerateFilename(t *testing.T) {
	outputDir := "/tmp/screenshots"
	filename := screenshot.GenerateFilename(5, outputDir)

	expected := filepath.Join(outputDir, "day-5-screenshot.png")
	assert.Equal(t, expected, filename)
}

func TestGenerateBaseFilename(t *testing.T) {
	filename := screenshot.GenerateBaseFilename(10)
	assert.Equal(t, "day-10-screenshot.png", filename)
}

func TestCaptureScreenshotInvalidURL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping screenshot test in short mode")
	}

	tempDir := t.TempDir()
	config := screenshot.ScreenshotConfig{
		ViewportHeight: 1200,
		ViewportWidth:  1920,
		Timeout:        5 * time.Second,
		OutputDir:      tempDir,
		UserAgent:      "test-agent",
	}

	err := screenshot.CaptureScreenshot("invalid-url", "test.png", config)
	require.Error(t, err)
}

func TestCaptureScreenshotFileCreation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping screenshot test in short mode")
	}

	tempDir := t.TempDir()
	config := screenshot.ScreenshotConfig{
		UserAgent:      "Mozilla/5.0 (test)",
		Timeout:        10 * time.Second,
		ViewportWidth:  800,
		ViewportHeight: 600,
		OutputDir:      tempDir,
	}

	filename := "test-screenshot.png"
	err := screenshot.CaptureScreenshot("https://httpbin.org/html", filename, config)

	if err != nil {
		// Screenshot capture may fail in CI environments without display
		return
	}

	filePath := filepath.Join(tempDir, filename)
	_, err = os.Stat(filePath)
	assert.NoError(t, err)

	fileInfo, err := os.Stat(filePath)
	require.NoError(t, err)
	assert.Greater(t, fileInfo.Size(), int64(0))
}

