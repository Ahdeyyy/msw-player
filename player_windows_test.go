package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestSubtitlePlaybackIntegration(t *testing.T) {
	if os.Getenv("MSWPLAYR_INTEGRATION") == "" {
		t.Skip("optional mpv integration")
	}
	dir := t.TempDir()
	srt := filepath.Join(dir, "captions.srt")
	if err := os.WriteFile(srt, []byte("1\n00:00:00,000 --> 00:00:04,000\nTest captions\n"), 0600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "multitrack.mkv")
	out, err := exec.Command("ffmpeg", "-v", "error", "-f", "lavfi", "-i", "testsrc2=size=320x180:rate=24", "-i", srt, "-map", "0:v", "-map", "1:s", "-map", "1:s", "-t", "4", "-c:v", "libx264", "-c:s", "srt", "-metadata:s:s:0", "language=eng", "-metadata:s:s:1", "language=fra", path).CombinedOutput()
	if err != nil {
		t.Fatalf("%v %s", err, out)
	}
	p := newPlayer(0)
	defer p.close()
	p.command("loadfile", path, "replace")
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
			time.Sleep(40 * time.Millisecond)
		}
		t.Fatalf("timeout: %+v", p.snapshot())
	}
	wait(func(s playback) bool { return s.Loaded && len(subtitleTracks(s.TrackJSON)) == 2 })
	p.command("set_property", "pause", true)
	tracks := subtitleTracks(p.snapshot().TrackJSON)
	id := tracks[1].ID
	p.command("set_property", "sid", id)
	wait(func(s playback) bool { return s.SubtitleID == strconv.Itoa(id) })
	p.command("set_property", "sid", "no")
	wait(func(s playback) bool { return s.SubtitleID == "no" })
	p.command("sub-add", srt, "select")
	wait(func(s playback) bool { return len(subtitleTracks(s.TrackJSON)) == 3 && s.SubtitleID != "no" })
}
func TestLoadRetainsUnchangedObservedProperties(t *testing.T) {
	p := &player{state: playback{Loaded: true, Ended: true, Width: 1920, Height: 1080, Duration: 100, Codec: "h264", Position: 99, Error: "old"}}
	p.prepareLoad()
	s := p.snapshot()
	if s.Generation == 0 {
		t.Fatal("load must invalidate an otherwise identical paused snapshot")
	}
	if s.Loaded || s.Ended || s.Position != 0 || s.Error != "" || s.Width != 1920 || s.Duration != 100 {
		t.Fatal(s)
	}
}

func TestPlaybackIntegration(t *testing.T) {
	if os.Getenv("MSWPLAYR_INTEGRATION") == "" {
		t.Skip("set MSWPLAYR_INTEGRATION=1 to exercise mpv and ffmpeg")
	}
	for _, sample := range []struct {
		ext, codec, audio string
		options           []string
	}{
		{"mp4", "libx264", "aac", nil}, {"mkv", "libx264", "aac", nil}, {"mov", "libx264", "aac", nil},
		{"mkv", "libx265", "aac", []string{"-x265-params", "log-level=error:pools=2"}},
		{"webm", "libvpx-vp9", "libopus", []string{"-deadline", "realtime", "-cpu-used", "8"}},
		{"mp4", "libaom-av1", "aac", []string{"-cpu-used", "8"}},
	} {
		t.Run(sample.ext+"-"+sample.codec, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "sample."+sample.ext)
			args := []string{"-v", "error", "-f", "lavfi", "-i", "testsrc2=size=320x180:rate=24", "-f", "lavfi", "-i", "sine=frequency=440", "-t", "4", "-c:v", sample.codec, "-pix_fmt", "yuv420p", "-c:a", sample.audio}
			args = append(args, sample.options...)
			args = append(args, path)
			out, err := exec.Command("ffmpeg", args...).CombinedOutput()
			if err != nil {
				t.Fatalf("ffmpeg: %v %s", err, out)
			}
			p := newPlayer(0)
			defer p.close()
			p.command("loadfile", path, "replace")
			wait := func(check func(playback) bool) {
				t.Helper()
				deadline := time.Now().Add(12 * time.Second)
				for time.Now().Before(deadline) {
					s := p.snapshot()
					if s.Error != "" {
						t.Fatal(s.Error)
					}
					if check(s) {
						return
					}
					time.Sleep(40 * time.Millisecond)
				}
				t.Fatalf("timed out: %+v", p.snapshot())
			}
			wait(func(s playback) bool { return s.Loaded && s.Duration > 3 && s.Position > .15 })
			p.command("set_property", "pause", true)
			wait(func(s playback) bool { return s.Paused })
			p.command("seek", 2, "absolute+exact")
			wait(func(s playback) bool { return s.Position >= 1.9 })
			p.command("set_property", "pause", false)
			wait(func(s playback) bool { return !s.Paused && s.Position > 2.1 })
		})
	}
}
