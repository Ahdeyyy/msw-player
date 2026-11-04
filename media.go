package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

var videoExtensions = []string{"mkv", "mp4", "mov", "webm", "avi", "m4v", "mpeg", "mpg", "ts", "m2ts", "wmv", "flv", "ogv"}

func supported(path string) bool {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	for _, v := range videoExtensions {
		if ext == v {
			return true
		}
	}
	return false
}
func clock(seconds float64) string {
	n := max(0, int(seconds))
	if n >= 3600 {
		return fmt.Sprintf("%d:%02d:%02d", n/3600, n/60%60, n%60)
	}
	return fmt.Sprintf("%d:%02d", n/60, n%60)
}

type playback struct {
	Generation            uint64
	Position, Duration    float64
	Paused, Ended, Loaded bool
	FrameReady            bool
	Width, Height         int
	Codec, Decoder        string
	Error                 string
	TrackJSON, SubtitleID string
	FPS                   float64
	AudioID, ChapterJSON  string
	Chapter               int
	InputSerial           uint64
}
