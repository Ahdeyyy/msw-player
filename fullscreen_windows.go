package main

import (
	"syscall"
	"time"
	"unsafe"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

type nativePoint struct{ X, Y int32 }

// Poll the native cursor because mpv's child HWND receives video mouse input.
// This runs on the UI thread, including while paused snapshots are unchanged.
func (a *app) tickFullscreenControls() {
	if a.window == nil {
		return
	}
	// Outside fullscreen the caption and audio menus are the main window's
	// popovers: only the fullscreen windows go, never the menus' state.
	if !a.fullscreen || a.page != "player" {
		a.hideFullscreenControls()
		a.controlsCursorKnown = false
		return
	}
	active := a.window.IsFocused() || (a.controlsWindow != nil && a.controlsWindow.IsFocused()) || (a.pickerWindow != nil && a.pickerWindow.IsFocused())
	if a.pickerVisible && a.window.IsFocused() {
		a.dismissFullscreenPicker()
	}
	if !active {
		a.dismissFullscreenPicker()
		a.hideFullscreenControls()
		a.controlsCursorKnown = false
		return
	}
	var point nativePoint
	if ok, _, _ := user32.NewProc("GetCursorPos").Call(uintptr(unsafe.Pointer(&point))); ok == 0 {
		return
	}
	local := point
	user32.NewProc("ScreenToClient").Call(a.window.NativeHandle(), uintptr(unsafe.Pointer(&local)))
	var client struct{ Left, Top, Right, Bottom int32 }
	getClient.Call(a.window.NativeHandle(), uintptr(unsafe.Pointer(&client)))
	inside := local.X >= 0 && local.Y >= 0 && local.X < client.Right && local.Y < client.Bottom
	now := time.Now()
	if inside && (!a.controlsCursorKnown || point != a.controlsCursor) {
		a.controlsUntil = now.Add(3 * time.Second)
	}
	a.controlsCursor, a.controlsCursorKnown = point, true
	if a.controlsVisible && a.controlsWindow != nil {
		if captured, _, _ := user32.NewProc("GetCapture").Call(); captured != 0 {
			a.controlsUntil = now.Add(3 * time.Second)
		}
		p := point
		user32.NewProc("ScreenToClient").Call(a.controlsWindow.NativeHandle(), uintptr(unsafe.Pointer(&p)))
		var r struct{ Left, Top, Right, Bottom int32 }
		getClient.Call(a.controlsWindow.NativeHandle(), uintptr(unsafe.Pointer(&r)))
		if p.X >= 0 && p.Y >= 0 && p.X < r.Right && p.Y < r.Bottom {
			a.controlsUntil = now.Add(3 * time.Second)
		}
	}
	if now.After(a.controlsUntil) && !a.captionsOpen && !a.audioOpen {
		a.hideFullscreenControls()
		return
	}
	if a.controlsWindow == nil {
		a.controlsWindow = mygo.NewWindow(mygo.WindowOptions{
			Title: "Playback controls", Parent: a.window, Hidden: true,
			Frameless: true, Transparent: true, Vibrancy: mygo.VibrancyAcrylic,
			CornerRadius: 16, SkipTaskbar: true, DisableResize: true,
			Width: 720, Height: 112,
			Content: ui.View(a.fullscreenControlsView),
		})
		a.controlsWindow.OnClosed(func() { a.controlsWindow = nil; a.controlsVisible = false })
		frostControlsWindow(a.controlsWindow.NativeHandle())
	}
	bounds := a.window.ContentBounds()
	target := fullscreenControlsBounds(bounds)
	if a.controlsWindow.Bounds() != target {
		a.controlsWindow.SetBounds(target)
		a.controlsWindow.Invalidate()
	}
	if !a.controlsVisible {
		a.controlsWindow.ShowInactive()
		a.controlsVisible = true
	}
	a.layoutFullscreenPicker(target, bounds)
}

// hideFullscreenControls hides the fullscreen windows. It leaves the menus'
// state alone; dismissFullscreenPicker closes the menus.
func (a *app) hideFullscreenControls() {
	a.hidePickerWindow()
	if a.controlsWindow != nil && a.controlsVisible {
		a.controlsWindow.Hide()
	}
	a.controlsVisible = false
}

func (a *app) closeFullscreenControls() {
	if a.pickerWindow != nil {
		a.pickerWindow.Close()
	}
	if a.controlsWindow != nil {
		a.controlsWindow.Close()
	}
}

func (a *app) fullscreenControlsView(c *ui.Context) {
	t := ui.DarkTheme()
	t.Font = "Martian Grotesk"
	c.SetTheme(t)
	c.Root().Background(ui.Transparent)
	c.OnShortcut(0, ui.KeyEscape, a.exitFullscreen)
	c.OnShortcut(0, ui.KeyF, a.toggleFullscreen)
	c.OnShortcut(0, ui.KeySpace, a.toggle)
	c.OnShortcut(0, ui.KeyM, func() { a.send("cycle", "mute") })
	c.OnShortcut(ui.Ctrl, ui.KeyO, a.browseFolder)
	c.OnShortcut(0, ui.KeyLeft, func() { a.jump(-5) })
	c.OnShortcut(0, ui.KeyRight, func() { a.jump(5) })
	c.OnShortcut(0, ui.KeyPageUp, func() { a.stepChapter(-1) })
	c.OnShortcut(0, ui.KeyPageDown, func() { a.stepChapter(1) })
	ui.Column(c).Fill().Children(func() {
		ui.Spacer(c)
		ui.Column(c).FillWidth().Height(112).Shrink(0).Background(ui.RGBA(20, 22, 28, .08)).Children(func() { a.playbackControls(c) })
	})
}

// The native acrylic window always stays at the control bar's height.
func fullscreenControlsBounds(bounds mygo.Rectangle) mygo.Rectangle {
	width := min(1000, bounds.Width-48)
	return mygo.Rectangle{X: bounds.X + (bounds.Width-width)/2, Y: bounds.Y + bounds.Height - 132, Width: width, Height: 112}
}

func (a *app) dismissFullscreenPicker() {
	a.captionsOpen, a.audioOpen = false, false
	a.hidePickerWindow()
}

// hidePickerWindow hides the picker and forgets what it showed, so that
// opening it again measures and paints the current menu before it shows.
func (a *app) hidePickerWindow() {
	if a.pickerWindow != nil && a.pickerVisible {
		a.pickerWindow.Hide()
	}
	a.pickerVisible = false
	a.pickerKind = ""
}

// menuKind is the track menu that is open, if any.
func (a *app) menuKind() string {
	switch {
	case a.audioOpen:
		return "audio"
	case a.captionsOpen:
		return "captions"
	}
	return ""
}

func (a *app) layoutFullscreenPicker(bar, screen mygo.Rectangle) {
	kind := a.menuKind()
	if kind == "" {
		a.hidePickerWindow()
		return
	}
	if a.pickerWindow == nil {
		a.pickerWindow = mygo.NewWindow(mygo.WindowOptions{
			Title: "Playback picker", Parent: a.controlsWindow, Hidden: true,
			Frameless: true, Transparent: true, Vibrancy: mygo.VibrancyAcrylic,
			SkipTaskbar: true, DisableResize: true, CornerRadius: 8,
			Width: 328, Height: 368,
			Content: ui.View(a.fullscreenPickerView),
		})
		a.pickerWindow.OnClosed(func() { a.pickerWindow = nil; a.pickerVisible = false })
		frostControlsWindow(a.pickerWindow.NativeHandle())
	}
	// A menu opened, or switched, is measured and painted while the window
	// is hidden. Shown at once, the window would show its last frame (the
	// previous menu or selection) and its old height for a frame or more.
	if kind != a.pickerKind {
		if a.pickerVisible {
			a.pickerWindow.Hide()
			a.pickerVisible = false
		}
		a.pickerKind, a.pickerSize = kind, ui.Rect{}
		a.pickerSerial++
		a.pickerOpened = time.Now()
		a.pickerWindow.Invalidate()
	}
	width, height := 308, 368
	if kind == "audio" {
		width = 328
	}
	if a.pickerSize.H > 0 {
		height = int(a.pickerSize.H + 0.999)
	}
	height = min(height, bar.Y-screen.Y-8)
	target := mygo.Rectangle{X: bar.X + bar.Width - width - 60, Y: bar.Y - height - 8, Width: width, Height: height}
	resized := a.pickerWindow.Bounds() != target
	if resized {
		a.pickerWindow.SetBounds(target)
		a.pickerWindow.Invalidate()
	}
	// Show once measured and painted at its final size. Should a hidden
	// window not draw, it shows anyway after a moment.
	ready := a.pickerReady == a.pickerSerial && !resized
	if !a.pickerVisible && (ready || time.Since(a.pickerOpened) > 250*time.Millisecond) {
		a.pickerWindow.Show()
		a.pickerVisible = true
	}
}

func (a *app) fullscreenPickerView(c *ui.Context) {
	t := ui.DarkTheme()
	t.Font = "Martian Grotesk"
	c.SetTheme(t)
	// The picker frosts the video like the bar under it; without a system
	// backdrop it keeps the opaque fill.
	fill := t.Background
	if c.Vibrancy() {
		fill = ui.RGBA(20, 22, 28, .4)
	}
	c.Root().Background(ui.Transparent)
	c.OnShortcut(0, ui.KeyEscape, a.dismissFullscreenPicker)
	// The content follows the kind the window was laid out for, never the
	// flags, which can change between the ticks that place the window.
	serial := a.pickerSerial
	ui.Column(c.Key(serial)).Bind(&a.pickerContent).Padding(4).Background(fill).Radius(8).Border(1, ui.RGBA(255, 255, 255, .1)).Children(func() {
		switch a.pickerKind {
		case "captions":
			a.captionPicker(c)
		case "audio":
			a.audioPicker(c)
		}
	})
	if a.pickerKind == "" {
		return
	}
	// Bounds are the last frame's: the first build of a menu has none of its
	// own, and the size is the menu's once two builds agree.
	size := a.pickerContent.Bounds(c)
	if a.pickerMeasured != serial {
		a.pickerMeasured = serial
		a.pickerSize = ui.Rect{}
		c.After(16 * time.Millisecond)
		return
	}
	if size != a.pickerSize || size.H <= 0 {
		a.pickerSize = size
		c.After(16 * time.Millisecond)
		return
	}
	if a.pickerReady != serial {
		a.pickerReady = serial
		// Place and show it now rather than at the next tick.
		if a.window != nil {
			a.window.Update(a.tickFullscreenControls)
		}
	}
}

// Use a lightly tinted acrylic blur over the live video. The system backdrop
// otherwise becomes an opaque inactive-window fill when shown without focus.
// Keep MyGo's standard acrylic if this optional Windows accent is unavailable.
func frostControlsWindow(hwnd uintptr) {
	accentProc := user32.NewProc("SetWindowCompositionAttribute")
	if accentProc.Find() != nil {
		return
	}
	type accentPolicy struct{ State, Flags, Color, Animation uint32 }
	type compositionAttribute struct {
		Attribute uint32
		Data      unsafe.Pointer
		Size      uintptr
	}
	policy := accentPolicy{State: 4, Flags: 2, Color: 0x381c1814}
	attribute := compositionAttribute{Attribute: 19, Data: unsafe.Pointer(&policy), Size: unsafe.Sizeof(policy)}
	dwm := syscall.NewLazyDLL("dwmapi.dll").NewProc("DwmSetWindowAttribute")
	backdrop := uint32(1) // DWMSBT_NONE: do not stack the default tint over ours.
	dwm.Call(hwnd, 38, uintptr(unsafe.Pointer(&backdrop)), unsafe.Sizeof(backdrop))
	if ok, _, _ := accentProc.Call(hwnd, uintptr(unsafe.Pointer(&attribute))); ok == 0 {
		backdrop = 3 // DWMSBT_TRANSIENTWINDOW: standard system acrylic fallback.
		dwm.Call(hwnd, 38, uintptr(unsafe.Pointer(&backdrop)), unsafe.Sizeof(backdrop))
	}
}
