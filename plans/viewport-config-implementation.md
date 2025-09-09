# Viewport Configuration Implementation Plan

## Overview

Add configurable viewport dimensions (width and height) via CLI flags to the screenshot-tweets tool, allowing users to optimize screenshots for different website layouts. This replaces the current hardcoded viewport and environment variable configuration with explicit CLI flags, defaulting to 800px width for better text readability.

## Current State Analysis

The tool currently has viewport dimensions hardcoded at 1920x1200 in `cmd/screenshot-tweets/main.go:81`, with environment variable fallbacks (`SCREENSHOT_VIEWPORT_WIDTH/HEIGHT`) defined in `config/config.go:24-25` but not actually used by the main command. The tool already correctly handles entries without URLs by filtering them in `GetEntriesWithoutScreenshots()`.

## Desired End State

Users can specify custom viewport dimensions via `--width|-w` and `--height|-h` CLI flags with sensible defaults (800x1200). The tool gracefully handles markdown entries without URLs (already working). Documentation is updated to reflect the new flags and behavior.

### Key Discoveries:
- Viewport config exists in two places: `config/config.go` and `screenshot/capture.go` 
- The main command doesn't use the config package's viewport settings (`cmd/screenshot-tweets/main.go:79-85`)
- Tests already validate entries without URLs are handled correctly (`markdown/parser_test.go:173-207`)
- Cobra flag patterns are well-established (`cmd/screenshot-tweets/main.go:36-42`)

## What We're NOT Doing

- Not adding viewport presets (e.g., --preset=mobile)
- Not supporting responsive viewport changes during capture
- Not adding aspect ratio constraints
- Not modifying the social media resize functionality

## Implementation Approach

Break the implementation into three phases: CLI integration, configuration cleanup, and documentation updates. Each phase delivers working functionality that can be tested independently.

## Phase 1: Add CLI Flags for Viewport Configuration

### Overview
Add `--width|-w` and `--height|-h` flags to the CLI interface and integrate them with the screenshot capture configuration.

### Changes Required:

#### 1. CLI Flag Definition
**File**: `cmd/screenshot-tweets/main.go`
**Changes**: Add viewport dimension flags and wire them to screenshot config

```go
// Add to var declarations (line ~19)
var (
    viewportWidth  int
    viewportHeight int
)

// Add to init() function (line ~40)
func init() {
    rootCmd.Flags().IntVarP(&viewportWidth, "width", "w", 800, "Viewport width for screenshots")
    rootCmd.Flags().IntVarP(&viewportHeight, "height", "h", 1200, "Viewport height for screenshots")
}

// Update runScreenshotAutomation function (line ~79)
config := screenshot.ScreenshotConfig{
    ViewportWidth:  viewportWidth,  // Use CLI flag value
    ViewportHeight: viewportHeight, // Use CLI flag value
}
```

**Function Responsibilities:**
- Parse and validate viewport dimensions from CLI input
- Pass validated dimensions to screenshot configuration
- Maintain backward compatibility with existing flags

**Testing Requirements:**
```go
func TestCLIFlagsViewportDimensions(t *testing.T)
func TestViewportDefaultValues(t *testing.T)
```

**Test Objectives:**
- Verify default values are 800x1200
- Test custom viewport dimensions are respected
- Validate flag parsing with short and long forms

**Context for implementation:**
- Follow existing flag patterns at `cmd/screenshot-tweets/main.go:36-42`
- Use `IntVarP` for integer flags with short aliases
- Default values: width=800, height=1200

## Phase 2: Remove Environment Variable Configuration

### Overview
Clean up the configuration system by removing unused environment variable support for viewport dimensions.

### Changes Required:

#### 1. Configuration Cleanup
**File**: `config/config.go`
**Changes**: Remove viewport-related environment variables

```go
// LoadConfig function - remove these lines (lines 24-25)
func LoadConfig() (*Config, error) {
    // Remove: ViewportWidth:  getIntFromEnv("SCREENSHOT_VIEWPORT_WIDTH", 1920),
    // Remove: ViewportHeight: getIntFromEnv("SCREENSHOT_VIEWPORT_HEIGHT", 1080),
}

// Config struct - remove viewport fields (lines 16-17)
type Config struct {
    // Remove: ViewportWidth  int
    // Remove: ViewportHeight int
}

// Update Validate() to remove viewport validation (lines 47-49)
func (c *Config) Validate() error {
    // Remove viewport dimension validation
}

// Update DefaultConfig() (lines 95-96)
func DefaultConfig() *Config {
    // Remove ViewportWidth and ViewportHeight fields
}
```

