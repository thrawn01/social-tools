package screenshot

import (
	"fmt"
	"image"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
)

type SocialMediaPlatform struct {
	Name   string `json:"name"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

var PlatformConfigs = map[string]SocialMediaPlatform{
	"twitter":  {Name: "Twitter/X", Width: 1200, Height: 628},
	"linkedin": {Name: "LinkedIn", Width: 1200, Height: 627},
}

func ResizeForSocialMedia(originalFile, baseFilename string) error {
	img, err := imaging.Open(originalFile)
	if err != nil {
		return fmt.Errorf("failed to open image %s: %w", originalFile, err)
	}

	baseDir := filepath.Dir(originalFile)
	nameWithoutExt := strings.TrimSuffix(baseFilename, filepath.Ext(baseFilename))

	for platform, config := range PlatformConfigs {
		resizedImg := SmartCrop(img, config.Width, config.Height)

		platformFilename := fmt.Sprintf("%s-%s.png", nameWithoutExt, platform)
		platformPath := filepath.Join(baseDir, platformFilename)

		if err := imaging.Save(resizedImg, platformPath); err != nil {
			return fmt.Errorf("failed to save %s optimized image: %w", platform, err)
		}
	}

	return nil
}

func SmartCrop(img image.Image, targetWidth, targetHeight int) image.Image {
	bounds := img.Bounds()
	originalWidth := bounds.Dx()
	originalHeight := bounds.Dy()

	targetRatio := float64(targetWidth) / float64(targetHeight)
	originalRatio := float64(originalWidth) / float64(originalHeight)

	var resized image.Image

	if originalRatio > targetRatio {
		newHeight := int(float64(originalWidth) / targetRatio)
		if newHeight <= originalHeight {
			resized = imaging.Crop(img, image.Rect(0, 0, originalWidth, newHeight))
		} else {
			resized = imaging.Resize(img, targetWidth, 0, imaging.Lanczos)
		}
	} else if originalRatio < targetRatio {
		newWidth := int(float64(originalHeight) * targetRatio)
		if newWidth <= originalWidth {
			resized = imaging.Crop(img, image.Rect(0, 0, newWidth, originalHeight))
		} else {
			resized = imaging.Resize(img, 0, targetHeight, imaging.Lanczos)
		}
	} else {
		resized = imaging.Resize(img, targetWidth, targetHeight, imaging.Lanczos)
	}

	resized = imaging.Fill(resized, targetWidth, targetHeight, imaging.Center, imaging.Lanczos)

	return imaging.Sharpen(resized, 0.5)
}

func GenerateSocialMediaFilenames(day int) map[string]string {
	baseFilename := fmt.Sprintf("day-%d-screenshot", day)
	filenames := make(map[string]string)

	for platform := range PlatformConfigs {
		filenames[platform] = fmt.Sprintf("%s-%s.png", baseFilename, platform)
	}

	return filenames
}

func GenerateAllFilenames(day int) map[string]string {
	filenames := make(map[string]string)

	filenames["original"] = fmt.Sprintf("day-%d-screenshot.png", day)

	for platform, filename := range GenerateSocialMediaFilenames(day) {
		filenames[platform] = filename
	}

	return filenames
}
