package main

import (
	"github.com/egoist/mygo/ui"
	"time"
)

type openTransition struct {
	source               ui.Rect
	image                *ui.Bitmap
	started, fadeStarted time.Time
	resume, finishing    bool
	position             float64
}

// The morph grows quickly out of the thumbnail and settles into the player.
// MyGo's animation clock paces its frames, not a timer.
var (
	morphAnimation = []ui.Keyframe{{At: 0, Value: 0}, {At: 360 * time.Millisecond, Value: 1, Ease: easeDrawer}}
	morphFade      = ui.Animation{Duration: 200 * time.Millisecond, Ease: easeOut}
)

func (a *app) selectVideo(c *ui.Context, index int, entry *thumbnail) {
	if index < 0 || index >= len(a.filtered) {
		return
	}
	path := a.filtered[index].Path
	if a.selectedVideo != path {
		a.selectedVideo = path
		return
	}
	source := entry.handle.Bounds(c)
	if source.W > 0 && source.H > 0 && !c.Preferences().ReduceMotion {
		position := a.resumePosition(path)
		a.transition = &openTransition{source: source, image: entry.image, started: c.Now(), resume: position > 0, position: position}
	}
	a.playLibrary(index)
}
func morphRect(source, target ui.Rect, progress float32) ui.Rect {
	p := max(float32(0), min(float32(1), progress))
	return ui.Rect{X: source.X + (target.X-source.X)*p, Y: source.Y + (target.Y-source.Y)*p, W: source.W + (target.W-source.W)*p, H: source.H + (target.H-source.H)*p}
}
func (a *app) transitionView(c *ui.Context) {
	tr := a.transition
	if tr == nil || a.page != "player" || a.fullscreen {
		return
	}
	w, h := c.Size()
	target := ui.Rect{X: 20, Y: 100, W: w - 40, H: h - 204}
	finished := c.Preferences().ReduceMotion
	ui.Overlay(c, func() {
		box := ui.Box(c.Key("opening-thumbnail")).Absolute().Background(ui.Hex("#080808")).Clip()
		progress := box.Keyframes("morph", morphAnimation)
		ready := (a.status.Loaded && (!tr.resume || (a.status.FrameReady && absPosition(a.status.Position-tr.position) < .75))) || a.notice != "" || c.Now().Sub(tr.started) > 8*time.Second
		shown := float32(1)
		if progress >= 1 && ready {
			if !tr.resume && tr.fadeStarted.IsZero() {
				tr.fadeStarted = c.Now()
			}
			shown = 0
		}
		// A resumed thumbnail is the frame mpv shows, so it cuts straight to it.
		opacity := shown
		if !tr.resume {
			opacity = box.AnimateTo("fade", shown, morphFade)
		}
		if opacity <= 0.001 {
			finished = true
		}
		// Corners go from the thumbnail's 6 to the viewport's 12.
		rect := morphRect(tr.source, target, progress)
		box.Left(rect.X).Top(rect.Y).Size(rect.W, rect.H).Radius(6 + 6*progress).Opacity(opacity).Children(func() {
			if tr.image != nil {
				ui.Image(c, tr.image).Fill().Fit(ui.Contain)
			} else {
				ui.Column(c).Fill().Center().Children(func() { ui.Icon(c, icons["video"]).Size(32, 32).TextColor(c.Theme().TextMuted) })
			}
		})
		if progress >= 1 && !ready {
			c.After(100 * time.Millisecond)
		}
	})
	if finished && !tr.finishing {
		tr.finishing = true
		finish := func() {
			if a.transition != tr {
				return
			}
			a.transition = nil
			if a.notice == "" {
				a.send("set_property", "pause", false)
			}
			a.layoutHost()
		}
		if a.window != nil {
			a.window.Update(finish)
		} else {
			finish()
		}
	}
}
func absPosition(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
