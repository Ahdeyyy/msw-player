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
	p = p * p * (3 - 2*p)
	return ui.Rect{X: source.X + (target.X-source.X)*p, Y: source.Y + (target.Y-source.Y)*p, W: source.W + (target.W-source.W)*p, H: source.H + (target.H-source.H)*p}
}
func (a *app) transitionView(c *ui.Context) {
	tr := a.transition
	if tr == nil || a.page != "player" || a.fullscreen {
		return
	}
	w, h := c.Size()
	target := ui.Rect{X: 20, Y: 100, W: w - 40, H: h - 204}
	progress := float32(c.Now().Sub(tr.started).Seconds() / .34)
	rect := morphRect(tr.source, target, progress)
	opacity := float32(1)
	ready := (a.status.Loaded && (!tr.resume || absPosition(a.status.Position-tr.position) < .75)) || a.notice != "" || c.Now().Sub(tr.started) > 8*time.Second
	if progress >= 1 && ready {
		if !tr.resume {
			if tr.fadeStarted.IsZero() {
				tr.fadeStarted = c.Now()
			}
			opacity = 1 - float32(c.Now().Sub(tr.fadeStarted).Seconds()/.18)
		} else {
			opacity = 0
		}
	}
	if c.Preferences().ReduceMotion {
		opacity = 0
	}
	if opacity <= 0 {
		if !tr.finishing {
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
		return
	}
	ui.Overlay(c, func() {
		ui.Box(c.Key("opening-thumbnail")).Absolute().Left(rect.X).Top(rect.Y).Size(rect.W, rect.H).Background(ui.Hex("#080808")).Radius(8 + 4*min(progress, 1)).Clip().Opacity(opacity).Children(func() {
			if tr.image != nil {
				ui.Image(c, tr.image).Fill().Fit(ui.Contain)
			} else {
				ui.Column(c).Fill().Center().Children(func() { ui.Icon(c, icons["video"]).Size(32, 32).TextColor(c.Theme().TextMuted) })
			}
		})
	})
	if progress < 1 || !tr.fadeStarted.IsZero() {
		c.After(16 * time.Millisecond)
	} else if !ready {
		c.After(100 * time.Millisecond)
	}
}
func absPosition(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
