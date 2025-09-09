# Screenshot Automation Tool Implementation Plan

## Overview

We're building a new Go project called `screenshot-tweets` that will automatically generate screenshots of article URLs and annotate markdown files with the screenshot references. The tool processes markdown files in `## Day X` format and adds `Screen Shot: filename.png` annotations for social media sharing.

## Project Structure

```
screenshot-tweets/
├── go.mod
├── go.sum
├── cmd/
│   └── screenshot-tweets/
│       └── main.go         # CLI entry point
├── markdown/
│   ├── parser.go           # Markdown parsing and processing
│   └── parser_test.go
├── screenshot/
│   ├── capture.go          # Screenshot capture with Rod
│   ├── capture_test.go
│   ├── resize.go           # Social media optimization
│   └── resize_test.go
├── config/
│   └── config.go           # Configuration and CLI flags
├── internal/
│   └── errors/
│       └── errors.go       # Custom error types
├── testdata/
│   ├── sample-input.md     # Test markdown files
│   └── expected-output.md
├── README.md
├── Makefile
└── .gitignore
```

## Current State Analysis

**New Project**: Starting from scratch with no existing codebase
- Clean slate for implementing best practices
- Can follow modern Go project layout standards
- No legacy code or dependencies to consider

### Key Requirements:
- Social media optimization: 1200x628 (Twitter/X) and 1200x627 (LinkedIn)
- Process existing Day-format markdown files
- Generate screenshots only for entries missing them
- Error handling with markdown annotation for retries

## Desired End State

After implementation completion:

1. **Standalone Go Application**
   ```bash
   ./screenshot-tweets -file path/to/tweets.md
   ```

2. **Input Markdown Format** (already exists):
   ```markdown
   ## Day 1
   [Tweet content]
   - URL: https://example.com/article
   ```

3. **Output Markdown Format** (annotated):
   ```markdown
   ## Day 1
   [Tweet content]
   - URL: https://example.com/article
   Screen Shot: day-1-screenshot.png
   ```

4. **Screenshot Files**: Generated in same directory as markdown file
5. **Error Handling**: Failed attempts annotated in markdown for retry

**Verification**: Run the tool on a test markdown file and confirm all URLs without screenshots get processed and annotated correctly.

## What We're NOT Doing

- Not building a web interface or GUI
- Not implementing batch processing of multiple markdown files simultaneously
- Not integrating with social media APIs for posting
- Not implementing real-time monitoring or scheduling
- Not creating a configuration file system (CLI flags only)

## Dependencies and Environment Setup

### Required Dependencies
```go
module screenshot-tweets

go 1.21

require (
    github.com/go-rod/rod v0.116.0
    github.com/disintegration/imaging v1.6.2
    github.com/spf13/cobra v1.8.0
)
```

### Browser Configuration
- **Browser Requirements**: Chrome/Chromium browser (auto-detected from system PATH)
- **Headless Mode**: Enabled by default for automation
- **Viewport**: 1920x1080 for capture, then resize for social media
- **User-Agent**: Standard desktop Chrome to avoid bot detection

### Development Setup
- Go 1.21+ installation
- Chrome/Chromium browser available in PATH
- Write permissions to target directories
- Git for version control

## Phase 1: Project Foundation and Basic Screenshot Automation

### Overview
Establish project structure and implement core screenshot functionality with Rod.

### Changes Required:

#### 1. Project Initialization
**Files**: Root directory structure
**Changes**: Create new Go project with standard layout

```bash
# Project initialization commands
go mod init screenshot-tweets
mkdir -p cmd/screenshot-tweets markdown screenshot config internal/errors testdata
touch cmd/screenshot-tweets/main.go markdown/parser.go screenshot/capture.go
```

#### 2. Core Data Structures
**File**: `markdown/parser.go`
**Changes**: Define core data structures

```go
// DayEntry represents a single day's tweet with URL and screenshot info
type DayEntry struct {
    Day        int    `json:"day"`
    Content    string `json:"content"`
    URL        string `json:"url"`
    Screenshot string `json:"screenshot"`
    HasScreenshot bool `json:"has_screenshot"`
    Error      string `json:"error,omitempty"`
}

// MarkdownFile represents the entire markdown file structure
type MarkdownFile struct {
    FilePath string     `json:"file_path"`
    Entries  []DayEntry `json:"entries"`
}

// ParseMarkdownFile reads and parses markdown file in Day format
func ParseMarkdownFile(filePath string) (*MarkdownFile, error)

// UpdateScreenshotReference adds screenshot annotation to markdown file
func (mf *MarkdownFile) UpdateScreenshotReference(day int, filename string) error

// WriteMarkdownFile saves the updated markdown file
func (mf *MarkdownFile) WriteMarkdownFile() error
```

