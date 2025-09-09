# Compact Screenshots Implementation Plan

## Overview

We're implementing a solution to fix oversized screenshots by:
1. Limiting regular website screenshots to 1200px height (viewport-only capture)
2. Extracting YouTube video thumbnails directly instead of full-page screenshots
3. Making this the new default behavior for the screenshot-tweets tool

## Current State Analysis

**Root Problem**: The tool currently uses `page.Screenshot(true, ...)` at `screenshot/capture.go:83` which captures full-page height, resulting in very large screenshots.

**Current Implementation**:
- Uses Rod browser automation library
- Fixed 1920x1080 viewport in `screenshot/capture.go:25-26` 
- All URLs processed identically through same pipeline
- No URL-type detection in `markdown/parser.go:77-79`
- Post-processing resize exists but doesn't address core issue

### Key Discoveries:
- Screenshot capture at `screenshot/capture.go:83` uses full-page mode: `page.Screenshot(true, ...)`
- Configuration managed via `ScreenshotConfig` struct at `screenshot/capture.go:15-21`
- YouTube URL example from input: `https://youtu.be/Kf5-HWJPTIE?si=01AaOhARAG9tfHFp`
- Existing resize functionality in `screenshot/resize.go` for social media formats

## Desired End State

After implementation completion:
1. Regular websites: Screenshots limited to 1200px height from top of page
2. YouTube URLs: Direct thumbnail extraction (no browser screenshot)
3. All existing functionality preserved (markdown annotation, file naming)
4. Tool validates correctly with existing tests

**Verification**: Run `screenshot-tweets -f tweets-context-engineering-week1.md -v` and confirm:
- Screenshots are ~1200px height maximum
- YouTube URLs produce thumbnail images instead of full page screenshots
- All markdown files are annotated correctly with "Screen Shot:" entries

## What We're NOT Doing

- Adding new command-line flags (this becomes the default behavior)
- Changing existing file naming conventions
- Modifying the markdown parsing logic
- Supporting other video platforms beyond YouTube
- Creating a configuration system for screenshot height

## Implementation Approach

**Phase 1**: Implement viewport-only screenshot capture for regular websites
**Phase 2**: Add YouTube URL detection and thumbnail extraction
**Phase 3**: Integration testing and validation

## Phase 1: Viewport-Only Screenshot Capture

### Overview
Modify the screenshot capture logic to capture only viewport-height content (1200px) instead of full-page content.

### Changes Required:

#### 1. Screenshot Capture Logic
**File**: `screenshot/capture.go`
**Changes**: Modify screenshot capture method to use viewport-only mode with 1200px height limit and add YouTube URL handling

```go
// Function signature to modify (interface unchanged)
func CaptureScreenshot(url, filename string, config ScreenshotConfig) error

// New internal functions to add
func isYouTubeURL(url string) bool
func extractYouTubeVideoID(url string) (string, error)
func captureYouTubeThumbnail(videoID, outputPath string) error
func captureViewportScreenshot(page *rod.Page, config ScreenshotConfig) ([]byte, error)
```

**Function Responsibilities:**
- Route YouTube URLs to thumbnail extraction using `isYouTubeURL()` check
- For regular URLs: Set viewport to 1920x1200 (increased from 1080) and use `page.Screenshot(false, ...)` for viewport-only
- YouTube detection using regex: `^https?://(www\\.)?(youtube\\.com/watch\\?v=|youtu\\.be/)([a-zA-Z0-9_-]{11})`
- Extract 11-character video IDs from both youtube.com and youtu.be formats
- Download thumbnails with fallback: maxresdefault → hqdefault → mqdefault → default
- If YouTube thumbnail fails, fall back to browser screenshot as backup
- Maintain existing error handling patterns from `capture.go:87-88`

#### 2. Configuration Updates
**File**: `screenshot/capture.go` AND `cmd/screenshot-tweets/main.go`
**Changes**: Update ScreenshotConfig struct and main.go config construction

```go
// Updated ScreenshotConfig struct in screenshot/capture.go
type ScreenshotConfig struct {
    ViewportWidth  int           `json:"viewport_width"`
    ViewportHeight int           `json:"viewport_height"`  // Update default to 1200
    Timeout        time.Duration `json:"timeout"`
    OutputDir      string        `json:"output_dir"`
    UserAgent      string        `json:"user_agent"`
}

// Updated default config function
func NewDefaultConfig() ScreenshotConfig
```

**Function Responsibilities:**
- Update `NewDefaultConfig()` to set `ViewportHeight: 1200` (changed from 1080) at line 26
- Update main.go config construction at lines 79-85 to use `ViewportHeight: 1200`
- Remove the separate MaxHeight field to keep config struct simple
- Maintain backward compatibility with existing config fields
- Follow struct field visual tapering pattern from guidelines

**Testing Requirements:**
```go
func TestViewportOnlyCapture(t *testing.T)
func TestScreenshotConfigDefaults(t *testing.T)
```

**Test Objectives:**
- Verify viewport-only screenshots are captured with correct dimensions
- Confirm MaxHeight configuration is applied correctly
- Validate existing screenshot functionality remains intact

**Context for implementation:**
- Follow error handling pattern from `capture.go:87-88`
- Use existing viewport setting logic from `capture.go:60-65` as reference
- Maintain browser setup sequence from `capture.go:34-58`

## Phase 2: YouTube Thumbnail Extraction

