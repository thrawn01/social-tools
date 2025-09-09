package screenshot_test

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"screenshot-tweets/screenshot"

	"github.com/disintegration/imaging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if (x+y)%20 < 10 {
				img.Set(x, y, color.RGBA{R: 100, G: 150, B: 200, A: 255})
			} else {
				img.Set(x, y, color.RGBA{R: 200, G: 100, B: 50, A: 255})
			}
		}
	}

	return img
}

func TestPlatformConfigs(t *testing.T) {
	assert.Contains(t, screenshot.PlatformConfigs, "twitter")
	assert.Contains(t, screenshot.PlatformConfigs, "linkedin")

	twitterConfig := screenshot.PlatformConfigs["twitter"]
	assert.Equal(t, 1200, twitterConfig.Width)
	assert.Equal(t, 628, twitterConfig.Height)

	linkedinConfig := screenshot.PlatformConfigs["linkedin"]
	assert.Equal(t, 1200, linkedinConfig.Width)
	assert.Equal(t, 627, linkedinConfig.Height)
}

func TestSmartCrop(t *testing.T) {
	for _, test := range []struct {
		name           string
		originalWidth  int
		originalHeight int
		targetWidth    int
		targetHeight   int
	}{
		{"landscape to square", 1920, 1080, 1200, 628},
		{"square to landscape", 1000, 1000, 1200, 627},
		{"portrait to landscape", 600, 800, 1200, 628},
		{"small to large", 400, 300, 1200, 628},
	} {
		t.Run(test.name, func(t *testing.T) {
			original := createTestImage(test.originalWidth, test.originalHeight)
			cropped := screenshot.SmartCrop(original, test.targetWidth, test.targetHeight)

			bounds := cropped.Bounds()
			assert.Equal(t, test.targetWidth, bounds.Dx())
			assert.Equal(t, test.targetHeight, bounds.Dy())
		})
	}
}

func TestResizeForSocialMedia(t *testing.T) {
	tempDir := t.TempDir()

	originalImg := createTestImage(1920, 1080)
	originalFile := filepath.Join(tempDir, "original.png")

	err := imaging.Save(originalImg, originalFile)
	require.NoError(t, err)

	baseFilename := "test-screenshot.png"
	err = screenshot.ResizeForSocialMedia(originalFile, baseFilename)
	require.NoError(t, err)

	twitterFile := filepath.Join(tempDir, "test-screenshot-twitter.png")
	linkedinFile := filepath.Join(tempDir, "test-screenshot-linkedin.png")

	_, err = os.Stat(twitterFile)
	assert.NoError(t, err)

	_, err = os.Stat(linkedinFile)
	assert.NoError(t, err)

	twitterImg, err := imaging.Open(twitterFile)
	require.NoError(t, err)
	twitterBounds := twitterImg.Bounds()
	assert.Equal(t, 1200, twitterBounds.Dx())
	assert.Equal(t, 628, twitterBounds.Dy())

	linkedinImg, err := imaging.Open(linkedinFile)
	require.NoError(t, err)
	linkedinBounds := linkedinImg.Bounds()
	assert.Equal(t, 1200, linkedinBounds.Dx())
	assert.Equal(t, 627, linkedinBounds.Dy())
}

func TestResizeForSocialMediaInvalidFile(t *testing.T) {
	err := screenshot.ResizeForSocialMedia("/non/existent/file.png", "test.png")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open image")
}

func TestGenerateSocialMediaFilenames(t *testing.T) {
	filenames := screenshot.GenerateSocialMediaFilenames(7)

	assert.Contains(t, filenames, "twitter")
	assert.Contains(t, filenames, "linkedin")

	assert.Equal(t, "day-7-screenshot-twitter.png", filenames["twitter"])
	assert.Equal(t, "day-7-screenshot-linkedin.png", filenames["linkedin"])
}

func TestGenerateAllFilenames(t *testing.T) {
	filenames := screenshot.GenerateAllFilenames(3)

	assert.Contains(t, filenames, "original")
	assert.Contains(t, filenames, "twitter")
	assert.Contains(t, filenames, "linkedin")

	assert.Equal(t, "day-3-screenshot.png", filenames["original"])
	assert.Equal(t, "day-3-screenshot-twitter.png", filenames["twitter"])
	assert.Equal(t, "day-3-screenshot-linkedin.png", filenames["linkedin"])
}

