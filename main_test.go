package main

import (
	"fmt"
	"github.com/egoist/mygo/ui"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestPlayerView(t *testing.T) {
	a := newApp()
	tt := ui.NewTester(a.view, 1120, 760)
	if !tt.HasText("Add your video folders") {
		t.Fatal(tt.Texts())
	}
	if err := tt.Click("Settings"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText("Video folders") {
		t.Fatal(tt.Texts())
	}
	if err := tt.Click("Dark appearance"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !a.dark {
		t.Fatal("theme did not change")
	}
	a.page = "player"
	a.queue = []string{"Movie.mkv"}
	a.current = 0
	a.status = playback{Loaded: true, Codec: "H.264", Width: 1920, Height: 1080, SubtitleID: "no"}
	tt.Frame()
	if tt.HasText("H.264") || tt.HasText("1920") {
		t.Fatal("player exposes diagnostics")
	}
	if err := tt.Click("Play queue"); err != nil {
		t.Fatal(err)
	}
	if !a.showQueue {
		t.Fatal("queue did not open")
	}
	tt.Frame()
	if err := tt.Click("Close queue"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.showQueue {
		t.Fatal("queue did not close")
	}
	if err := tt.Click("Fullscreen"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !a.fullscreen {
		t.Fatal("fullscreen did not open")
	}
	if len(tt.Texts()) != 0 {
		t.Fatalf("fullscreen contains controls: %v", tt.Texts())
	}
	a.exitFullscreen()
	tt.Frame()
	if !tt.HasText("Library") {
		t.Fatal("controls did not return")
	}
}
func TestFormatsAndClock(t *testing.T) {
	for _, name := range []string{"movie.MKV", "movie.mp4", "movie.mov", "movie.webm"} {
		if !supported(name) {
			t.Fatal(name)
		}
	}
	if supported("movie.txt") {
		t.Fatal("accepted non-video")
	}
	for sec, want := range map[float64]string{0: "0:00", 65: "1:05", 3661: "1:01:01", -2: "0:00"} {
		if got := clock(sec); got != want {
			t.Fatalf("%v: %s", sec, got)
		}
	}
}
func TestRender(t *testing.T) {
	if os.Getenv("MSWPLAYR_RENDER") == "" {
		t.Skip("optional visual review")
	}
	a := newApp()
	for _, dark := range []bool{false, true} {
		a.dark = dark
		for _, size := range [][2]int{{1120, 760}, {760, 540}} {
			for _, page := range []string{"library", "settings", "player", "queue", "captions", "fullscreen"} {
				a.page = page
				a.fullscreen = page == "fullscreen"
				a.showQueue = page == "queue"
				a.captionsOpen = page == "captions"
				a.folders = []string{"C:/Videos", "D:/Series"}
				a.library = []videoFile{{Name: "The Grand Budapest Hotel", Path: "C:/Videos/The Grand Budapest Hotel.mkv", Folder: "C:/Videos"}, {Name: "Planet Earth S01E01", Path: "D:/Series/Planet Earth S01E01.mkv", Folder: "D:/Series"}}
				a.filterLibrary()
				a.current = 0
				a.queue = []string{a.library[0].Path, a.library[1].Path}
				a.status = playback{Loaded: true, Duration: 5442, Position: 122, TrackJSON: `[{"id":3,"type":"sub","lang":"eng","title":"English"},{"id":6,"type":"sub","lang":"fra","title":"French"}]`}
				a.seek = 122
				if page == "queue" || page == "captions" || page == "fullscreen" {
					a.page = "player"
				}
				img := ui.Render(a.view, size[0], size[1], 1)
				os.MkdirAll("build", 0755)
				name := "light"
				if dark {
					name = "dark"
				}
				f, err := os.Create(filepath.Join("build", fmt.Sprintf("%s-%s-%d.png", name, page, size[0])))
				if err != nil {
					t.Fatal(err)
				}
				png.Encode(f, img)
				f.Close()
			}
		}
	}
}
