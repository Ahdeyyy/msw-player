package main

import (
	"encoding/json"
	"fmt"
	"github.com/egoist/mygo/ui"
	"strconv"
	"time"
)

type videoChapter struct {
	Title string  `json:"title"`
	Time  float64 `json:"time"`
}

func audioTracks(raw string) []mediaTrack {
	var all []mediaTrack
	json.Unmarshal([]byte(raw), &all)
	out := []mediaTrack{}
	for _, track := range all {
		if track.Type == "audio" {
			out = append(out, track)
		}
	}
	return out
}
func chapters(raw string) []videoChapter {
	var list []videoChapter
	json.Unmarshal([]byte(raw), &list)
	return list
}
func (c videoChapter) label(index int) string {
	title := c.Title
	if title == "" {
		title = fmt.Sprintf("Chapter %d", index+1)
	}
	return clock(c.Time) + "  " + title
}
func (a *app) selectAudio(id string) {
	value := any("no")
	if n, err := strconv.Atoi(id); err == nil {
		value = n
	}
	a.send("set_property", "aid", value)
	a.status.AudioID = id
	a.pendingAudio.set(id, time.Now())
	a.audioOpen = false
	a.layoutHost()
}
func (a *app) selectChapter(index int, close bool) {
	if !a.status.Loaded || a.transition != nil {
		return
	}
	list := chapters(a.status.ChapterJSON)
	if index < 0 || index >= len(list) {
		return
	}
	a.send("set_property", "chapter", index)
	a.status.Chapter = index
	a.pendingChapter.set(strconv.Itoa(index), time.Now())
	if a.status.Ended {
		a.send("set_property", "pause", false)
	}
	if close {
		a.chaptersOpen = false
		a.layoutHost()
	}
}
func (a *app) stepChapter(delta int) { a.selectChapter(a.status.Chapter+delta, false) }
func (a *app) audioMenu(c *ui.Context) ui.Element {
	trigger := icon(c, "audio", "Audio tracks").Disabled(!a.status.Loaded).OnClick(func() { a.audioOpen = !a.audioOpen; a.captionsOpen = false; a.layoutHost() })
	if a.fullscreen {
		return ui.Element{}
	}
	popup := frost(c, enterFrom(ui.Popover(c, trigger, &a.audioOpen, func() { a.audioPicker(c) }), trigger))
	if popup.Dismissed() {
		a.layoutHost()
	}
	return popup
}
func (a *app) audioPicker(c *ui.Context) {
	ui.Column(c).Width(320).Gap(8).Children(func() {
		ui.Text(c, "Audio tracks").FontSize(16).Bold().Padding(8)
		ui.Scroll(c).FillWidth().MaxHeight(260).MinHeight(0).Padding(0, scrollbarGutter, 0, 0).Children(func() {
			ui.Column(c).FillWidth().Gap(4).Children(func() {
				option := func(label, id string) {
					b := ui.Button(c.Key(id), label).FillWidth().Background(ui.Transparent).Border(0, ui.Transparent).Transition(colorFade)
					if a.status.AudioID == id {
						b.Background(c.Theme().Selection).TextColor(c.Theme().Accent)
					}
					b.OnClick(func() { a.selectAudio(id) })
				}
				option("Off", "no")
				tracks := audioTracks(a.status.TrackJSON)
				if len(tracks) == 0 {
					ui.Text(c, "No audio tracks.").TextColor(c.Theme().TextMuted)
				}
				for _, track := range tracks {
					option(track.label(), strconv.Itoa(track.ID))
				}
			})
		})
	})
}

func (a *app) chapterModal(c *ui.Context) {
	w, h := c.Size()
	dialog := frostedModal(c, &a.chaptersOpen, func() {
		ui.Column(c).Width(min(500, w-80)).Height(min(440, h-120)).Gap(16).Children(func() {
			ui.Row(c).FillWidth().AlignItems(ui.Center).Children(func() {
				ui.Text(c, "Chapters").FontSize(16).Bold()
				ui.Spacer(c)
				icon(c, "close", "Close chapters").OnClick(func() { a.chaptersOpen = false; a.layoutHost() })
			})
			list := chapters(a.status.ChapterJSON)
			if len(list) == 0 {
				ui.Text(c, "This video has no chapters.").TextColor(c.Theme().TextMuted)
			}
			ui.List(c.Key("chapters"), &a.chapterList, len(list)).FillWidth().Grow(1).MinHeight(0).Padding(0, scrollbarGutter, 0, 0).ItemKey(func(i int) any { return i }).Rows(func(row ui.ListRow) {
				i := row.Index
				navButton(row.Context, list[i].label(i), i == a.status.Chapter).OnClick(func() { a.selectChapter(i, true) })
			})
			ui.Row(c).FillWidth().Gap(12).Children(func() {
				ui.Button(c, "Previous chapter").Disabled(a.status.Chapter <= 0).OnClick(func() { a.stepChapter(-1) })
				ui.Spacer(c)
				ui.Button(c, "Next chapter").Disabled(a.status.Chapter+1 >= len(list)).OnClick(func() { a.stepChapter(1) })
			})
		})
	})
	enter(dialog)
	if dialog.Dismissed() {
		a.layoutHost()
	}
}
func (a *app) handlePlayerInput(action string) {
	switch action {
	case "fullscreen":
		if a.pip && a.page != "player" {
			a.showPlayer()
		} else {
			a.toggleFullscreen()
		}
	case "exit-fullscreen":
		if a.pip && a.page != "player" {
			a.closePip()
		} else {
			a.exitFullscreen()
		}
	case "pause":
		a.toggle()
	case "back":
		a.jump(-5)
	case "forward":
		a.jump(5)
	case "mute":
		a.send("cycle", "mute")
	case "next-chapter":
		a.stepChapter(1)
	case "previous-chapter":
		a.stepChapter(-1)
	case "settings":
		a.openSettings()
	case "add-folder":
		a.browseFolder()
	}
}
func (a *app) drainPlayerInput(p *player) {
	for {
		select {
		case action := <-p.input:
			a.handlePlayerInput(action)
		default:
			return
		}
	}
}

func (a *app) chapterStepButton(c *ui.Context, delta int) {
	name, label := "chapter-next", "Next chapter"
	if delta < 0 {
		name, label = "chapter-back", "Previous chapter"
	}
	target := a.status.Chapter + delta
	icon(c, name, label).Disabled(!a.status.Loaded || a.transition != nil || target < 0 || target >= len(chapters(a.status.ChapterJSON))).OnClick(func() { a.stepChapter(delta) })
}
