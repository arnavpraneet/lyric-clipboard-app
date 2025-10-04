package lyrics

import (
	"bufio"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// LyricLine represents a single line of lyrics with its timestamp
type LyricLine struct {
	Time time.Duration
	Text string
}

// SyncedLyrics contains all lyric lines sorted by timestamp
type SyncedLyrics struct {
	Lines []LyricLine
}

// ParseLRC parses LRC format lyrics into structured data
// LRC format: [mm:ss.xx]Lyric text
func ParseLRC(lrcContent string) (*SyncedLyrics, error) {
	// Regex to match LRC timestamp format [mm:ss.xx] or [mm:ss]
	timeRegex := regexp.MustCompile(`\[(\d{2}):(\d{2})\.?(\d{2})?\]`)

	scanner := bufio.NewScanner(strings.NewReader(lrcContent))
	var lines []LyricLine

	for scanner.Scan() {
		line := scanner.Text()

		// Find all timestamp matches in the line
		matches := timeRegex.FindAllStringSubmatch(line, -1)
		if len(matches) == 0 {
			continue
		}

		// Extract the lyric text (everything after the last timestamp)
		text := timeRegex.ReplaceAllString(line, "")
		text = strings.TrimSpace(text)

		// Skip empty lyrics (metadata lines like [ar:Artist])
		if text == "" {
			continue
		}

		// Process each timestamp (some lines have multiple timestamps)
		for _, match := range matches {
			minutes, _ := strconv.Atoi(match[1])
			seconds, _ := strconv.Atoi(match[2])
			var centiseconds int
			if match[3] != "" {
				centiseconds, _ = strconv.Atoi(match[3])
			}

			timestamp := time.Duration(minutes)*time.Minute +
				time.Duration(seconds)*time.Second +
				time.Duration(centiseconds)*10*time.Millisecond

			lines = append(lines, LyricLine{
				Time: timestamp,
				Text: text,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading LRC content: %w", err)
	}

	if len(lines) == 0 {
		return nil, fmt.Errorf("no valid lyrics found in LRC content")
	}

	// Sort lines by timestamp
	sort.Slice(lines, func(i, j int) bool {
		return lines[i].Time < lines[j].Time
	})

	return &SyncedLyrics{Lines: lines}, nil
}

// GetLineAtTime returns the lyric line that should be displayed at the given time
func (sl *SyncedLyrics) GetLineAtTime(position time.Duration) *LyricLine {
	if len(sl.Lines) == 0 {
		return nil
	}

	// Find the last line whose timestamp is <= position
	var currentLine *LyricLine
	for i := range sl.Lines {
		if sl.Lines[i].Time <= position {
			currentLine = &sl.Lines[i]
		} else {
			break
		}
	}

	return currentLine
}
