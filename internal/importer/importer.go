package importer

import (
	"strings"

	"github.com/pi/flash-cards/internal/models"
)

// delimiters is the ordered list of delimiters to try during auto-detection.
var delimiters = []string{"\t", ";", " - ", " = ", ","}

// DetectDelimiter analyzes content and returns the best delimiter for splitting
// lines into two-part flash card pairs. It tries each candidate delimiter and
// picks the one that produces the most lines with exactly two parts.
// If no delimiter succeeds on any line, it defaults to tab.
func DetectDelimiter(content string) string {
	lines := strings.Split(content, "\n")

	bestDelimiter := "\t"
	bestCount := 0

	for _, delim := range delimiters {
		count := 0
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			parts := strings.SplitN(line, delim, 3)
			if len(parts) == 2 {
				count++
			}
		}
		if count > bestCount {
			bestCount = count
			bestDelimiter = delim
		}
	}

	return bestDelimiter
}

// Parse splits the content into flash card pairs using auto-detected delimiters.
// Lines that are empty or do not split into exactly two non-empty parts are
// silently skipped.
func Parse(content string) ([]models.ImportCard, error) {
	delimiter := DetectDelimiter(content)
	lines := strings.Split(content, "\n")

	var cards []models.ImportCard
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, delimiter, 3)
		if len(parts) != 2 {
			continue
		}

		czech := strings.TrimSpace(parts[0])
		english := strings.TrimSpace(parts[1])
		if czech == "" || english == "" {
			continue
		}

		cards = append(cards, models.ImportCard{
			Czech:   czech,
			English: english,
		})
	}

	return cards, nil
}
