package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestThumbnailSchedulingRecovers(t *testing.T) {
	a := newApp()
	path := "film.mkv"
	entry := a.thumbnailFor(path) // A view can build before the worker starts.
	a.thumbJobs = make(chan thumbnailJob, 1)
	if a.thumbnailFor(path) != entry || !entry.pending || len(a.thumbJobs) != 1 {
		t.Fatal("early entry was never scheduled after worker startup")
	}
	a.thumbnailFor(path)
	if len(a.thumbJobs) != 1 {
		t.Fatal("duplicate job for a pending entry")
	}
	other := a.thumbnailFor("other.mkv") // Queue is full.
	if other.pending {
		t.Fatal("unscheduled entry marked pending")
	}
	<-a.thumbJobs
	if a.thumbnailFor("other.mkv") != other || !other.pending {
		t.Fatal("saturated queue permanently lost a thumbnail")
	}
	<-a.thumbJobs
	entry.pending = false
	entry.retryAt = time.Now().Add(time.Minute)
	a.thumbnailFor(path)
	if len(a.thumbJobs) != 0 {
		t.Fatal("failed job ignored retry delay")
	}
	entry.retryAt = time.Now().Add(-time.Second)
	a.thumbnailFor(path)
	if !entry.pending || len(a.thumbJobs) != 1 {
		t.Fatal("failed thumbnail was never retried")
	}
}

func TestLargeLibraryThumbnailsRenderAfterLoad(t *testing.T) {
	a := newApp()
	a.thumbJobs = make(chan thumbnailJob, 24)
	for i := 0; i < 150; i++ {
		a.library = append(a.library, videoFile{Path: fmt.Sprintf("film-%03d.mkv", i), Name: fmt.Sprintf("Film %03d", i)})
	}
	a.filterLibrary()
	tt := ui.NewTester(a.view, 1120, 760)
	img := image.NewRGBA(image.Rect(0, 0, 32, 18))
	for y := 0; y < 18; y++ {
		for x := 0; x < 32; x++ {
			img.Set(x, y, color.RGBA{R: 240, G: 30, B: 40, A: 255})
		}
	}
	bitmap := ui.NewBitmap(img)
	for _, atEnd := range []bool{false, true, false} {
		if atEnd {
			a.libraryList.ScrollToEnd()
		} else {
			a.libraryList.ScrollTo(0, ui.Start)
		}
		tt.Frame()
		for pass := 0; pass < 8; pass++ {
			for len(a.thumbJobs) > 0 {
				job := <-a.thumbJobs
				if a.thumbs[job.path] == job.entry {
					job.entry.image, job.entry.pending = bitmap, false
				}
			}
			tt.Frame()
		}
		first, last := a.libraryList.Visible()
		if first < 0 || last < first {
			t.Fatal("library has no visible rows")
		}
		for i := first; i <= last; i++ {
			entry := a.thumbs[a.filtered[i].Path]
			if entry == nil || entry.image == nil {
				t.Fatalf("visible thumbnail %d never loaded", i)
			}
		}
		pixels := tt.Image()
		red := 0
		for y := 0; y < pixels.Bounds().Dy(); y++ {
			for x := 0; x < pixels.Bounds().Dx(); x++ {
				p := pixels.RGBAAt(x, y)
				if p.R > 200 && p.G < 60 && p.B < 60 {
					red++
				}
			}
		}
		if red < 2000 {
			t.Fatal("loaded thumbnails are absent from the rendered list")
		}
	}
}

func TestDarkOpeningThumbnailIntegration(t *testing.T) {
	if os.Getenv("MSWPLAYR_INTEGRATION") == "" {
		t.Skip("requires mpv and ffmpeg")
	}
	dir := t.TempDir()
	video, cache := filepath.Join(dir, "Dark opening.mkv"), filepath.Join(dir, "cache")
	output, err := exec.Command("ffmpeg", "-v", "error", "-f", "lavfi", "-i", "color=black:s=320x180:r=24:d=2[a];color=red:s=320x180:r=24:d=2[b];[a][b]concat=n=2:v=1:a=0", "-c:v", "libx264", video).CombinedOutput()
	if err != nil {
		t.Fatalf("%v: %s", err, output)
	}
	preview, err := extractThumbnail(context.Background(), cache, video, 0)
	if err != nil {
		t.Fatal(err)
	}
	brightness, err := thumbnailBrightness(preview)
	if err != nil || brightness < 8 {
		t.Fatalf("unwatched preview retained the black opening: %v %v", brightness, err)
	}
	files, _ := filepath.Glob(filepath.Join(cache, "*.jpg"))
	if len(files) != 1 {
		t.Fatal(files)
	}
	if err := os.WriteFile(files[0], []byte("truncated jpeg"), 0600); err != nil {
		t.Fatal(err)
	}
	repaired, err := extractThumbnail(context.Background(), cache, video, 0)
	if err != nil || !bytes.Equal(repaired, preview) {
		t.Fatalf("damaged cached preview was not repaired: %v", err)
	}
	saved, err := extractThumbnail(context.Background(), cache, video, .5)
	if err != nil {
		t.Fatal(err)
	}
	brightness, err = thumbnailBrightness(saved)
	if err != nil || brightness > 1 {
		t.Fatalf("saved playback position must remain exact, even in a dark scene: %v %v", brightness, err)
	}
}
