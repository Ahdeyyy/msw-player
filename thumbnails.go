package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/egoist/mygo/ui"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type bookmark struct {
	Position  float64 `json:"position"`
	Duration  float64 `json:"duration"`
	Completed bool    `json:"completed,omitempty"`
}

func boolOption(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}
func historyKey(path string) string { return strings.ToLower(filepath.Clean(path)) }
func (a *app) resumePosition(path string) float64 {
	b := a.history[historyKey(path)]
	if b.Completed || math.IsNaN(b.Position) || math.IsInf(b.Position, 0) || b.Position < 0 {
		return 0
	}
	if b.Duration > 0 && b.Position >= b.Duration-.5 {
		return 0
	}
	return b.Position
}
func (a *app) checkpoint(force bool) {
	if a.current < 0 || a.current >= len(a.queue) || !a.status.Loaded {
		return
	}
	if !force && time.Since(a.lastCheckpoint) < 10*time.Second && !a.status.Paused && !a.status.Ended {
		return
	}
	b := bookmark{Position: a.status.Position, Duration: a.status.Duration, Completed: a.status.Ended}
	if math.IsNaN(b.Position) || math.IsInf(b.Position, 0) || b.Position < 0 {
		return
	}
	key := historyKey(a.queue[a.current])
	if a.history[key] == b {
		return
	}
	a.history[key] = b
	a.lastCheckpoint = time.Now()
	a.persist()
}

type thumbnail struct {
	image    *ui.Bitmap
	handle   ui.Handle
	position float64
	pending  bool
	used     uint64
}
type thumbnailJob struct {
	path     string
	position float64
	entry    *thumbnail
}

func (a *app) startThumbnailWorker() {
	ctx, cancel := context.WithCancel(context.Background())
	a.thumbCancel = cancel
	a.thumbJobs = make(chan thumbnailJob, 24)
	window := a.window
	cacheDir := a.cacheDir
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case job := <-a.thumbJobs:
				data, err := extractThumbnail(ctx, cacheDir, job.path, job.position)
				if ctx.Err() != nil {
					return
				}
				var bitmap *ui.Bitmap
				if err == nil {
					bitmap, _ = ui.DecodeBitmap(data)
				}
				window.Update(func() {
					if a.thumbs[job.path] != job.entry {
						return
					}
					job.entry.pending = false
					job.entry.image = bitmap
				})
			}
		}
	}()
}

// Called only for visible rows. The pending entry guards the background job
// across repeated builds; scrolling cannot create duplicate decode work.
func (a *app) thumbnailFor(path string) *thumbnail {
	position := a.resumePosition(path)
	entry := a.thumbs[path]
	if entry == nil || entry.position != position {
		entry = &thumbnail{position: position}
		a.thumbs[path] = entry
		if a.thumbJobs != nil {
			select {
			case a.thumbJobs <- thumbnailJob{path, position, entry}:
				entry.pending = true
			default:
				delete(a.thumbs, path)
			}
		}
	}
	a.thumbClock++
	entry.used = a.thumbClock
	if len(a.thumbs) > 96 {
		var oldest string
		stamp := ^uint64(0)
		for key, value := range a.thumbs {
			if key != path && value.used < stamp {
				oldest, stamp = key, value.used
			}
		}
		delete(a.thumbs, oldest)
	}
	return entry
}
func extractThumbnail(ctx context.Context, cacheDir, path string, position float64) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	prefix := fmt.Sprintf("%x", sha256.Sum256([]byte(historyKey(path))))[:24]
	signature := fmt.Sprintf("%s:%d:%d:%.3f", path, info.Size(), info.ModTime().UnixNano(), position)
	key := fmt.Sprintf("%x", sha256.Sum256([]byte(signature)))[:24]
	destination := filepath.Join(cacheDir, prefix+"-"+key+".jpg")
	if b, err := os.ReadFile(destination); err == nil {
		return b, nil
	}
	if err = os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, err
	}
	temp, err := os.MkdirTemp(cacheDir, "frame-")
	if err != nil {
		return nil, err
	}
	defer removeThumbnailTemp(cacheDir, temp)
	mpv, err := findMPV()
	if err != nil {
		return nil, err
	}
	jobCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(jobCtx, mpv, "--no-config", "--no-terminal", "--no-audio", "--sid=no", "--no-osc", "--osd-level=0", "--vo=image", "--vo-image-format=jpg", "--vo-image-outdir="+temp, "--frames=1", "--hr-seek=yes", fmt.Sprintf("--start=%.3f", position), "--vf=lavfi=[scale=480:-2]", "--", path)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("thumbnail: %w: %s", err, output)
	}
	files, _ := filepath.Glob(filepath.Join(temp, "*.jpg"))
	if len(files) == 0 {
		return nil, fmt.Errorf("no video frame")
	}
	data, err := os.ReadFile(files[0])
	if err != nil {
		return nil, err
	}
	if err = os.Rename(files[0], destination); err != nil {
		return nil, err
	}
	// One cached frame per source file, even as its saved position changes.
	old, _ := filepath.Glob(filepath.Join(cacheDir, prefix+"-*.jpg"))
	for _, file := range old {
		if file != destination {
			_ = os.Remove(file)
		}
	}
	return data, nil
}

func removeThumbnailTemp(cacheDir, temp string) {
	root, err := filepath.Abs(cacheDir)
	if err != nil {
		return
	}
	target, err := filepath.Abs(temp)
	if err != nil {
		return
	}
	rel, err := filepath.Rel(root, target)
	if err == nil && rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && strings.HasPrefix(filepath.Base(target), "frame-") {
		_ = os.RemoveAll(target)
	}
}
