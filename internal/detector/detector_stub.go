//go:build !linux && !darwin && !windows

package detector

import "fmt"

// StubDetector is a placeholder for unsupported platforms
type StubDetector struct{}

// NewDetector creates a stub detector for unsupported platforms
func NewDetector() (Detector, error) {
	return nil, fmt.Errorf("song detection not implemented for this platform")
}

func (d *StubDetector) GetCurrentSong() (*SongInfo, error) {
	return nil, fmt.Errorf("not implemented")
}

func (d *StubDetector) Close() error {
	return nil
}
