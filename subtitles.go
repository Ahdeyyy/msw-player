package main

import (
	"encoding/json"
	"fmt"
	"github.com/egoist/mygo"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type mediaTrack struct {
	ID              int    `json:"id"`
	Type            string `json:"type"`
	Lang            string `json:"lang"`
	Title           string `json:"title"`
	External        bool   `json:"external"`
	Forced          bool   `json:"forced"`
	HearingImpaired bool   `json:"hearing-impaired"`
}

func trackID(data []byte) string {
	var id int
	if json.Unmarshal(data, &id) == nil {
		return strconv.Itoa(id)
	}
	var text string
	if json.Unmarshal(data, &text) == nil && text != "" {
		return text
	}
	return "no"
}
func subtitleTracks(raw string) []mediaTrack {
	var all []mediaTrack
	json.Unmarshal([]byte(raw), &all)
	var out []mediaTrack
	for _, t := range all {
		if t.Type == "sub" {
			out = append(out, t)
		}
	}
	return out
}
func (t mediaTrack) label() string {
	parts := []string{}
	if t.Title != "" {
		parts = append(parts, t.Title)
	}
	if t.Lang != "" {
		parts = append(parts, t.Lang)
	}
	if len(parts) == 0 {
		parts = append(parts, "Subtitle")
	}
	if t.Forced {
		parts = append(parts, "forced")
	}
	if t.HearingImpaired {
		parts = append(parts, "SDH")
	}
	if t.External {
		parts = append(parts, "external")
	}
	return fmt.Sprintf("%s (%d)", strings.Join(parts, " / "), t.ID)
}

var releaseSuffix = regexp.MustCompile(`(?i)\b(480p|720p|1080p|2160p|4k|web[ -]?dl|webrip|bluray|brrip|hdtv|dvdrip|x264|x265|h264|h265|hevc)\b.*$`)

func videoSearchTitle(path string) string {
	s := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	s = strings.NewReplacer(".", " ", "_", " ").Replace(s)
	s = releaseSuffix.ReplaceAllString(s, "")
	return strings.Join(strings.Fields(s), " ")
}
func (a *app) openSubtitleSearch() {
	if a.current < 0 {
		return
	}
	a.captionsOpen = false
	if a.subtitleQuery == "" {
		a.subtitleQuery = videoSearchTitle(a.queue[a.current])
	}
	a.onlineOpen = true
	a.layoutHost()
	if len(a.online.Matches) == 0 && len(a.online.Titles) == 0 && !a.online.Busy {
		a.findSubtitles(false, nil)
	}
}
func (a *app) chooseSubtitle(id string) {
	a.cancelSubtitleJob()
	a.online.AutoChecked = true
	a.online.PendingTitle = ""
	a.online.Message = ""
	a.online.Error = ""
	a.subtitleID = id
	value := any("no")
	if n, err := strconv.Atoi(id); err == nil {
		value = n
	}
	a.send("set_property", "sid", value)
	a.captionsOpen = false
	a.layoutHost()
}
func (a *app) loadSubtitle() {
	if a.window == nil || a.current < 0 {
		return
	}
	window := a.window
	file := a.queue[a.current]
	go func() {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{Parent: window, Title: "Load subtitle file", Filters: []mygo.FileFilter{{Name: "Subtitles", Extensions: []string{"srt", "ass", "ssa", "vtt", "sub", "idx"}}}})
		window.Update(func() {
			if a.current < 0 || a.queue[a.current] != file {
				return
			}
			if err != nil {
				a.notice = err.Error()
				return
			}
			if len(paths) > 0 {
				a.cancelSubtitleJob()
				a.online.AutoChecked = true
				a.online.PendingTitle = ""
				a.online.Message = "Local subtitle selected."
				a.online.Error = ""
				a.send("sub-add", paths[0], "select")
				a.captionsOpen = false
				a.onlineOpen = false
				a.layoutHost()
			}
		})
	}()
}

// Coordinates are physical pixels; layout dimensions are MyGo logical units.
func videoGeometry(width, height int, scale float64, fullscreen bool) [4]int {
	if fullscreen {
		return [4]int{0, 0, max(1, width), max(1, height)}
	}
	px := func(v float64) int { return int(v * scale) }
	return [4]int{px(20), px(100), max(1, width-px(40)), max(1, height-px(204))}
}
