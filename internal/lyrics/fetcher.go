package lyrics

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

// LRCLibResponse represents the JSON response from lrclib.net API
type LRCLibResponse struct {
	SyncedLyrics   *string `json:"syncedLyrics"`
	PlainLyrics    *string `json:"plainLyrics"`
	TrackName      string  `json:"trackName"`
	ArtistName     string  `json:"artistName"`
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
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Parse JSON response
	var lrcResponse LRCLibResponse
	if err := json.Unmarshal(body, &lrcResponse); err != nil {
		return "", fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Check if syncedLyrics is available
	if lrcResponse.SyncedLyrics == nil || *lrcResponse.SyncedLyrics == "" {
		return "", fmt.Errorf("no synced lyrics available for this song")
	}

	return *lrcResponse.SyncedLyrics, nil
}

// ClearCache clears the lyrics cache
func (f *Fetcher) ClearCache() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.cache = make(map[string]*SyncedLyrics)
}
