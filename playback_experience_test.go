package main

import (
	"image/png"
	"os"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestResumeWaitsForVideoFrame(t *testing.T) {
	a := controlsApp()
	tr := &openTransition{source: ui.Rect{X: 200, Y: 150, W: 100, H: 56}, resume: true, position: 3}
	a.transition = tr
	a.status.Position = 3
	tt := ui.NewTester(func(c *ui.Context) {
		if tr.started.IsZero() {
			tr.started = c.Now()
		}
		a.view(c)
	}, 1120, 760)
	tt.Advance(time.Second)
	if a.transition == nil {
		t.Fatal("loaded position revealed before a video frame was ready")
	}
	a.status.FrameReady = true
	tt.Frame()
	if a.transition != nil {
		t.Fatal("ready resume frame did not complete handoff")
	}
	command := <-a.engine.commands
	if command[1] != "pause" || command[2] != false {
		t.Fatal(command)
	}
}

func TestFrameReadinessLifecycle(t *testing.T) {
	p := &player{}
	p.prepareLoad()
	p.event([]byte(`{"event":"file-loaded"}`))
	if p.snapshot().FrameReady {
		t.Fatal("file-loaded is not a rendered frame")
	}
	p.event([]byte(`{"event":"playback-restart"}`))
	if !p.snapshot().FrameReady {
		t.Fatal("missing playback readiness")
	}
	p.event([]byte(`{"event":"seek"}`))
	if p.snapshot().FrameReady {
		t.Fatal("seek retained stale readiness")
	}
	p.event([]byte(`{"event":"playback-restart"}`))
	p.prepareLoad()
	if p.snapshot().FrameReady {
		t.Fatal("new file retained stale readiness")
	}
}

func TestChapterButtonsInBothControlBars(t *testing.T) {
	for _, fullscreen := range []bool{false, true} {
		a := controlsApp()
		render := a.view
		height := 540
		if fullscreen {
			render = a.fullscreenControlsView
			height = 112
		}
		tt := ui.NewTester(render, 760, height)
		if err := tt.Click("Next chapter"); err != nil {
			t.Fatal(err)
		}
		command := <-a.engine.commands
		if command[1] != "chapter" || command[2] != 1 {
			t.Fatal(command)
		}
		tt.Frame()
		if err := tt.Click("Previous chapter"); err != nil {
			t.Fatal(err)
		}
		command = <-a.engine.commands
		if command[2] != 0 {
			t.Fatal(command)
		}
		if os.Getenv("MSWPLAYR_RENDER") != "" && fullscreen {
			file, err := os.Create("build/controls-fullscreen-bar.png")
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()
			if err := png.Encode(file, tt.Image()); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestSubtitlePlacementPresetsApply(t *testing.T) {
	a := controlsApp()
	a.settingsTarget = "subtitle-appearance"
	tt := ui.NewTester(a.settingsWindowView, 760, 900)
	if err := tt.Click("Place subtitles at top"); err != nil {
		t.Fatal(err)
	}
	if a.subtitleStyle.Position != 0 {
		t.Fatal(a.subtitleStyle.Position)
	}
	a.subtitleStyle.Alignment = "Right"
	a.subtitleAppearanceChanged()
	seen := map[string]any{}
	for len(a.engine.commands) > 0 {
		cmd := <-a.engine.commands
		seen[cmd[1].(string)] = cmd[2]
	}
	if seen["sub-pos"] != float64(0) || seen["sub-align-x"] != "right" {
		t.Fatal(seen)
	}
	if normalizedSubtitleAppearance(subtitleAppearance{}).Alignment != "Center" {
		t.Fatal("old settings need centered alignment")
	}
}

func TestQueueButtonConfirmsThenReverts(t *testing.T) {
	a := newApp()
	a.folders = []string{"C:/Videos"}
	a.library = []videoFile{{Path: "C:/Videos/A.mkv", Name: "A", Folder: "C:/Videos"}}
	a.filterLibrary()
	tt := ui.NewTester(a.view, 1120, 760)
	tt.Frame()
	if err := tt.Click("Add to queue: A"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if len(a.queue) != 1 || a.queuedPath != "C:/Videos/A.mkv" {
		t.Fatalf("queue %v, confirming %q", a.queue, a.queuedPath)
	}
	if !tt.FramePending() {
		t.Fatal("check icon schedules no frame to revert")
	}
	tt.Advance(queuedFeedback + 50*time.Millisecond)
	tt.Frame()
	if _, ok := tt.Find("Add to queue: A"); !ok {
		t.Fatal("queue button lost its label")
	}
}
