# Social Tools

### Build from Source

```bash
git clone <repository-url>
cd social-tools
make install
```
### ScreenShot for Tweets

```bash
./bin/screenshot-tweets --file path/to/your/tweets.md
```

**Custom viewport dimensions:**
```bash
screenshot-tweets --file tweets.md --width 1024 --height 768
```

### Viewport Optimization

Many websites are optimized for narrower viewports (around 800px), which results in larger, more readable text in screenshots. The default dimensions of 800x600 provide good readability while capturing the essential above-the-fold content. Adjust the dimensions based on your specific needs:

- 800x600 (default): Optimal for text-heavy articles and blogs, captures key content without excessive scrolling
- 1024x768: Good for technical documentation sites with more complex layouts
- 1920x1200: Full desktop view for complex layouts and dashboards

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