**Function Responsibilities:**
- Parse markdown file using regex patterns for `## Day X` sections
- Extract URLs matching pattern: `^- URL: (https?://.+)$`
- Detect existing `Screen Shot:` annotations to avoid duplicates
- Preserve original markdown formatting and content
- Handle edge cases like missing URLs or malformed Day sections

#### 3. Screenshot Capture Module
**File**: `screenshot/capture.go`
**Changes**: Implement Rod-based screenshot automation

```go
// ScreenshotConfig defines capture settings
type ScreenshotConfig struct {
    ViewportWidth  int           `json:"viewport_width"`
    ViewportHeight int           `json:"viewport_height"`
    Timeout        time.Duration `json:"timeout"`
    OutputDir      string        `json:"output_dir"`
    UserAgent      string        `json:"user_agent"`
}

// CaptureScreenshot takes a screenshot of the given URL
func CaptureScreenshot(url, filename string, config ScreenshotConfig) error

// WaitForPageLoad implements comprehensive page loading detection
func WaitForPageLoad(page *rod.Page) error

// GenerateFilename creates screenshot filename from day number
func GenerateFilename(day int, outputDir string) string
```

**Function Responsibilities:**
- Initialize Rod browser with headless configuration
- Navigate to URL and wait for complete page loading using:
  - `page.WaitLoad()` for DOM ready
  - `page.Eval('() => document.fonts.ready')` for font loading  
  - `page.WaitRequestIdle()` for network idle
  - `page.WaitStable()` for content stability
- Capture screenshot at 1920x1080 viewport
- Save PNG files with consistent naming: `day-X-screenshot.png`
- Handle errors with detailed error messages

#### 4. CLI Entry Point
**File**: `cmd/screenshot-tweets/main.go`
**Changes**: Command-line interface using Cobra

```go
// CLI flags and configuration
var (
    markdownFile string
    outputDir    string
    verbose      bool
    dryRun       bool
    timeout      time.Duration
)

// main function with cobra CLI setup
func main()

// runScreenshotAutomation executes the main logic
func runScreenshotAutomation(cmd *cobra.Command, args []string) error
```

**Function Responsibilities:**
- Parse CLI flags for file path, output directory, verbosity
- Validate input file exists and is readable
- Coordinate between markdown parsing and screenshot capture
- Provide progress feedback and error reporting
- Support dry-run mode for testing

### Testing Requirements:

```go
// markdown/parser_test.go
func TestParseMarkdownFile(t *testing.T)
func TestUpdateScreenshotReference(t *testing.T)
func TestWriteMarkdownFile(t *testing.T)

// screenshot/capture_test.go  
func TestCaptureScreenshot(t *testing.T)
func TestWaitForPageLoad(t *testing.T)
func TestGenerateFilename(t *testing.T)
```

**Test Objectives:**
- Verify markdown parsing handles various Day formats correctly
- Test screenshot capture with different website types
- Validate error handling and file I/O operations
- Confirm CLI flag parsing and help text display
- Test with sample markdown files in `testdata/`

**Context for Implementation:**
- Follow Go project layout standards from `golang-standards/project-layout`
- Use Rod's comprehensive page load detection for reliability
- Implement proper error handling with custom error types
- Add structured logging for debugging and verbose output
- Ensure proper resource cleanup with defer statements

### Validation Commands
```bash
go mod tidy
go build -o screenshot-tweets ./cmd/screenshot-tweets
go test ./...
./screenshot-tweets --help
./screenshot-tweets --dry-run --file testdata/sample-input.md
```

## Phase 2: Social Media Optimization and Image Processing

### Overview
Add image resizing, optimization for Twitter/LinkedIn, and enhanced error handling.

### Changes Required:

#### 1. Social Media Image Processing
**File**: `screenshot/resize.go`
**Changes**: Add social media optimization capabilities

```go
// SocialMediaPlatform defines platform-specific requirements
type SocialMediaPlatform struct {
    Name   string `json:"name"`
    Width  int    `json:"width"`
    Height int    `json:"height"`
}

// PlatformConfigs defines standard social media dimensions
var PlatformConfigs = map[string]SocialMediaPlatform{
    "twitter":  {Name: "Twitter/X", Width: 1200, Height: 628},
    "linkedin": {Name: "LinkedIn", Width: 1200, Height: 627},
}

// ResizeForSocialMedia creates optimized versions for social platforms
func ResizeForSocialMedia(originalFile, baseFilename string) error

// SmartCrop implements content-aware cropping
func SmartCrop(img image.Image, targetWidth, targetHeight int) image.Image
```

