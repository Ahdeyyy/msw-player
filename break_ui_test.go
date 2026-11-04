package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// Worst-case data for the library, queue and track menus: realistic file
// names and folders rather than random strings. Rendered by
// MSWPLAYR_RENDER=1 go test -run TestBreakUI into build/break/.

var worstFolders = []string{
	`C:\Users\Aleksandra Wiśniewska-Kowalczyk\Videos\Familienurlaub Sommer 2019 – Ostsee (unbearbeitet, Rohmaterial von Jo)`,
	`D:\Jo`,
	`\\northwind-industries-holdings.example.com\media\Archive\映画コレクション`,
}

var worstNames = []string{
	"The.Lord.of.the.Rings.The.Fellowship.of.the.Ring.2001.EXTENDED.2160p.UHD.BluRay.x265.10bit.HDR.DTS-HD.MA.TrueHD.7.1.Atmos-SWTYBLZ",
	"A",
	"千と千尋の神隠し (2001) [1080p] [日本語字幕]",
	"الرسالة (1976) النسخة العربية الكاملة",
	"🎄 Christmas 2023 – Jo’s first steps 👣",
	"Mùa hè chiều thẳng đứng (2000) – Trần Anh Hùng",
	"Bartholomew Fitzgerald – Q3 all-hands recording (part 2 of 2, audio fixed, FINAL v12)",
	"ซีรีส์ ตอนที่ 12 – ฤดูร้อนที่ไม่มีวันลืม",
}

func breakBitmap(t *testing.T, w, h int) *ui.Bitmap {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{uint8(40 + 160*x/w), uint8(60 + 100*y/h), 140, 255})
		}
	}
	var b bytes.Buffer
	png.Encode(&b, img)
	bm, err := ui.DecodeBitmap(b.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	return bm
}

// breakFixture fills a with one of the data states.
func breakFixture(t *testing.T, a *app, state string) {
	wide, tall := breakBitmap(t, 320, 180), breakBitmap(t, 90, 160)
	add := func(folder, name string, i int) {
		path := filepath.Join(folder, name+".mkv")
		a.library = append(a.library, videoFile{Path: path, Name: name, Folder: folder})
		if i%3 == 0 {
			a.history[historyKey(path)] = bookmark{Position: 3*3600 + 12*60 + 45, Duration: 4 * 3600}
		}
		pos := a.resumePosition(path)
		switch i % 4 {
		case 0:
			a.thumbs[path] = &thumbnail{image: wide, position: pos}
		case 1:
			a.thumbs[path] = &thumbnail{image: tall, position: pos} // portrait phone video
		case 2:
			a.thumbs[path] = &thumbnail{position: pos} // still extracting or unreadable
		}
	}
	switch state {
	case "demo":
		a.folders = []string{`C:\Videos`, `C:\Videos\Trips`}
		for i := 0; i < 12; i++ {
			add(a.folders[i%2], fmt.Sprintf("Film %02d", i), 0)
		}
	case "worst":
		a.folders = worstFolders
		for i, name := range worstNames {
			add(worstFolders[i%len(worstFolders)], name, i)
		}
		// A nested subfolder whose relative path is itself long.
		add(filepath.Join(worstFolders[0], "Tag 3 – Strand", "GoPro HERO11 Black", "DCIM", "100GOPRO"), "GX010042", 3)
		a.libraryNotice = fmt.Sprintf("Some folders could not be read (%d). Check that your sources are available.", 2)
	case "empty":
		a.folders = []string{`C:\Videos`}
	case "one":
		a.folders = []string{`C:\Videos`}
		add(`C:\Videos`, "Film", 0)
	case "huge":
		a.folders = worstFolders
		for i := 0; i < 1200; i++ {
			add(worstFolders[i%len(worstFolders)], fmt.Sprintf("%s %04d", worstNames[i%len(worstNames)], i), i)
		}
	}
	for _, v := range a.library {
		a.queue = append(a.queue, v.Path)
	}
	if len(a.library) > 1 {
		a.selectedVideo = a.library[1].Path
	}
}

func breakTracks(state string) string {
	if state != "worst" {
		return `[{"id":1,"type":"sub","lang":"en"},{"id":1,"type":"audio","lang":"en"}]`
	}
	return `[
{"id":1,"type":"sub","lang":"eng","title":"English (SDH) – Commentary by director Peter Jackson, writers Fran Walsh & Philippa Boyens","hearing-impaired":true},
{"id":2,"type":"sub","lang":"jpn","title":"日本語 字幕 (劇場公開版)","forced":true},
{"id":3,"type":"sub","lang":"ara","title":"العربية","external":true},
{"id":4,"type":"sub"},
{"id":1,"type":"audio","lang":"eng","title":"English – Dolby TrueHD 7.1 Atmos (Director’s & Writers’ Commentary Track)"},
{"id":2,"type":"audio"}
]`
}

func saveBreak(t *testing.T, tt *ui.Tester, name string) {
	os.MkdirAll(filepath.Join("build", "break"), 0755)
	f, err := os.Create(filepath.Join("build", "break", name+".png"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	png.Encode(f, tt.Image())
}

func TestBreakUI(t *testing.T) {
	if os.Getenv("MSWPLAYR_RENDER") == "" {
		t.Skip("optional visual review")
	}
	type size struct {
		name string
		w, h int
		zoom float32
	}
	sizes := []size{{"1120", 1120, 760, 1}, {"760", 760, 540, 1}, {"zoom200", 1120, 760, 2}}
	for _, state := range []string{"demo", "worst", "empty", "one", "huge"} {
		for _, dark := range []bool{false, true} {
			theme := "light"
			if dark {
				theme = "dark"
			}
			for _, s := range sizes {
				name := fmt.Sprintf("%s-%s-%s", state, theme, s.name)
				a := newApp()
				a.dark = dark
				breakFixture(t, a, state)
				a.filterLibrary()
				tt := ui.NewTester(a.view, s.w, s.h)
				tt.SetZoom(s.zoom)
				tt.Frame()
				saveBreak(t, tt, "library-"+name)

				a.groupByFolder = true
				a.filterLibrary()
				tt.Frame()
				saveBreak(t, tt, "grouped-"+name)

				a.groupByFolder = false
				a.filterLibrary()
				a.showQueue = true
				tt.Frame()
				tt.Advance(400 * time.Millisecond)
				tt.Frame()
				saveBreak(t, tt, "queue-"+name)
				a.showQueue = false

				if len(a.queue) > 0 {
					a.page, a.current = "player", 0
					a.status = playback{Loaded: true, Paused: true, Duration: 4 * 3600, Position: 3*3600 + 12*60 + 45, TrackJSON: breakTracks(state), AudioID: "1"}
					a.seek = a.status.Position
					tt.Frame()
					saveBreak(t, tt, "player-"+name)
					if err := tt.Click("Captions"); err != nil {
						t.Fatal(err)
					}
					tt.Frame()
					tt.Advance(400 * time.Millisecond)
					tt.Frame()
					saveBreak(t, tt, "captions-"+name)
					a.captionsOpen = false
					if err := tt.Click("Audio tracks"); err != nil {
						t.Fatal(err)
					}
					tt.Frame()
					tt.Advance(400 * time.Millisecond)
					tt.Frame()
					saveBreak(t, tt, "audio-"+name)
				}
			}
		}
	}
	// Right-to-left, worst case only.
	a := newApp()
	breakFixture(t, a, "worst")
	a.filterLibrary()
	tt := ui.NewTester(a.view, 1120, 760)
	tt.SetDirection(ui.RightToLeft)
	tt.Frame()
	saveBreak(t, tt, "library-worst-rtl")
}
