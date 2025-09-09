package markdown_test

import (
	"os"
	"path/filepath"
	"testing"

	"screenshot-tweets/markdown"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMarkdownFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.md")

	content := `# Daily Tweet Log

## Day 1
Just discovered this amazing article about Go best practices!
- URL: https://go.dev/blog/go1.21

## Day 2
Working on a new project using Rod for browser automation.
- URL: https://pkg.go.dev/github.com/go-rod/rod
Screen Shot: day-2-screenshot.png

## Day 3
This is an entry without a URL to test the parser.

## Day 4
Another article with a URL.
- URL: https://example.com/article`

	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	mf, err := markdown.ParseMarkdownFile(testFile)
	require.NoError(t, err)
	require.NotNil(t, mf)

	assert.Equal(t, testFile, mf.FilePath)
	assert.Len(t, mf.Entries, 4)

	assert.Equal(t, 1, mf.Entries[0].Day)
	assert.Equal(t, "https://go.dev/blog/go1.21", mf.Entries[0].URL)
	assert.False(t, mf.Entries[0].HasScreenshot)
	assert.Empty(t, mf.Entries[0].Screenshot)

	assert.Equal(t, 2, mf.Entries[1].Day)
	assert.Equal(t, "https://pkg.go.dev/github.com/go-rod/rod", mf.Entries[1].URL)
	assert.True(t, mf.Entries[1].HasScreenshot)
	assert.Equal(t, "day-2-screenshot.png", mf.Entries[1].Screenshot)

	assert.Equal(t, 3, mf.Entries[2].Day)
	assert.Empty(t, mf.Entries[2].URL)
	assert.False(t, mf.Entries[2].HasScreenshot)

	assert.Equal(t, 4, mf.Entries[3].Day)
	assert.Equal(t, "https://example.com/article", mf.Entries[3].URL)
	assert.False(t, mf.Entries[3].HasScreenshot)
}

func TestParseMarkdownFileNonExistent(t *testing.T) {
	_, err := markdown.ParseMarkdownFile("/non/existent/file.md")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open file")
}

func TestUpdateScreenshotReference(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.md")

	content := `# Daily Tweet Log

## Day 1
Just discovered this amazing article!
- URL: https://go.dev/blog/go1.21

## Day 2
Another great article.
- URL: https://example.com/article`

	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	mf, err := markdown.ParseMarkdownFile(testFile)
	require.NoError(t, err)

	err = mf.UpdateScreenshotReference(1, "day-1-screenshot.png")
	require.NoError(t, err)

	found := false
	for _, entry := range mf.Entries {
		if entry.Day == 1 {
			assert.True(t, entry.HasScreenshot)
			assert.Equal(t, "day-1-screenshot.png", entry.Screenshot)
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestUpdateScreenshotReferenceAlreadyExists(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.md")

	content := `## Day 1
Article with existing screenshot.
- URL: https://example.com
Screen Shot: existing-screenshot.png`

	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	mf, err := markdown.ParseMarkdownFile(testFile)
	require.NoError(t, err)

	err = mf.UpdateScreenshotReference(1, "new-screenshot.png")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already has a screenshot reference")
}

func TestUpdateScreenshotReferenceNonExistentDay(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.md")

	content := `## Day 1
Some content.
- URL: https://example.com`

	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	mf, err := markdown.ParseMarkdownFile(testFile)
	require.NoError(t, err)

	err = mf.UpdateScreenshotReference(999, "screenshot.png")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "day 999 not found")
}

func TestWriteMarkdownFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.md")

	originalContent := `## Day 1
Test content.
- URL: https://example.com`

	err := os.WriteFile(testFile, []byte(originalContent), 0644)
	require.NoError(t, err)

	mf, err := markdown.ParseMarkdownFile(testFile)
	require.NoError(t, err)

	err = mf.UpdateScreenshotReference(1, "day-1-screenshot.png")
	require.NoError(t, err)

	err = mf.WriteMarkdownFile()
	require.NoError(t, err)

	updatedContent, err := os.ReadFile(testFile)
	require.NoError(t, err)

	assert.Contains(t, string(updatedContent), "Screen Shot: day-1-screenshot.png")
	assert.Contains(t, string(updatedContent), "## Day 1")
	assert.Contains(t, string(updatedContent), "- URL: https://example.com")
}

func TestGetEntriesWithoutScreenshots(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.md")

	content := `## Day 1
Entry with URL but no screenshot.
- URL: https://example.com/1

## Day 2
Entry with URL and screenshot.
- URL: https://example.com/2
Screen Shot: day-2-screenshot.png

## Day 3
Entry without URL.

## Day 4
Entry with URL but no screenshot.
- URL: https://example.com/4`

	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	mf, err := markdown.ParseMarkdownFile(testFile)
	require.NoError(t, err)

	entries := mf.GetEntriesWithoutScreenshots()
	assert.Len(t, entries, 2)

	assert.Equal(t, 1, entries[0].Day)
	assert.Equal(t, "https://example.com/1", entries[0].URL)

	assert.Equal(t, 4, entries[1].Day)
	assert.Equal(t, "https://example.com/4", entries[1].URL)
}

