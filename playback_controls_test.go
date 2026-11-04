package main

import (
	"fmt"
	"github.com/egoist/mygo/ui"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func controlsApp() *app {
	a := newApp()
	a.page = "player"
	a.current = 0
	a.queue = []string{"Film.mkv"}
	a.engine = &player{commands: make(chan []any, 128), input: make(chan string, 32), cancel: func() {}}
	a.status = playback{Loaded: true, Duration: 9, AudioID: "2", TrackJSON: `[{"id":2,"type":"audio","lang":"eng","title":"Stereo"},{"id":5,"type":"audio","lang":"fra","title":"Commentary"}]`, ChapterJSON: `[{"title":"Opening","time":0},{"title":"Middle","time":3},{"title":"Credits","time":6}]`, Chapter: 0}
	return a
}
func TestExactAudioAndChapters(t *testing.T) {
	a := controlsApp()
	tt := ui.NewTester(a.view, 1120, 760)
	if err := tt.Click("Audio tracks"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	tracks := audioTracks(a.status.TrackJSON)
	if err := tt.Click(tracks[1].label()); err != nil {
		t.Fatal(err)
	}
	command := <-a.engine.commands
	if command[0] != "set_property" || command[1] != "aid" || command[2] != 5 {
		t.Fatal(command)
	}
	tt.Frame()
	if err := tt.Click("Chapters"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !a.chaptersOpen || !tt.HasText("0:03  Middle") {
		t.Fatal(tt.Texts())
	}
	if err := tt.Click("0:03  Middle"); err != nil {
		t.Fatal(err)
	}
	command = <-a.engine.commands
	if command[1] != "chapter" || command[2] != 1 || a.chaptersOpen {
		t.Fatal(command)
	}
	a.status.Chapter = 2
	a.stepChapter(1)
	if len(a.engine.commands) != 0 {
		t.Fatal("stepped beyond last chapter")
	}
	a.status.Chapter = 0
	a.stepChapter(-1)
	if len(a.engine.commands) != 0 {
		t.Fatal("stepped before first chapter")
	}
}
func TestDoubleClickViewportAndSettingsPreservePlayer(t *testing.T) {
	a := controlsApp()
	tt := ui.NewTester(a.view, 1120, 760)
	r, ok := tt.Find("Video player")
	if !ok {
		t.Fatal("missing player")
	}
	x, y := r.X+r.W/2, r.Y+r.H/2
	tt.ClickAt(x, y)
	tt.ClickAt(x, y)
	tt.Frame()
	if !a.fullscreen {
		t.Fatal("double-click did not enter fullscreen")
	}
	tt.Advance(time.Second)
	tt.ClickAt(500, 300)
	tt.ClickAt(500, 300)
	tt.Frame()
	if a.fullscreen {
		t.Fatal("double-click did not leave fullscreen")
	}
	if err := tt.Click("Settings"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.page != "player" || !a.settingsOpen {
		t.Fatal("settings replaced playback page")
	}
	for len(a.engine.commands) > 0 {
		command := <-a.engine.commands
		if command[0] == "set_property" && command[1] == "pause" {
			t.Fatal("opening settings paused player")
		}
	}
	if err := tt.Click("Done"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.settingsOpen || a.page != "player" {
		t.Fatal("settings did not return to same player")
	}
}
func TestSubtitleAppearancePersistenceAndCommands(t *testing.T) {
	a := controlsApp()
	a.subtitleStyle.Font = "Arial"
	a.subtitleStyle.Size = 52
	a.subtitleStyle.Color = "Yellow"
	a.subtitleStyle.Background = true
	a.subtitleStyle.Bold = true
	a.subtitleStyle.OverrideASS = true
	a.subtitleStyle.Position = 65
	a.subtitleStyle.Alignment = "Right"
	a.applySubtitleAppearance()
	seen := map[string]any{}
	for len(a.engine.commands) > 0 {
		command := <-a.engine.commands
		seen[command[1].(string)] = command[2]
	}
	if seen["sub-font"] != "Arial" || seen["sub-font-size"] != float64(52) || seen["sub-color"] != "#ffe36a" || seen["sub-border-style"] != "background-box" || seen["sub-ass-override"] != "force" {
		t.Fatal(seen)
	}
	a.subtitleStyle.Size = 54
	a.applySubtitleAppearance()
	if len(a.engine.commands) != 1 {
		t.Fatal("unchanged styling resent")
	}
	<-a.engine.commands
	path := filepath.Join(t.TempDir(), "settings.json")
	a.store = newSettingsStore(path)
	a.persist()
	a.store.close()
	prefs, err := loadSettings(path)
	if err != nil || prefs.SubtitleStyle == nil || prefs.SubtitleStyle.Size != 54 || prefs.SubtitleStyle.Color != "Yellow" || prefs.SubtitleStyle.Position != 65 || prefs.SubtitleStyle.Alignment != "Right" {
		t.Fatalf("%+v %v", prefs, err)
	}
}
func TestNativeInputEventsAreQueued(t *testing.T) {
	a := controlsApp()
	a.engine.event([]byte(`{"event":"client-message","args":["mswplayr","fullscreen"]}`))
	a.engine.event([]byte(`{"event":"client-message","args":["mswplayr","fullscreen"]}`))
	if a.engine.snapshot().InputSerial != 2 {
		t.Fatal("input serial did not advance")
	}
	a.drainPlayerInput(a.engine)
	if a.fullscreen {
		t.Fatal("lost a fullscreen toggle")
	}
}
func TestCaptionAppearanceActionStaysVisible(t *testing.T) {
	a := controlsApp()
	tracks := []string{}
	for i := 0; i < 20; i++ {
		tracks = append(tracks, fmt.Sprintf(`{"id":%d,"type":"sub","lang":"eng","title":"Track %d"}`, i+1, i+1))
	}
	a.status.TrackJSON = "[" + strings.Join(tracks, ",") + "]"
	tt := ui.NewTester(a.view, 760, 540)
	if err := tt.Click("Captions"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	r, ok := tt.Find("Subtitle appearance…")
	if !ok || r.Y+r.H > 540 {
		t.Fatal(r)
	}
	if err := tt.Click("Subtitle appearance…"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !a.settingsOpen || a.page != "player" || a.captionsOpen {
		t.Fatal("appearance action did not open separate settings")
	}
}
func TestAudioChaptersAndNativeInputIntegration(t *testing.T) {
	if os.Getenv("MSWPLAYR_INTEGRATION") == "" {
		t.Skip("requires mpv and ffmpeg")
	}
	root := t.TempDir()
	metadata := filepath.Join(root, "chapters.txt")
	os.WriteFile(metadata, []byte(";FFMETADATA1\n[CHAPTER]\nTIMEBASE=1/1000\nSTART=0\nEND=3000\ntitle=Opening\n[CHAPTER]\nTIMEBASE=1/1000\nSTART=3000\nEND=6000\ntitle=Middle\n[CHAPTER]\nTIMEBASE=1/1000\nSTART=6000\nEND=9000\ntitle=Credits\n"), 0600)
	video := filepath.Join(root, "tracks.mkv")
	out, err := exec.Command("ffmpeg", "-v", "error", "-f", "lavfi", "-i", "color=c=black:size=640x360:rate=24", "-f", "lavfi", "-i", "sine=frequency=220", "-f", "lavfi", "-i", "sine=frequency=880", "-f", "ffmetadata", "-i", metadata, "-map", "0:v", "-map", "1:a", "-map", "2:a", "-map_chapters", "3", "-t", "9", "-c:v", "libx264", "-c:a", "aac", "-metadata:s:a:0", "language=eng", "-metadata:s:a:0", "title=Stereo", "-metadata:s:a:1", "language=fra", "-metadata:s:a:1", "title=Commentary", video).CombinedOutput()
	if err != nil {
		t.Fatalf("%v %s", err, out)
	}
	p := newPlayer(0)
	defer p.close()
	p.command("loadfile", video, "replace", -1, map[string]string{"pause": "yes"})
	wait := func(check func(playback) bool) {
		t.Helper()
		deadline := time.Now().Add(10 * time.Second)
		for time.Now().Before(deadline) {
			s := p.snapshot()
			if s.Error != "" {
				t.Fatal(s.Error)
			}
			if check(s) {
				return
			}
			time.Sleep(30 * time.Millisecond)
		}
		t.Fatalf("timeout: %+v", p.snapshot())
	}
	wait(func(s playback) bool {
		return s.Loaded && len(audioTracks(s.TrackJSON)) == 2 && len(chapters(s.ChapterJSON)) == 3
	})
	tracks := audioTracks(p.snapshot().TrackJSON)
	p.command("set_property", "aid", tracks[1].ID)
	wait(func(s playback) bool { return s.AudioID == strconv.Itoa(tracks[1].ID) })
	p.command("set_property", "chapter", 2)
	wait(func(s playback) bool { return s.Chapter == 2 && s.Position >= 5.9 })
	p.command("set_property", "chapter", 1)
	wait(func(s playback) bool { return s.Chapter == 1 && s.Position >= 2.9 && s.Position < 4 })
	p.command("keypress", "MBTN_LEFT_DBL")
	select {
	case action := <-p.input:
		if action != "fullscreen" {
			t.Fatal(action)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("native double-click binding did not emit fullscreen action")
	}
	p.command("keypress", "ESC")
	select {
	case action := <-p.input:
		if action != "exit-fullscreen" {
			t.Fatal(action)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("fullscreen escape binding missing")
	}
	// Verify the styling changes actual subtitle pixels, not just UI state.
	srt := filepath.Join(root, "style.srt")
	os.WriteFile(srt, []byte("1\n00:00:00,000 --> 00:00:09,000\nSubtitle styling test\n"), 0600)
	p.command("sub-add", srt, "select")
	wait(func(s playback) bool { return len(subtitleTracks(s.TrackJSON)) == 1 })
	a := newApp()
	a.engine = p
	a.subtitleStyle.Size = 52
	a.subtitleStyle.Color = "Yellow"
	a.subtitleStyle.OverrideASS = true
	a.subtitleStyle.Bold = true
	a.applySubtitleAppearance()
	found := false
	for attempt := 0; attempt < 6; attempt++ {
		file := filepath.Join(root, fmt.Sprintf("style-%d.png", attempt))
		p.command("screenshot-to-file", file, "subtitles")
		deadline := time.Now().Add(time.Second)
		for time.Now().Before(deadline) {
			if _, err := os.Stat(file); err == nil {
				break
			}
			time.Sleep(25 * time.Millisecond)
		}
		f, err := os.Open(file)
		if err != nil {
			continue
		}
		img, err := png.Decode(f)
		f.Close()
		if err != nil {
			continue
		}
		yellow := 0
		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
				r, g, b, _ := img.At(x, y).RGBA()
				if r > 45000 && g > 40000 && b < 30000 {
					yellow++
				}
			}
		}
		if yellow > 40 {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("mpv did not render the chosen subtitle color")
	}
}
func TestControlsRender(t *testing.T) {
	if os.Getenv("MSWPLAYR_RENDER") == "" {
		t.Skip("optional visual review")
	}
	for _, view := range []string{"audio", "chapters", "settings", "subtitle-style"} {
		a := controlsApp()
		a.dark = true
		if view == "audio" {
			a.audioOpen = true
		}
		if view == "chapters" {
			a.chaptersOpen = true
		}
		render := a.view
		if strings.HasPrefix(view, "subtitle-") {
			a.settingsTarget = "subtitle-appearance"
			render = a.settingsWindowView
		} else if view == "settings" {
			render = a.settingsWindowView
		}
		tt := ui.NewTester(render, 760, 600)
		tt.Frame()
		os.MkdirAll("build", 0755)
		f, err := os.Create(filepath.Join("build", fmt.Sprintf("controls-%s.png", view)))
		if err != nil {
			t.Fatal(err)
		}
		png.Encode(f, tt.Image())
		f.Close()
	}
}
