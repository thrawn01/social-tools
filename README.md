# Screenshot Tweets

A command-line tool that automatically generates screenshots of article URLs and annotates markdown files with screenshot references for social media sharing.

## Features

- **Automated Screenshot Capture**: Uses headless browser automation to capture high-quality screenshots
- **Social Media Optimization**: Generates optimized images for Twitter/X (1200x628) and LinkedIn (1200x627)
- **Markdown Integration**: Parses markdown files in "## Day X" format and adds "Screen Shot: filename.png" annotations
- **Smart Processing**: Only processes entries that don't already have screenshots
- **Comprehensive Error Handling**: Robust error categorization with retry logic for transient failures
- **CLI Interface**: User-friendly command-line interface with verbose output and dry-run modes

## Installation

### Prerequisites

- Go 1.21 or later
- Chrome or Chromium browser (must be available in PATH)
- Write permissions to target directories

### Build from Source

```bash
git clone <repository-url>
cd screenshot-tweets
make build
```

### Install to GOPATH

```bash
make install
```

## Usage

### Basic Usage

```bash
./bin/screenshot-tweets --file path/to/your/tweets.md
```

### Command-Line Options

```bash
screenshot-tweets --help

A CLI tool that processes markdown files in "## Day X" format,
captures screenshots of URLs found in each day's entry, and adds
"Screen Shot: filename.png" annotations for social media sharing.

Usage:
  screenshot-tweets [flags]

Flags:
  -f, --file string      Path to the markdown file to process (required)
      --dry-run          Show what would be done without making changes
  -h, --help             help for screenshot-tweets
      --height int       Viewport height for screenshots (default 1200)
  -o, --output string    Directory to save screenshot files (default ".")
  -t, --timeout duration Timeout for screenshot capture (default 30s)
  -v, --verbose          Enable verbose logging
  -w, --width int        Viewport width for screenshots (default 800)
```

### Examples

**Process a markdown file:**
```bash
screenshot-tweets --file daily-tweets.md --verbose
```

**Dry run to see what would be processed:**
```bash
screenshot-tweets --file daily-tweets.md --dry-run
```

**Custom output directory and timeout:**
```bash
screenshot-tweets --file tweets.md --output ./screenshots --timeout 45s
```

**Custom viewport dimensions:**
```bash
screenshot-tweets --file tweets.md --width 1024 --height 768
```

### Viewport Optimization

Many websites are optimized for narrower viewports (around 800px), which results in larger, more readable text in screenshots. The default width of 800px provides good readability for most sites. Adjust the width based on your specific needs:

- 800px (default): Optimal for text-heavy articles and blogs
- 1024px: Good for technical documentation sites
- 1920px: Full desktop view for complex layouts

## Input Format

Your markdown file should follow this format:

```markdown
# Daily Tweet Log

## Day 1
Just discovered this amazing article about Go best practices!
- URL: https://go.dev/blog/go1.21

## Day 2
Working on a new project using Rod for browser automation.
- URL: https://pkg.go.dev/github.com/go-rod/rod

## Day 3
Entry without URL (will be skipped)

## Day 4
Another great resource about containerization.
- URL: https://docs.docker.com/develop/dev-best-practices/
```

## Output

After processing, your markdown file will be updated with screenshot references:

```markdown
## Day 1
Just discovered this amazing article about Go best practices!
- URL: https://go.dev/blog/go1.21
Screen Shot: day-1-screenshot.png

## Day 2
Working on a new project using Rod for browser automation.
- URL: https://pkg.go.dev/github.com/go-rod/rod
Screen Shot: day-2-screenshot.png
```

The tool will also generate social media optimized versions:
- `day-1-screenshot.png` (original)
- `day-1-screenshot-twitter.png` (1200x628)
- `day-1-screenshot-linkedin.png` (1200x627)

## Configuration

You can configure the tool using environment variables:

```bash
export SCREENSHOT_DEFAULT_TIMEOUT=45s
export SCREENSHOT_MAX_RETRIES=5
export SCREENSHOT_USER_AGENT="Custom-Agent/1.0"
export SCREENSHOT_BROWSER_PATH="/path/to/chrome"
```

## Development

### Setup Development Environment

```bash
make dev-setup
```

### Available Make Targets

```bash
make help                # Show available targets
make build              # Build the application
make test               # Run tests
make test-short         # Run tests in short mode (skip integration tests)
make test-coverage      # Run tests with coverage report
make lint               # Run linter
make fmt                # Format Go code
make vet                # Run go vet
make clean              # Clean build artifacts
make install            # Install binary to $GOPATH/bin
make run-example        # Run example with test data
make ci                 # Run all CI checks
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run only fast tests (skip browser automation tests)
make test-short
```

### Project Structure

```
screenshot-tweets/
├── cmd/screenshot-tweets/    # CLI entry point
│   └── main.go
├── markdown/                 # Markdown parsing
│   ├── parser.go
│   └── parser_test.go
├── screenshot/              # Screenshot capture and processing
│   ├── capture.go
│   ├── capture_test.go
│   ├── resize.go
│   └── resize_test.go
├── config/                  # Configuration management
│   ├── config.go
│   └── config_test.go
├── internal/errors/         # Custom error types
│   ├── errors.go
│   └── errors_test.go
├── testdata/               # Test fixtures
│   ├── sample-input.md
│   └── expected-output.md
├── Makefile                # Build automation
├── README.md
└── go.mod
```

## Error Handling

The tool categorizes errors and handles them appropriately:

- **Retryable Errors**: Timeouts, server errors, network issues
- **Permanent Errors**: 404 Not Found, DNS errors, forbidden access
- **Browser Errors**: Chrome launch failures, invalid configurations

Retryable errors will be attempted up to 3 times with exponential backoff.

## Troubleshooting

### Common Issues

**"failed to launch browser"**
- Ensure Chrome/Chromium is installed and available in PATH
- Try setting `SCREENSHOT_BROWSER_PATH` to specific browser location

**"screenshot capture timeout"**
- Increase timeout with `--timeout 60s`
- Check network connectivity to target URLs

**"permission denied"**
- Ensure write permissions to output directory
- Check file permissions on markdown file

**"no entries found that need screenshots"**
- Verify markdown format follows "## Day X" pattern
- Ensure URLs are formatted as "- URL: https://..."
- Check that entries don't already have "Screen Shot:" annotations

### Debug Mode

Run with verbose flag to see detailed processing information:

```bash
screenshot-tweets --file tweets.md --verbose
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Run tests and linting (`make ci`)
6. Commit your changes
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Code Style

- Follow standard Go formatting (`make fmt`)
- Add tests for new functionality
- Use meaningful variable names
- Document exported functions and types
- Keep functions focused and small

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Changelog

### v1.0.0
- Initial release
- Basic screenshot capture functionality
- Social media optimization
- Markdown parsing and annotation
- CLI interface with comprehensive options
- Error handling and retry logic
- Comprehensive test suite