**Function Responsibilities:**
- Generate multiple format versions (original + Twitter + LinkedIn)
- Implement smart cropping algorithm:
  - Priority 1: Preserve top 628px of content (ensures headline visibility)
  - Priority 2: Detect article headlines (h1 tags) and ensure inclusion
  - Fallback: Center crop for optimal content balance
- Optimize PNG compression while maintaining text readability
- Create filename variants: `day-1-screenshot-twitter.png`, `day-1-screenshot-linkedin.png`

#### 2. Enhanced Error Handling
**File**: `internal/errors/errors.go`
**Changes**: Custom error types for better error handling

```go
// ScreenshotError represents screenshot-specific errors
type ScreenshotError struct {
    URL       string
    Day       int
    ErrorType string
    Message   string
    Timestamp time.Time
}

// Error implements the error interface
func (e ScreenshotError) Error() string

// IsRetryableError determines if error should trigger retry
func IsRetryableError(err error) bool
```

**Function Responsibilities:**
- Categorize errors (network timeout, 404, page load failure)
- Annotate markdown with detailed error information
- Distinguish between permanent and temporary failures
- Support retry logic with exponential backoff

### Testing Requirements:
```go
func TestResizeForSocialMedia(t *testing.T)
func TestSmartCrop(t *testing.T)
func TestScreenshotError(t *testing.T)
```

**Test Objectives:**
- Verify image resizing maintains aspect ratios and quality
- Test smart cropping with various article layouts
- Validate error categorization and retry logic
- Confirm multiple format generation produces correct dimensions

### Validation Commands
```bash
go test ./... -v
./screenshot-tweets --file testdata/sample-input.md --verbose
# Manual verification of generated image dimensions
identify day-1-screenshot*.png  # ImageMagick command to check dimensions
```

## Phase 3: Production Readiness and Documentation

### Overview
Add comprehensive documentation, build automation, and production-ready features.

### Changes Required:

#### 1. Build and Documentation
**File**: `Makefile`
**Changes**: Add build automation and development tools

```makefile
.PHONY: build test clean install lint

build:
	go build -o bin/screenshot-tweets ./cmd/screenshot-tweets

test:
	go test ./... -v

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

lint:
	golangci-lint run

install:
	go install ./...

clean:
	rm -rf bin/ coverage.out
```

#### 2. Documentation
**File**: `README.md`
**Changes**: Comprehensive user and developer documentation

**Content Requirements:**
- Installation instructions
- Usage examples with sample markdown files
- CLI flag documentation
- Troubleshooting guide
- Contributing guidelines

#### 3. Configuration Enhancements
**File**: `config/config.go`
**Changes**: Advanced configuration options

```go
// Config represents application configuration
type Config struct {
    BrowserPath    string        `json:"browser_path"`
    DefaultTimeout time.Duration `json:"default_timeout"`
    MaxRetries     int          `json:"max_retries"`
    UserAgent      string        `json:"user_agent"`
    OutputFormats  []string      `json:"output_formats"`
}

// LoadConfig loads configuration from environment and flags
func LoadConfig() (*Config, error)
```

**Function Responsibilities:**
- Support environment variable configuration
- Validate configuration values
- Provide sensible defaults
- Support custom browser binary paths

### Testing Requirements:
```go
func TestMakefileTargets(t *testing.T)    // Integration test
func TestLoadConfig(t *testing.T)
func TestDocumentationExamples(t *testing.T)
```

**Test Objectives:**
- Validate all Makefile targets execute successfully
- Test configuration loading from various sources
- Verify README examples work as documented
- Confirm installation and packaging process

### Validation Commands
```bash
make build
make test
make test-coverage
make lint
make install
screenshot-tweets --version
```

## Final Deliverable

A complete Go project that provides:
- **Reliable screenshot automation** using Rod browser automation
- **Social media optimization** with platform-specific image formats
- **Robust error handling** with retry capabilities and markdown annotation
- **Professional CLI interface** with comprehensive help and options
- **Complete test coverage** with integration and unit tests
- **Documentation** for users and contributors
- **Build automation** for development and distribution

The tool will be ready for production use with proper error handling, logging, and the ability to process markdown files efficiently while generating high-quality screenshots optimized for social media sharing.