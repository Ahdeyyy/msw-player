package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/egoist/mygo/ui"
	"image"
	_ "image/jpeg"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestExperienceRender(t *testing.T) {
	if os.Getenv("MSWPLAYR_RENDER") == "" {
		t.Skip("optional visual review")
	}
	sample := filepath.Join("build", "qa-videos", "Sample.mkv")
	if _, err := os.Stat(sample); err != nil {
		t.Skip("optional sample missing")
	}
	data, err := extractThumbnail(context.Background(), filepath.Join("build", "visual-cache"), sample, 12)
	if err != nil {
		t.Fatal(err)
	}
	bitmap, err := ui.DecodeBitmap(data)
	if err != nil {
		t.Fatal(err)
	}
	for _, density := range []string{"Compact", "Comfortable", "Spacious"} {
		a := newApp()
		a.dark = true
		a.density = density
		a.folders = []string{"C:/Videos"}
		for i := 0; i < 40; i++ {
			path := fmt.Sprintf("C:/Videos/Film-%02d.mkv", i)
			a.library = append(a.library, videoFile{Path: path, Name: fmt.Sprintf("Film %02d", i), Folder: "C:/Videos"})
			a.history[historyKey(path)] = bookmark{Position: 12, Duration: 600}
			a.thumbs[path] = &thumbnail{image: bitmap, position: 12}
		}
		a.selectedVideo = a.library[1].Path
		a.filterLibrary()
		var bounds ui.Rect
		tt := ui.NewTester(func(c *ui.Context) { a.view(c); bounds = a.libraryList.Handle.Bounds(c) }, 1120, 760)
		tt.Frame()
		tt.Move(bounds.X+bounds.W-3, bounds.Y+50)
		row, ok := tt.Find("Add to queue: Film 00")
		if !ok || bounds.X+bounds.W-(row.X+row.W) < scrollbarGutter-1 {
			t.Fatalf("missing scrollbar gutter: list %+v control %+v", bounds, row)
		}
		f, err := os.Create(filepath.Join("build", "experience-"+density+".png"))
		if err != nil {
			t.Fatal(err)
		}
		png.Encode(f, tt.Image())
		f.Close()
		if density == "Comfortable" {
			tr := &openTransition{source: ui.Rect{X: 250, Y: 180, W: 104, H: 58.5}, image: bitmap, resume: true, position: 12}
			a.transition = tr
			a.page = "player"
			a.queue = []string{a.library[0].Path}
			a.current = 0
			at := ui.NewTester(func(c *ui.Context) {
				if tr.started.IsZero() {
					tr.started = c.Now()
				}
				a.view(c)
			}, 1120, 760)
			at.Advance(170 * time.Millisecond)
			f, _ := os.Create(filepath.Join("build", "experience-morph.png"))
			png.Encode(f, at.Image())
			f.Close()
		}
	}
}

