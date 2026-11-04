package main

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"image/png"
	"os"
	"testing"
)

func TestFullscreenPickerDoesNotChangeBar(t *testing.T) {
	for _, picker := range []string{"captions", "audio"} {
		t.Run(picker, func(t *testing.T) {
			a := controlsApp()
			a.fullscreen = true
			screen := mygo.Rectangle{Width: 1920, Height: 1080}
			before := fullscreenControlsBounds(screen)
			bar := ui.NewTester(a.fullscreenControlsView, before.Width, before.Height)
			seekBefore, _ := bar.Find("Seek")
			trigger := "Captions"
			if picker == "audio" {
				trigger = "Audio tracks"
			}
			if err := bar.Click(trigger); err != nil {
				t.Fatal(err)
			}
			bar.Frame()
			bar.Frame()
			if !a.captionsOpen && !a.audioOpen {
				t.Fatal("picker did not open")
			}
			seekAfter, _ := bar.Find("Seek")
			if before != fullscreenControlsBounds(screen) || seekBefore != seekAfter || before.Height != 112 {
				t.Fatal("picker moved or resized controls")
			}
			if bar.HasText("Off") {
				t.Fatal("picker rendered inside control bar")
			}
			a.pickerKind = picker
			popup := ui.NewTester(a.fullscreenPickerView, 328, 368)
			popup.Frame()
			popup.Frame()
			if a.pickerSize.H <= 0 || a.pickerSize.H > 368 {
				t.Fatal(a.pickerSize)
			}
			if picker == "captions" && !popup.HasText("Subtitle appearance\u2026") {
				t.Fatal(popup.Texts())
			}
			if os.Getenv("MSWPLAYR_RENDER") != "" {
				os.MkdirAll("build", 0755)
				f, err := os.Create("build/fullscreen-picker-" + picker + ".png")
				if err != nil {
					t.Fatal(err)
				}
				png.Encode(f, popup.Image())
				f.Close()
			}
			if err := popup.Click("Off"); err != nil {
				t.Fatal(err)
			}
			popup.Frame()
			popup.Frame()
			bar.Frame()
			if a.audioOpen || a.captionsOpen {
				t.Fatal("picker survived selection")
			}
			if err := bar.Click(trigger); err != nil {
				t.Fatal(err)
			}
			bar.Frame()
			popup.Frame()
			popup.Frame()
			if !popup.HasText("Off") {
				t.Fatal("picker did not reopen")
			}
			a.dismissFullscreenPicker()
			popup.Frame()
			popup.Frame()
			bar.Frame()
			if a.audioOpen || a.captionsOpen || before != fullscreenControlsBounds(screen) {
				t.Fatal("dismissal changed controls")
			}
		})
	}
}
