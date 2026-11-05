package main

import (
	"path/filepath"
	"unsafe"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// pipBar is the height of the picture-in-picture control strip, in DIPs.
const pipBar = 44

var setParent = user32.NewProc("SetParent")

// playing is whether the current video is running, not paused or opening.
func (a *app) playing() bool {
	return a.current >= 0 && a.status.Loaded && !a.status.Paused && !a.status.Ended && a.transition == nil && a.notice == ""
}

// togglePip moves the video between the player and a small window that stays
// above other windows, leaving the library open behind it.
func (a *app) togglePip() {
	if a.pip {
		a.showPlayer()
		return
	}
	if a.current < 0 || !a.status.Loaded {
		return
	}
	if a.fullscreen {
		a.exitFullscreen()
	}
	a.pip = true
	a.navigate("library")
}

// showPlayer brings the video back from picture-in-picture into the window.
func (a *app) showPlayer() {
	a.navigate("player")
	if a.window != nil {
		if a.window.IsMinimized() {
			a.window.Restore()
		}
		a.window.Focus()
	}
}

// closePip stops picture-in-picture, pausing the video it showed.
func (a *app) closePip() {
	if !a.pip {
		return
	}
	a.checkpoint(true)
	a.send("set_property", "pause", true)
	a.pip = false
	a.layoutHost()
}

// adoptHost moves the video window into parent, hidden until it is laid out
// there. A window destroyed with the video still inside takes mpv's with it.
func (a *app) adoptHost(parent uintptr) {
	if a.host == 0 || parent == 0 || a.hostParent == parent {
		return
	}
	showWindow.Call(a.host, 0)
	a.hostVisible = false
	setParent.Call(a.host, parent)
	a.hostParent = parent
	a.hostGeometry, a.hostHole = [4]int{}, [4]int{}
}

// pipBounds puts a 16:9 window in the bottom-right corner of the work area
// of the display the main window is on.
func pipBounds(window mygo.Rectangle, displays []mygo.Display) mygo.Rectangle {
	area := mygo.Rectangle{Width: 1280, Height: 720}
	cx, cy := window.X+window.Width/2, window.Y+window.Height/2
	for i, d := range displays {
		w := d.WorkArea
		if i == 0 || d.Primary {
			area = w
		}
		if cx >= w.X && cy >= w.Y && cx < w.X+w.Width && cy < w.Y+w.Height {
			area = w
			break
		}
	}
	width := max(320, min(480, area.Width/4))
	height := width*9/16 + pipBar
	return mygo.Rectangle{X: area.X + area.Width - width - 24, Y: area.Y + area.Height - height - 24, Width: width, Height: height}
}

func (a *app) layoutPip() {
	if a.pipWindow == nil {
		b := pipBounds(a.window.Bounds(), mygo.Screen.Displays())
		a.pipWindow = mygo.NewWindow(mygo.WindowOptions{
			Title: "mswplayr picture in picture", Hidden: true,
			Frameless: true, AlwaysOnTop: true, SkipTaskbar: true,
			DisableMaximize: true, DisableFullScreen: true,
			X: b.X, Y: b.Y, Width: b.Width, Height: b.Height, MinWidth: 240, MinHeight: 135 + pipBar,
			BackgroundColor: "#000000", Content: ui.View(a.pipView),
		})
		a.pipWindow.OnResize(a.layoutHost)
		// Alt+F4 closes the window: take the video out first.
		a.pipWindow.OnClose(func(e *mygo.CloseEvent) {
			a.closePip()
			a.adoptHost(a.window.NativeHandle())
		})
		a.pipWindow.OnClosed(func() { a.pipWindow, a.pipVisible = nil, false })
	}
	pip := a.pipWindow.NativeHandle()
	a.adoptHost(pip)
	if !a.pipVisible {
		a.pipWindow.ShowInactive()
		a.pipVisible = true
	}
	a.pipWindow.Invalidate()
	if !a.status.Loaded || a.notice != "" {
		if a.hostVisible {
			showWindow.Call(a.host, 0)
			a.hostVisible = false
		}
		return
	}
	var r struct{ Left, Top, Right, Bottom int32 }
	getClient.Call(pip, uintptr(unsafe.Pointer(&r)))
	dpi, _, _ := getDPI.Call(pip)
	s := float64(dpi) / 96
	if s == 0 {
		s = 1
	}
	geometry := [4]int{0, 0, int(r.Right), max(1, int(r.Bottom)-int(pipBar*s))}
	if geometry == a.hostGeometry && a.hostVisible && !a.hostBehind {
		return
	}
	a.hostGeometry, a.hostBehind, a.hostHole = geometry, false, [4]int{}
	moveWindow.Call(a.host, 0, 0, uintptr(geometry[2]), uintptr(geometry[3]), 1)
	user32.NewProc("SetWindowRgn").Call(a.host, 0, 1)
	showWindow.Call(a.host, 5)
	a.hostVisible = true
	setWindowPos.Call(a.host, 0, 0, 0, 0, 0, 0x0001|0x0002|0x0010)
}

// hidePip hides the picture-in-picture window once the video has left it.
func (a *app) hidePip() {
	if a.pipWindow != nil && a.pipVisible {
		a.pipWindow.Hide()
	}
	a.pipVisible = false
}

func (a *app) pipView(c *ui.Context) {
	t := ui.DarkTheme()
	t.Font = "Martian Grotesk"
	c.SetTheme(t)
	c.Root().Background(ui.Hex("#000000"))
	c.OnShortcut(0, ui.KeySpace, a.toggle)
	c.OnShortcut(0, ui.KeyLeft, func() { a.jump(-5) })
	c.OnShortcut(0, ui.KeyRight, func() { a.jump(5) })
	c.OnShortcut(0, ui.KeyM, func() { a.send("cycle", "mute") })
	c.OnShortcut(0, ui.KeyEscape, a.closePip)
	title := ""
	if a.current >= 0 && a.current < len(a.queue) {
		title = filepath.Base(a.queue[a.current])
	}
	w, _ := c.Size()
	ui.Column(c).Fill().Children(func() {
		// The video window covers this area; it shows while mpv opens a file.
		ui.Column(c).Label("Video player").FillWidth().Grow(1).MinHeight(0).Center().Children(func() {
			if !a.status.Loaded && a.notice == "" {
				ui.Spinner(c).Size(18, 18).TextColor(ui.Hex("#a4afc4"))
			} else if a.notice != "" {
				ui.Text(c, "Playback unavailable").TextColor(ui.Hex("#a4afc4"))
			}
		})
		progress := float32(0)
		if a.status.Duration > 0 {
			progress = float32(min(1, max(0, a.status.Position/a.status.Duration)))
		}
		ui.Box(c).FillWidth().Height(2).Shrink(0).Background(ui.RGBA(255, 255, 255, .14)).Children(func() {
			ui.Box(c).Width(w * progress).FillHeight().Background(t.Accent)
		})
		ui.Row(c).FillWidth().Height(pipBar-2).Shrink(0).Padding(0, 4, 0, 6).Gap(2).AlignItems(ui.Center).DragWindow().Children(func() {
			name, label := "play", "Play"
			if a.status.Loaded && !a.status.Paused && !a.status.Ended {
				name, label = "pause", "Pause"
			}
			keyedIcon(c, "play-pause", name, label).Size(32, 32).Disabled(!a.status.Loaded).OnClick(a.toggle)
			ui.Text(c, title).FontSize(12).SingleLine().Grow(1).MinWidth(0).Padding(0, 6).TextColor(t.TextMuted)
			icon(c, "pip-exit", "Back to player").Size(32, 32).OnClick(a.showPlayer)
			icon(c, "close", "Close picture in picture").Size(32, 32).OnClick(a.closePip)
		})
	})
}