### Overview
Complete the YouTube URL detection and thumbnail extraction functionality within the screenshot package to avoid circular dependencies.

### Changes Required:

#### 1. YouTube Thumbnail Download Logic
**File**: `screenshot/capture.go` (add to existing file)
**Changes**: Add YouTube thumbnail download functions to screenshot package

```go
// New internal functions to add to screenshot/capture.go
func buildThumbnailURL(videoID, quality string) string
func downloadThumbnailWithFallback(videoID, outputPath string) error
func downloadThumbnail(videoID, quality, outputPath string) error
```

**Function Responsibilities:**
- Construct thumbnail URLs using pattern: `https://img.youtube.com/vi/<VIDEO_ID>/<quality>.jpg`
- Download thumbnails with quality fallback order: ["maxresdefault", "hqdefault", "mqdefault", "default"]
- Use `http.Get()` and `io.Copy` for memory-efficient streaming download
- Handle HTTP 404 errors for missing thumbnail qualities and continue fallback
- Create output file using existing patterns from `capture.go:90-92`
- Return first successful download or error if all qualities fail

#### 2. YouTube URL Processing Integration
**File**: `screenshot/capture.go` (modify existing CaptureScreenshot function)
**Changes**: Update main CaptureScreenshot function to handle both regular and YouTube URLs

```go
// Updated CaptureScreenshot function logic:
func CaptureScreenshot(url, filename string, config ScreenshotConfig) error {
    // Check if URL is YouTube
    if isYouTubeURL(url) {
        // Extract video ID and download thumbnail
        if videoID, err := extractYouTubeVideoID(url); err == nil {
            outputPath := filepath.Join(config.OutputDir, filename)
            if err := downloadThumbnailWithFallback(videoID, outputPath); err == nil {
                return nil  // Success with thumbnail
            }
            // Fall through to browser screenshot if thumbnail fails
        }
    }
    
    // Regular browser screenshot (existing logic with viewport-only mode)
    return captureRegularScreenshot(url, filename, config)
}
```

**Function Responsibilities:**
- First attempt YouTube thumbnail extraction for YouTube URLs
- Fall back to browser screenshot if thumbnail extraction fails
- Route regular URLs directly to viewport-only browser capture
- Maintain existing error handling and file path construction
- Preserve all existing function interfaces and behavior

**Testing Requirements:**
```go
func TestYouTubeURLDetection(t *testing.T)
func TestVideoIDExtraction(t *testing.T)
func TestThumbnailDownload(t *testing.T)
func TestYouTubeScreenshotRouting(t *testing.T)
func TestThumbnailFallback(t *testing.T)
```

**Test Objectives:**
- Verify YouTube URL detection for both youtube.com and youtu.be formats including parameters
- Confirm video ID extraction handles URLs like `https://youtu.be/Kf5-HWJPTIE?si=01AaOhARAG9tfHFp`
- Test thumbnail download with specific video IDs and quality fallback logic
- Validate screenshot routing chooses thumbnail vs browser capture correctly
- Test fallback to browser screenshot when YouTube thumbnail completely fails

**Context for implementation:**
- Add regex import: `import "regexp"` to screenshot/capture.go
- Add HTTP imports: `import "net/http"` and `import "io"` for thumbnail download
- Use file creation patterns from `capture.go:90-92`
- Follow error wrapping pattern from `capture.go:87-88`
- Test with real YouTube video IDs for integration tests

## Phase 3: Integration and Validation

### Overview
Integration testing and validation of complete compact screenshot functionality.

### Changes Required:

#### 1. Main Application Integration
**File**: `cmd/screenshot-tweets/main.go`
**Changes**: Update viewport height in config construction

**Function Responsibilities:**
- Update config construction at lines 81-82 to change `ViewportHeight: 1080` to `ViewportHeight: 1200`
- Existing `runScreenshotAutomation()` should work unchanged at line 51
- Error handling and progress reporting should work with both screenshot types
- No changes needed to command-line flag handling or file processing logic

#### 2. End-to-End Testing
**File**: Update existing test files
**Changes**: Enhance existing tests to cover new functionality

```go
func TestCompactScreenshotWorkflow(t *testing.T)
func TestYouTubeWorkflow(t *testing.T)
```

**Test Objectives:**
- Test complete workflow with mixed regular and YouTube URLs
- Verify markdown file annotation works for both screenshot types
- Confirm file naming and output directory handling
- Validate verbose logging shows correct processing path

**Validation Commands:**
- `go test ./...` - Run all tests including new compact screenshot tests
- `go build ./cmd/screenshot-tweets` - Verify application builds successfully
- `./screenshot-tweets -f testdata/sample-input.md -v --dry-run` - Test dry-run functionality
- `./screenshot-tweets -f tweets-context-engineering-week1.md -v` - Full integration test

**Context for implementation:**
- Use existing test patterns from `*_test.go` files
- Follow table-driven test style from project guidelines
- Use `require` for critical assertions, `assert` for value comparisons
- Test both success and error scenarios for YouTube thumbnail extraction

## Success Criteria

The implementation will be complete when:
1. Regular website screenshots are limited to 1200px height
2. YouTube URLs produce thumbnail images without browser automation
3. All existing tests pass with no regressions
4. New tests validate both screenshot modes work correctly
5. The tool processes the example input file correctly with compact output

This plan provides clear architectural guidance while preserving the existing tool's functionality and adding the requested compact screenshot behavior.