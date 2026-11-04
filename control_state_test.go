package main

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// The fullscreen tick hides its windows every 80 ms outside fullscreen; that
// must not close the window's own caption and audio popovers.
func TestHidingFullscreenWindowsKeepsPopoversOpen(t *testing.T) {
	a := controlsApp()
	tt := ui.NewTester(a.view, 1120, 760)
	if err := tt.Click("Captions"); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		a.hideFullscreenControls()
		tt.Advance(80 * time.Millisecond)
		tt.Frame()
	}
	if !a.captionsOpen || !tt.HasText("Load subtitle file…") {
		t.Fatal("caption popover closed by the fullscreen tick")
	}
	a.dismissFullscreenPicker()
	if a.captionsOpen || a.audioOpen {
		t.Fatal("dismissing the picker left a menu open")
	}
}

// A snapshot taken before mpv applied a choice must not flip the menus back.
func TestChoicesSurviveStaleSnapshots(t *testing.T) {
	a := controlsApp()
	a.status.SubtitleID = "no"
	a.chooseSubtitle("3")
	a.selectAudio("5")
	a.selectChapter(2, false)
	stale := a.status
	stale.SubtitleID, stale.AudioID, stale.Chapter, stale.Position = "", "2", 0, 4
	a.applyStatus(stale)
	if a.subtitleID != "3" || a.status.AudioID != "5" || a.status.Chapter != 2 {
		t.Fatalf("stale snapshot won: sid %q aid %q chapter %d", a.subtitleID, a.status.AudioID, a.status.Chapter)
	}
	// Stepping again starts from the chapter chosen, not the stale one.
	a.stepChapter(-1)
	if a.status.Chapter != 1 {
		t.Fatal(a.status.Chapter)
	}
	confirmed := stale
	confirmed.SubtitleID, confirmed.AudioID, confirmed.Chapter = "3", "5", 1
	a.applyStatus(confirmed)
	// Once confirmed, mpv's later values are taken as they come.
	later := confirmed
	later.SubtitleID, later.AudioID, later.Chapter = "", "2", 0
	a.applyStatus(later)
	if a.subtitleID != "no" || a.status.AudioID != "2" || a.status.Chapter != 0 {
		t.Fatalf("confirmed choice held on: sid %q aid %q chapter %d", a.subtitleID, a.status.AudioID, a.status.Chapter)
	}
}

func TestUnconfirmedChoiceGivesWay(t *testing.T) {
	var o optimistic
	now := time.Now()
	o.set("4", now)
	if o.resolve("1", now.Add(time.Second)) != "4" {
		t.Fatal("dropped before mpv could answer")
	}
	if o.resolve("1", now.Add(optimisticHold+time.Millisecond)) != "1" || o.value != "" {
		t.Fatal("held a choice mpv never applied")
	}
}

func TestSeekThumbHoldsWhileDragging(t *testing.T) {
	a := controlsApp()
	a.seek = 7
	a.seekHeld = time.Now().Add(time.Second)
	a.applyStatus(playback{Loaded: true, Duration: 9, Position: 2})
	if a.seek != 7 {
		t.Fatal("poll moved the thumb under the pointer")
	}
	a.seekHeld = time.Time{}
	a.applyStatus(playback{Loaded: true, Duration: 9, Position: 2})
	if a.seek != 2 {
		t.Fatal("thumb stopped following playback")
	}
}

// The fullscreen picker is ready only once it has measured the menu it was
// opened for, and switching menus measures again.
func TestPickerMeasuresEachMenuBeforeShowing(t *testing.T) {
	a := controlsApp()
	a.fullscreen = true
	a.captionsOpen, a.pickerKind, a.pickerSerial = true, "captions", 1
	tt := ui.NewTester(a.fullscreenPickerView, 328, 368)
	tt.Frame()
	if a.pickerReady == 1 {
		t.Fatal("ready on its first build, from stale bounds")
	}
	for i := 0; i < 4 && a.pickerReady != 1; i++ {
		tt.Advance(20 * time.Millisecond)
		tt.Frame()
	}
	if a.pickerReady != 1 || a.pickerSize.H <= 0 {
		t.Fatalf("captions never measured: ready %d size %+v", a.pickerReady, a.pickerSize)
	}
	captions := a.pickerSize
	// Switching flags alone, before the tick re-places the window, keeps the
	// content it was laid out for.
	a.captionsOpen, a.audioOpen = false, true
	tt.Frame()
	if !tt.HasText("Subtitle appearance…") {
		t.Fatal("content followed the flags ahead of the window")
	}
	a.pickerKind, a.pickerSerial = "audio", 2
	tt.Frame()
	for i := 0; i < 4 && a.pickerReady != 2; i++ {
		tt.Advance(20 * time.Millisecond)
		tt.Frame()
	}
	if a.pickerReady != 2 || a.pickerSize == captions || !tt.HasText("Commentary / fra (5)") {
		t.Fatalf("audio not measured: ready %d size %+v texts %v", a.pickerReady, a.pickerSize, tt.Texts())
	}
}

func TestSubtitleSearchFromFullscreenOpensInWindow(t *testing.T) {
	a := controlsApp()
	a.fullscreen, a.captionsOpen = true, true
	a.subtitleService = nil
	a.online.Busy = true // keep the search from starting a request
	a.openSubtitleSearch()
	if a.fullscreen || !a.onlineOpen || a.captionsOpen {
		t.Fatalf("fullscreen %v, search open %v, captions %v", a.fullscreen, a.onlineOpen, a.captionsOpen)
	}
	tt := ui.NewTester(a.view, 1120, 760)
	tt.Frame()
	if !tt.HasText("Find subtitles") {
		t.Fatal("search dialog not shown")
	}
}

func TestMenuHoleCoversOnlyTheMenu(t *testing.T) {
	// Video at (20,100) DIPs, 1080×556, at 150%.
	geometry := videoGeometry(1680, 1140, 1.5, false)
	hole := menuHole(ui.Rect{X: 740, Y: 386, W: 312, H: 300}, geometry, 1.5)
	// From 1110-30, 579-150 px in, cut at the video's bottom edge.
	if hole != [4]int{1080, 429, 468, geometry[3] - 429} {
		t.Fatalf("hole %v in video %v", hole, geometry)
	}
	if menuHole(ui.Rect{}, geometry, 1.5) != [4]int{} {
		t.Fatal("hole without a menu")
	}
	if menuHole(ui.Rect{X: 0, Y: 0, W: 10, H: 10}, geometry, 1.5) != [4]int{} {
		t.Fatal("hole for a menu outside the video")
	}
}

func TestOpenTrackMenuKeepsVideoAndRecordsHole(t *testing.T) {
	a := controlsApp()
	tt := ui.NewTester(a.view, 1120, 760)
	if err := tt.Click("Audio tracks"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	tt.Advance(250 * time.Millisecond)
	tt.Frame()
	if a.menuHole.W <= 0 || a.menuHole.H <= 0 {
		t.Fatalf("menu bounds not recorded: %+v", a.menuHole)
	}
	if err := tt.Click("Commentary / fra (5)"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	tt.Frame()
	if a.audioOpen || a.menuHole != (ui.Rect{}) {
		t.Fatalf("hole outlived the menu: open %v hole %+v", a.audioOpen, a.menuHole)
	}
}