func TestSelectionResumeAndTransition(t *testing.T) {
	a := newApp()
	path := "C:/Videos/Film.mkv"
	a.library = []videoFile{{Path: path, Name: "Film", Folder: "C:/Videos"}}
	a.filterLibrary()
	a.history[historyKey(path)] = bookmark{Position: 72.5, Duration: 600}
	a.engine = &player{commands: make(chan []any, 128), cancel: func() {}}
	tt := ui.NewTester(a.view, 1120, 760)
	if err := tt.Click("Video: Film"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.selectedVideo != path || a.page != "library" || len(a.engine.commands) != 0 {
		t.Fatal("first click must only select")
	}
	if err := tt.Click("Video: Film"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.page != "player" || a.transition == nil {
		t.Fatal("second click must start the morph")
	}
	command := <-a.engine.commands
	if command[0] != "loadfile" {
		t.Fatal(command)
	}
	options := command[4].(map[string]string)
	if options["start"] != "72.500" || options["pause"] != "yes" {
		t.Fatal(options)
	}
	a.status = playback{Loaded: true, Position: 72.5, Duration: 600}
	tt.Advance(600 * time.Millisecond)
	tt.Frame()
	if a.transition != nil {
		t.Fatal("resume morph did not hand off to player")
	}
	a.navigate("library")
	tt.Frame()
	tt.SetPreferences(ui.Preferences{ReduceMotion: true})
	if err := tt.Click("Video: Film"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.transition != nil {
		t.Fatal("reduced motion should skip morph")
	}
}
func TestUnwatchedMorphFadesAndFinishes(t *testing.T) {
	a := newApp()
	a.page = "player"
	a.engine = &player{commands: make(chan []any, 16), cancel: func() {}}
	tr := &openTransition{source: ui.Rect{X: 250, Y: 180, W: 100, H: 56}, started: time.Unix(1, 0)}
	a.transition = tr
	// Start the timer from the Tester's clock, then sample the loaded handoff.
	tt := ui.NewTester(func(c *ui.Context) {
		if tr.started.Equal(time.Unix(1, 0)) {
			tr.started = c.Now()
		}
		a.view(c)
	}, 1120, 760)
	a.status.Loaded = true
	tt.Advance(400 * time.Millisecond)
	if a.transition == nil || tr.fadeStarted.IsZero() {
		t.Fatal("unwatched image should fade after morph")
	}
	tt.Advance(250 * time.Millisecond)
	if a.transition != nil {
		t.Fatal("fade did not finish")
	}
}
func TestAppearanceDensityAndHistoryPersistence(t *testing.T) {
	a := newApp()
	a.dark = true
	tone := a.theme()
	if tone.Background.R != tone.Background.G || tone.Background.G != tone.Background.B || tone.Accent.R != tone.Accent.B {
		t.Fatal("dark palette is tinted")
	}
	a.page = "settings"
	tt := ui.NewTester(a.settingsWindowView, 1120, 760)
	if err := tt.Click("Compact"); err != nil {
		t.Fatal(err)
	}
	compact := a.densityMetrics()
	if err := tt.Click("Spacious"); err != nil {
		t.Fatal(err)
	}
	if a.densityMetrics().row <= compact.row || a.theme().Spacing <= 3 {
		t.Fatal("density did not increase spacing")
	}
	a.current = 0
	a.queue = []string{"C:/Film.mkv"}
	a.status = playback{Loaded: true, Position: 44.25, Duration: 500}
	a.checkpoint(true)
	path := filepath.Join(t.TempDir(), "settings.json")
	a.store = newSettingsStore(path)
	a.persist()
	a.store.close()
	prefs, err := loadSettings(path)
	if err != nil || prefs.Density != "Spacious" || prefs.History[historyKey(a.queue[0])].Position != 44.25 {
		t.Fatalf("%+v %v", prefs, err)
	}
	a.history[historyKey(a.queue[0])] = bookmark{Position: 500, Duration: 500, Completed: true}
	if a.resumePosition(a.queue[0]) != 0 {
		t.Fatal("completed video should restart")
	}
}
func TestMorphGeometry(t *testing.T) {
	source, target := ui.Rect{X: 240, Y: 160, W: 100, H: 56}, ui.Rect{X: 20, Y: 100, W: 1080, H: 556}
	if morphRect(source, target, 0) != source || morphRect(source, target, 1) != target {
		t.Fatal("morph endpoints")
	}
	mid := morphRect(source, target, .5)
	if mid.W <= source.W || mid.W >= target.W {
		t.Fatal(mid)
	}
}

func TestThumbnailAndResumeIntegration(t *testing.T) {
	if os.Getenv("MSWPLAYR_INTEGRATION") == "" {
		t.Skip("requires mpv and ffmpeg")
	}
	dir := t.TempDir()
	video := filepath.Join(dir, "Test film.mkv")
	cache := filepath.Join(dir, "cache")
	out, err := exec.Command("ffmpeg", "-v", "error", "-f", "lavfi", "-i", "color=red:s=320x180:r=24:d=1[a];color=blue:s=320x180:r=24:d=1[b];[a][b]concat=n=2:v=1:a=0", "-c:v", "libx264", video).CombinedOutput()
	if err != nil {
		t.Fatalf("%v %s", err, out)
	}
	first, err := extractThumbnail(context.Background(), cache, video, 0)
	if err != nil {
		t.Fatal(err)
	}
	saved, err := extractThumbnail(context.Background(), cache, video, 1.5)
	if err != nil {
		t.Fatal(err)
	}
	red, _, err := image.Decode(bytes.NewReader(first))
	if err != nil {
		t.Fatal(err)
	}
	blue, _, err := image.Decode(bytes.NewReader(saved))
	if err != nil {
		t.Fatal(err)
	}
	r, _, b, _ := red.At(100, 100).RGBA()
	r2, _, b2, _ := blue.At(100, 100).RGBA()
	if r <= b || b2 <= r2 {
		t.Fatal("thumbnail does not match requested position")
	}
	cached, err := extractThumbnail(context.Background(), cache, video, 1.5)
	if err != nil || !bytes.Equal(cached, saved) {
		t.Fatal("disk cache missed")
	}
	files, _ := filepath.Glob(filepath.Join(cache, "*.jpg"))
	if len(files) != 1 {
		t.Fatal("old thumbnails were not pruned")
	}
	p := newPlayer(0)
	defer p.close()
	p.command("loadfile", video, "replace", -1, map[string]string{"start": "1.500", "pause": "yes"})
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		s := p.snapshot()
		if s.Error != "" {
			t.Fatal(s.Error)
		}
		if s.Loaded && s.Paused && s.Position >= 1.45 {
			return
		}
		time.Sleep(40 * time.Millisecond)
	}
	t.Fatalf("resume failed: %+v", p.snapshot())
}
