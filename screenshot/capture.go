package screenshot

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

const (
	filePermissions = 0644
	httpTimeout     = 10 * time.Second
)

var (
	youtubeURLRegex   = regexp.MustCompile(`^https?://(www\.)?(youtube\.com/watch\?v=|youtu\.be/|youtube\.com/embed/)`)
	youtubeIDPatterns = []*regexp.Regexp{
		regexp.MustCompile(`youtube\.com/watch\?v=([a-zA-Z0-9_-]{11})`),
		regexp.MustCompile(`youtu\.be/([a-zA-Z0-9_-]{11})`),
		regexp.MustCompile(`youtube\.com/embed/([a-zA-Z0-9_-]{11})`),
	}
)

type ScreenshotConfig struct {
	ViewportWidth  int           `json:"viewport_width"`
	ViewportHeight int           `json:"viewport_height"`
	Timeout        time.Duration `json:"timeout"`
	OutputDir      string        `json:"output_dir"`
	UserAgent      string        `json:"user_agent"`
}

func NewDefaultConfig() ScreenshotConfig {
	return ScreenshotConfig{
		ViewportWidth:  800,
		ViewportHeight: 600,
		Timeout:        30 * time.Second,
		OutputDir:      ".",
		UserAgent:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}
}

func CaptureScreenshot(url, filename string, config ScreenshotConfig) error {
	// Check if URL is YouTube and try thumbnail extraction first
	if isYouTubeURL(url) {
		if videoID, err := extractYouTubeVideoID(url); err == nil {
			if err := downloadThumbnailWithFallback(videoID, filepath.Join(config.OutputDir, filename)); err == nil {
				return nil
			}
		}
	}

	// Fall back to regular browser screenshot
	return captureRegularScreenshot(url, filename, config)
}

func captureRegularScreenshot(url, filename string, config ScreenshotConfig) error {
	launcher := launcher.New().Headless(true)

	u, err := launcher.Launch()
	if err != nil {
		return fmt.Errorf("failed to launch browser: %w", err)
	}

	browser := rod.New().ControlURL(u)
	if err := browser.Connect(); err != nil {
		launcher.Cleanup()
		return fmt.Errorf("failed to connect to browser: %w", err)
	}
	defer func() {
		browser.Close()
		launcher.Cleanup()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	page, err := browser.Context(ctx).Page(proto.TargetCreateTarget{URL: ""})
	if err != nil {
		return fmt.Errorf("failed to create page: %w", err)
	}
	defer page.Close()

	if err := page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width:  config.ViewportWidth,
		Height: config.ViewportHeight,
	}); err != nil {
		return fmt.Errorf("failed to set viewport: %w", err)
	}

	if err := page.SetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent: config.UserAgent,
	}); err != nil {
		return fmt.Errorf("failed to set user agent: %w", err)
	}

	if err := page.Navigate(url); err != nil {
		return fmt.Errorf("failed to navigate to URL: %w", err)
	}

	if err := WaitForPageLoad(page); err != nil {
		return fmt.Errorf("failed waiting for page load: %w", err)
	}

	screenshot, err := page.Screenshot(false, &proto.PageCaptureScreenshot{
		Format: proto.PageCaptureScreenshotFormatPng,
	})
	if err != nil {
		return fmt.Errorf("failed to capture screenshot: %w", err)
	}

	if err := os.WriteFile(filepath.Join(config.OutputDir, filename), screenshot, filePermissions); err != nil {
		return fmt.Errorf("failed to write screenshot file: %w", err)
	}

	return nil
}

func WaitForPageLoad(page *rod.Page) error {
	if err := page.WaitLoad(); err != nil {
		return fmt.Errorf("failed waiting for DOM load: %w", err)
	}

	_, err := page.Eval(`() => document.fonts.ready`)
	if err != nil {
		return fmt.Errorf("failed waiting for fonts: %w", err)
	}

	page.WaitRequestIdle(3*time.Second, []string{}, []string{}, []proto.NetworkResourceType{})

	if err := page.WaitStable(1 * time.Second); err != nil {
		return fmt.Errorf("failed waiting for content stability: %w", err)
	}

	time.Sleep(2 * time.Second)

	return nil
}

func GenerateFilename(day int, outputDir string) string {
	filename := fmt.Sprintf("day-%d-screenshot.png", day)
	return filepath.Join(outputDir, filename)
}

func GenerateBaseFilename(day int) string {
	return fmt.Sprintf("day-%d-screenshot.png", day)
}

func isYouTubeURL(url string) bool {
	return youtubeURLRegex.MatchString(url)
}

func extractYouTubeVideoID(url string) (string, error) {
	for _, re := range youtubeIDPatterns {
		matches := re.FindStringSubmatch(url)
		if len(matches) > 1 {
			return matches[1], nil
		}
	}

	return "", fmt.Errorf("could not extract video ID from URL: %s", url)
}

func downloadThumbnailWithFallback(videoID, outputPath string) error {
	qualities := []string{"maxresdefault", "hqdefault", "mqdefault", "default"}

	for _, quality := range qualities {
		if err := downloadThumbnail(videoID, quality, outputPath); err == nil {
			return nil
		}
	}

	return fmt.Errorf("failed to download thumbnail for video ID: %s", videoID)
}

func downloadThumbnail(videoID, quality, outputPath string) error {
	client := &http.Client{Timeout: httpTimeout}

	resp, err := client.Get(fmt.Sprintf("https://img.youtube.com/vi/%s/%s.jpg", videoID, quality))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write thumbnail: %w", err)
	}

	return nil
}

