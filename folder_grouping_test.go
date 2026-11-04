package main

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestFolderGroupingActionsAndPersistence(t *testing.T) {
	a := newApp()
	root := filepath.Join(t.TempDir(), "Videos")
	for _, item := range []struct{ folder, name string }{{"Series", "A"}, {"Movies", "B"}, {"Series", "C"}} {
		a.library = append(a.library, videoFile{Path: filepath.Join(root, item.folder, item.name+".mkv"), Folder: root, Name: item.name})
	}
	a.filterLibrary()
	path := filepath.Join(t.TempDir(), "settings.json")
	a.store = newSettingsStore(path)
	tt := ui.NewTester(a.view, 1120, 760)
	if err := tt.Click("Group by folder"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !a.groupByFolder || len(a.libraryRows) != 5 || a.libraryRows[0].count != 1 || a.libraryRows[2].count != 2 {
		t.Fatal(a.libraryRows)
	}
	if a.filtered[0].Name != "B" || a.filtered[1].Name != "A" {
		t.Fatal(a.filtered)
	}
	if err := tt.Click("Add to queue: A"); err != nil {
		t.Fatal(err)
	}
	if len(a.queue) != 1 || a.queue[0] != a.library[0].Path {
		t.Fatal(a.queue)
	}
	if err := tt.Click("Video: A"); err != nil {
		t.Fatal(err)
	}
	if a.selectedVideo != a.library[0].Path {
		t.Fatal(a.selectedVideo)
	}
	if os.Getenv("MSWPLAYR_RENDER") != "" {
		os.MkdirAll("build", 0755)
		f, err := os.Create("build/library-grouped.png")
		if err != nil {
			t.Fatal(err)
		}
		png.Encode(f, tt.Image())
		f.Close()
	}
	a.store.close()
	prefs, err := loadSettings(path)
	if err != nil || !prefs.GroupByFolder {
		t.Fatalf("%+v %v", prefs, err)
	}
	a.store = nil
	a.search = "C"
	a.filterLibrary()
	if len(a.libraryRows) != 2 || a.libraryRows[0].count != 1 || a.filtered[0].Name != "C" {
		t.Fatal(a.libraryRows)
	}
	a.search = ""
	a.groupByFolder = false
	a.filterLibrary()
	if len(a.libraryRows) != 3 || a.filtered[0].Name != "A" {
		t.Fatal(a.filtered)
	}
}
