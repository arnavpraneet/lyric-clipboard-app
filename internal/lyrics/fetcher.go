package lyrics

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Fetcher handles fetching and caching of song lyrics
type Fetcher struct {
	client *http.Client
	cache  map[string]*SyncedLyrics
	mu     sync.RWMutex
}

// NewFetcher creates a new lyrics fetcher with caching
func NewFetcher() *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		cache: make(map[string]*SyncedLyrics),
	}
}

// getCacheKey generates a cache key from artist and title
func (f *Fetcher) getCacheKey(artist, title string) string {
	return fmt.Sprintf("%s|||%s", artist, title)
}

// FetchLyrics fetches synced lyrics for a song
// Returns cached lyrics if available, otherwise fetches from source
func (f *Fetcher) FetchLyrics(artist, title string) (*SyncedLyrics, error) {
	cacheKey := f.getCacheKey(artist, title)

	// Check cache first
	f.mu.RLock()
	if lyrics, exists := f.cache[cacheKey]; exists {
		f.mu.RUnlock()
		return lyrics, nil
	}
	f.mu.RUnlock()

	// Fetch lyrics from source
	lyrics, err := f.fetchFromSource(artist, title)
	if err != nil {
		return nil, err
	}

	// Cache the result
	f.mu.Lock()
	f.cache[cacheKey] = lyrics
	f.mu.Unlock()

	return lyrics, nil
}

// fetchFromSource fetches lyrics from an external source
// Currently uses lrclib.net API as the primary source
func (f *Fetcher) fetchFromSource(artist, title string) (*SyncedLyrics, error) {
	// Try lrclib.net API
	lrcContent, err := f.fetchFromLRCLib(artist, title)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch lyrics: %w", err)
	}

	// Parse the LRC content
	lyrics, err := ParseLRC(lrcContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse lyrics: %w", err)
	}

	return lyrics, nil
}

// fetchFromLRCLib fetches lyrics from lrclib.net
func (f *Fetcher) fetchFromLRCLib(artist, title string) (string, error) {
	baseURL := "https://lrclib.net/api/get"

	// Build query parameters
	params := url.Values{}
	params.Add("artist_name", artist)
	params.Add("track_name", title)

	requestURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := f.client.Get(requestURL)
	if err != nil {
		return "", fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// lrclib.net returns JSON, we need to extract the syncedLyrics field
	// For simplicity, we'll do basic string parsing
	// In production, use encoding/json
	bodyStr := string(body)

	// Simple extraction of syncedLyrics field
	// Expected format: {"syncedLyrics":"[00:12.34]Line 1\n[00:15.67]Line 2",...}
	start := findJSONField(bodyStr, "syncedLyrics")
	if start == -1 {
		return "", fmt.Errorf("syncedLyrics field not found in response")
	}

	lrcContent := extractJSONString(bodyStr[start:])
	if lrcContent == "" {
		return "", fmt.Errorf("no synced lyrics available for this song")
	}

	return lrcContent, nil
}

// findJSONField finds the start position of a JSON field value
func findJSONField(json, field string) int {
	needle := fmt.Sprintf(`"%s":"`, field)
	pos := indexOf(json, needle)
	if pos == -1 {
		return -1
	}
	return pos + len(needle)
}

// extractJSONString extracts a JSON string value (handles basic escaping)
func extractJSONString(json string) string {
	var result strings.Builder
	escaped := false

	for i := 0; i < len(json); i++ {
		ch := json[i]

		if escaped {
			// Handle escape sequences
			switch ch {
			case 'n':
				result.WriteByte('\n')
			case 't':
				result.WriteByte('\t')
			case 'r':
				result.WriteByte('\r')
			case '"', '\\':
				result.WriteByte(ch)
			default:
				result.WriteByte(ch)
			}
			escaped = false
		} else if ch == '\\' {
			escaped = true
		} else if ch == '"' {
			// End of string
			break
		} else {
			result.WriteByte(ch)
		}
	}

	return result.String()
}

// indexOf finds the first occurrence of substr in s
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// ClearCache clears the lyrics cache
func (f *Fetcher) ClearCache() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.cache = make(map[string]*SyncedLyrics)
}