**Function Responsibilities:**
- Maintain other environment variable configurations
- Keep validation for remaining config fields
- Preserve backward compatibility for non-viewport settings

**Testing Requirements:**
```go
func TestConfigWithoutViewport(t *testing.T)
func TestConfigValidationWithoutViewport(t *testing.T)
```

**Test Objectives:**
- Verify config loads without viewport fields
- Test validation still works for other fields
- Ensure environment variable tests pass without viewport

**Context for implementation:**
- Environment variable tests at `config/config_test.go:39-62`
- Validation patterns at `config/config_test.go:109-204`
- Keep other environment variables intact

#### 2. Update Screenshot Config Defaults
**File**: `screenshot/capture.go`
**Changes**: Update default viewport dimensions

```go
// NewDefaultConfig function (line 40)
func NewDefaultConfig() ScreenshotConfig {
    return ScreenshotConfig{
        ViewportWidth:  800,  // Changed from 1920
        ViewportHeight: 1200, // Keep at 1200
    }
}
```

**Function Responsibilities:**
- Provide sensible defaults when config not specified
- Maintain consistency with CLI defaults

**Testing Requirements:**
```go
func TestNewDefaultConfigViewport(t *testing.T)
```

**Test Objectives:**
- Verify default width is 800
- Verify default height is 1200

**Context for implementation:**
- Current defaults at `screenshot/capture.go:42-43`

## Phase 3: Update Documentation

### Overview
Update README and help text to document the new viewport configuration flags.

### Changes Required:

#### 1. README Updates
**File**: `README.md`
**Changes**: Add viewport flag documentation

```markdown
# In "Command-Line Options" section (line ~47)
Flags:
  -w, --width int        Viewport width for screenshots (default 800)
  -h, --height int       Viewport height for screenshots (default 1200)

# In "Examples" section (line ~69)
**Custom viewport dimensions:**
```bash
screenshot-tweets --file tweets.md --width 1024 --height 768
```

# In "Configuration" section - REMOVE (lines 133-134)
# Remove: export SCREENSHOT_VIEWPORT_WIDTH=1920
# Remove: export SCREENSHOT_VIEWPORT_HEIGHT=1080

# Add note about viewport optimization (after line ~75)
### Viewport Optimization
Many websites are optimized for narrower viewports (around 800px), which results in larger, more readable text in screenshots. The default width of 800px provides good readability for most sites. Adjust the width based on your specific needs:
- 800px (default): Optimal for text-heavy articles and blogs
- 1024px: Good for technical documentation sites  
- 1920px: Full desktop view for complex layouts
```

**Function Responsibilities:**
- Document new CLI flags with examples
- Explain viewport optimization rationale
- Remove obsolete environment variable documentation

**Testing Requirements:**
```go
func TestDocumentationCompleteness(t *testing.T)
```

**Test Objectives:**
- Verify all CLI flags are documented
- Check examples work as documented
- Ensure no references to removed env vars

**Context for implementation:**
- Current README structure at `README.md:44-63`
- Configuration section at `README.md:126-137`
- Examples section at `README.md:65-82`

#### 2. CLI Help Text Updates
**File**: `cmd/screenshot-tweets/main.go`
**Changes**: Already handled by cobra flag definitions in Phase 1

**Function Responsibilities:**
- Automatic help generation via cobra
- Flag descriptions in init() function

**Testing Requirements:**
```go
func TestHelpTextIncludesViewportFlags(t *testing.T)
```

**Test Objectives:**
- Verify --help shows viewport flags
- Check flag descriptions are clear
- Test short form aliases work

**Context for implementation:**
- Cobra automatically generates help from flag definitions
- Pattern established at `cmd/screenshot-tweets/main.go:26-32`

## Validation Commands

After each phase, run these commands to validate the implementation:

```bash
# Phase 1: Test CLI flag parsing
make build
./bin/screenshot-tweets --help | grep -E "width|height"
./bin/screenshot-tweets --file testdata/sample-input.md --dry-run --width 1024 --height 768 --verbose

# Phase 2: Test configuration changes  
make test
make lint
make vet

# Phase 3: Documentation validation
grep -r "SCREENSHOT_VIEWPORT" . # Should return no results
./bin/screenshot-tweets --help # Verify help text

# Final validation
make ci # Run all CI checks
```

## Notes on Existing URL Handling

The tool already correctly handles entries without URLs:
- `GetEntriesWithoutScreenshots()` filters for entries with non-empty URLs (`markdown/parser.go:198-199`)
- Tests confirm this behavior (`markdown/parser_test.go:173-207`)
- No changes needed for URL handling requirement