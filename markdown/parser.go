package markdown

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type DayEntry struct {
	Day           int    `json:"day"`
	Content       string `json:"content"`
	URL           string `json:"url"`
	Screenshot    string `json:"screenshot"`
	HasScreenshot bool   `json:"has_screenshot"`
	Error         string `json:"error,omitempty"`
}

type MarkdownFile struct {
	FilePath string     `json:"file_path"`
	Entries  []DayEntry `json:"entries"`
	content  []string
}

const screenshotPrefix = "Screen Shot: "

var (
	dayHeaderRegex  = regexp.MustCompile(`^## Day (\d+)`)
	urlRegex        = regexp.MustCompile(`^- URL: (https?://.+)$`)
	screenshotRegex = regexp.MustCompile(`^Screen Shot: (.+)$`)
)

func ParseMarkdownFile(filePath string) (*MarkdownFile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	mf := &MarkdownFile{
		FilePath: filePath,
		Entries:  []DayEntry{},
		content:  []string{},
	}

	scanner := bufio.NewScanner(file)
	var currentEntry *DayEntry
	var contentBuffer []string

	for scanner.Scan() {
		line := scanner.Text()
		mf.content = append(mf.content, line)

		if matches := dayHeaderRegex.FindStringSubmatch(line); matches != nil {
			if currentEntry != nil {
				currentEntry.Content = strings.Join(contentBuffer, "\n")
				mf.Entries = append(mf.Entries, *currentEntry)
			}

			day, err := strconv.Atoi(matches[1])
			if err != nil {
				return nil, fmt.Errorf("invalid day number in header %q: %w", line, err)
			}
			currentEntry = &DayEntry{
				Day:           day,
				HasScreenshot: false,
			}
			contentBuffer = []string{line}
			continue
		}

		if currentEntry != nil {
			contentBuffer = append(contentBuffer, line)

			if matches := urlRegex.FindStringSubmatch(strings.TrimSpace(line)); matches != nil {
				currentEntry.URL = matches[1]
			}

			if matches := screenshotRegex.FindStringSubmatch(strings.TrimSpace(line)); matches != nil {
				currentEntry.Screenshot = matches[1]
				currentEntry.HasScreenshot = true
			}
		}
	}

	if currentEntry != nil {
		currentEntry.Content = strings.Join(contentBuffer, "\n")
		mf.Entries = append(mf.Entries, *currentEntry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return mf, nil
}

func (mf *MarkdownFile) UpdateScreenshotReference(day int, filename string) error {
	for i, entry := range mf.Entries {
		if entry.Day == day {
			if entry.HasScreenshot {
				return fmt.Errorf("day %d already has a screenshot reference", day)
			}

			mf.Entries[i].Screenshot = filename
			mf.Entries[i].HasScreenshot = true

			dayHeaderPattern := fmt.Sprintf("## Day %d", day)
			var insertIndex = -1

			for j, line := range mf.content {
				if strings.Contains(line, dayHeaderPattern) {
					for k := j + 1; k < len(mf.content); k++ {
						if strings.HasPrefix(mf.content[k], "## Day") {
							insertIndex = k - 1
							break
						}
					}
					if insertIndex == -1 {
						insertIndex = len(mf.content) - 1
					}
					break
				}
			}

			if insertIndex == -1 {
				return fmt.Errorf("day %d not found in markdown content", day)
			}

			screenshotLine := screenshotPrefix + filename
			newContent := make([]string, len(mf.content)+1)
			copy(newContent[:insertIndex+1], mf.content[:insertIndex+1])
			newContent[insertIndex+1] = screenshotLine
			copy(newContent[insertIndex+2:], mf.content[insertIndex+1:])
			mf.content = newContent

			return nil
		}
	}

	return fmt.Errorf("day %d not found", day)
}

func (mf *MarkdownFile) WriteMarkdownFile() error {
	backupPath := mf.FilePath + ".backup"

	originalData, err := os.ReadFile(mf.FilePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read original file for backup: %w", err)
	}

	if len(originalData) > 0 {
		if err := os.WriteFile(backupPath, originalData, 0644); err != nil {
			return fmt.Errorf("failed to create backup file: %w", err)
		}
	}

	file, err := os.Create(mf.FilePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", mf.FilePath, err)
	}

	writer := bufio.NewWriter(file)
	for _, line := range mf.content {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			file.Close()
			if len(originalData) > 0 {
				os.WriteFile(mf.FilePath, originalData, 0644)
				os.Remove(backupPath)
			}
			return fmt.Errorf("error writing to file: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		file.Close()
		if len(originalData) > 0 {
			os.WriteFile(mf.FilePath, originalData, 0644)
			os.Remove(backupPath)
		}
		return fmt.Errorf("error flushing to file: %w", err)
	}

	file.Close()

	if len(originalData) > 0 {
		os.Remove(backupPath)
	}

	return nil
}

func (mf *MarkdownFile) GetEntriesWithoutScreenshots() []DayEntry {
	var entries []DayEntry
	for _, entry := range mf.Entries {
		if !entry.HasScreenshot && entry.URL != "" {
			entries = append(entries, entry)
		}
	}
	return entries
}

