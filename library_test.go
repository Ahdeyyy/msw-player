package main

import (
	"context"
	"github.com/egoist/mygo/ui"
	"os"
	"path/filepath"
	"testing"
)

func TestFolderScanAndFilter(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "Series")
	os.Mkdir(nested, 0755)
	for _, path := range []string{filepath.Join(root, "Film.MKV"), filepath.Join(nested, "Episode.mp4"), filepath.Join(root, "notes.txt")} {
		if err := os.WriteFile(path, []byte("x"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	files, warnings := scanFolders(context.Background(), []string{root, nested, filepath.Join(root, "missing")})
	if len(files) != 2 || len(warnings) != 1 {
		t.Fatalf("files %v warnings %v", files, warnings)
	}
	a := newApp()
	a.library = files
	a.search = "episode"
	a.filterLibrary()
	if len(a.filtered) != 1 || a.filtered[0].Name != "Episode" {
		t.Fatal(a.filtered)
	}
	a.search = ""
	a.folderFilter = nested
	a.filterLibrary()
	if len(a.filtered) != 1 {
		t.Fatal(a.filtered)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	files, _ = scanFolders(ctx, []string{root})
	if len(files) != 0 {
		t.Fatal("cancelled scan continued")
	}
}
func TestSavedSources(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	store := newSettingsStore(path)
	store.jobs <- preferences{Folders: []string{"C:/Videos"}}
	store.jobs <- preferences{Folders: []string{"C:/Movies", "D:/Series"}, Dark: true}
	store.close()
	prefs, err := loadSettings(path)
	if err != nil || !prefs.Dark || len(prefs.Folders) != 2 {
		t.Fatalf("%+v %v", prefs, err)
	}
}
func TestCaptionMenu(t *testing.T) {
	a := newApp()
	a.page = "player"
	a.current = 0
	a.queue = []string{"Film.2024.1080p.mkv"}
	a.status = playback{Loaded: true, TrackJSON: `[{"id":2,"type":"sub","lang":"eng","title":"English"},{"id":5,"type":"sub","lang":"fra","title":"French","forced":true},{"id":1,"type":"audio"}]`}
	a.subtitleID = "no"
	tt := ui.NewTester(a.view, 1120, 760)
	if err := tt.Click("Captions"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	tracks := subtitleTracks(a.status.TrackJSON)
	if len(tracks) != 2 {
		t.Fatal(tracks)
	}
	if err := tt.Click(tracks[1].label()); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.subtitleID != "5" || a.captionsOpen {
		t.Fatal("exact subtitle not selected")
	}
	if err := tt.Click("Captions"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if err := tt.Click("Off"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.subtitleID != "no" {
		t.Fatal("captions not disabled")
	}
	if err := tt.Click("Captions"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if err := tt.Click("Find subtitles online…"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !a.onlineOpen || a.subtitleQuery != "Film 2024" {
		t.Fatalf("search %+v %s", a.onlineOpen, a.subtitleQuery)
	}
	if err := tt.Click("Close subtitle search"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.onlineOpen {
		t.Fatal("search did not close")
	}
}
func TestVideoGeometry(t *testing.T) {
	for _, s := range []float64{1, 1.5, 2} {
		full := videoGeometry(1920, 1080, s, true)
		if full != [4]int{0, 0, 1920, 1080} {
			t.Fatal(full)
		}
		normal := videoGeometry(1920, 1080, s, false)
		if normal[0] != int(20*s) || normal[1] != int(100*s) || normal[3] != 1080-int(204*s) {
			t.Fatal(normal)
		}
	}
}
func TestSubtitleSearch(t *testing.T) {
	if trackID([]byte("false")) != "no" || trackID([]byte("7")) != "7" {
		t.Fatal("invalid track IDs")
	}
}
