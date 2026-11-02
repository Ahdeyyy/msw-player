package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Microsoft/go-winio"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

type player struct {
	mu       sync.Mutex
	state    playback
	commands chan []any
	cancel   context.CancelFunc
	input    chan string
}

func findMPV() (string, error) {
	if path := os.Getenv("MSWPLAYR_MPV"); path != "" {
		return path, nil
	}
	exe, _ := os.Executable()
	for _, path := range []string{filepath.Join(filepath.Dir(exe), "mpv.exe"), "runtime/mpv.exe"} {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return filepath.Abs(path)
		}
	}
	path, err := exec.LookPath("mpv.exe")
	if err != nil {
		return "", fmt.Errorf("mpv is missing. Run scripts/setup-mpv.ps1, then reopen the video.")
	}
	return path, nil
}
func newPlayer(host uintptr) *player {
	ctx, cancel := context.WithCancel(context.Background())
	p := &player{commands: make(chan []any, 128), cancel: cancel, input: make(chan string, 32)}
	go p.run(ctx, host)
	return p
}
func (p *player) snapshot() playback { p.mu.Lock(); defer p.mu.Unlock(); return p.state }
func (p *player) prepareLoad() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state.Generation++
	p.state.Loaded = false
	p.state.Ended = false
	p.state.Position = 0
	p.state.Error = ""
}
func (p *player) fail(err error) {
	p.mu.Lock()
	p.state.Error = err.Error()
	p.state.Loaded = false
	p.mu.Unlock()
}
func (p *player) command(args ...any) {
	select {
	case p.commands <- args:
	default:
		p.fail(fmt.Errorf("Playback is busy. Try again."))
	}
}
func (p *player) close() { p.cancel() }
func (p *player) run(ctx context.Context, host uintptr) {
	defer p.cancel()
	path, err := findMPV()
	if err != nil {
		p.fail(err)
		return
	}
	pipe := fmt.Sprintf(`\\.\pipe\mswplayr-%d-%d`, os.Getpid(), time.Now().UnixNano())
	bindings, err := os.CreateTemp("", "mswplayr-input-*.conf")
	if err != nil {
		p.fail(err)
		return
	}
	defer os.Remove(bindings.Name())
	if _, err = bindings.WriteString(playerInputBindings); err != nil {
		bindings.Close()
		p.fail(err)
		return
	}
	bindings.Close()
	cmd := exec.CommandContext(ctx, path, "--no-config", "--idle=yes", "--keep-open=yes", "--force-window=yes", fmt.Sprintf("--wid=%d", host), "--input-ipc-server="+pipe, "--input-conf="+bindings.Name(), "--hwdec=auto-safe", "--vo=gpu-next", "--gpu-api=d3d11", "--osc=no", "--osd-level=0", "--input-default-bindings=no", "--input-vo-keyboard=yes", "--terminal=no", "--volume-max=100", "--demuxer-max-bytes=64MiB", "--demuxer-max-back-bytes=16MiB")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err = cmd.Start(); err != nil {
		p.fail(err)
		return
	}
	defer func() { _ = cmd.Process.Kill(); _ = cmd.Wait() }()
	dialCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	var conn net.Conn
	for dialCtx.Err() == nil {
		timeout := 200 * time.Millisecond
		conn, err = winio.DialPipe(pipe, &timeout)
		if err == nil {
			break
		}
		select {
		case <-dialCtx.Done():
		case <-time.After(50 * time.Millisecond):
		}
	}
	if err != nil || conn == nil {
		p.fail(fmt.Errorf("Could not connect to mpv: %v", err))
		return
	}
	defer conn.Close()
	go func() { <-ctx.Done(); conn.Close() }()
	enc := json.NewEncoder(conn)
	for id, name := range []string{"time-pos", "duration", "pause", "eof-reached", "video-codec", "hwdec-current", "width", "height", "idle-active", "track-list", "sid", "container-fps", "aid", "chapter-list", "chapter"} {
		if err = enc.Encode(map[string]any{"command": []any{"observe_property", id + 1, name}}); err != nil {
			p.fail(err)
			return
		}
	}
	readDone := make(chan struct{})
	go func() {
		defer close(readDone)
		scanner := bufio.NewScanner(conn)
		scanner.Buffer(make([]byte, 4096), 1024*1024)
		for scanner.Scan() {
			p.event(scanner.Bytes())
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case <-readDone:
			if ctx.Err() == nil {
				p.fail(fmt.Errorf("Playback engine disconnected. Reopen the player."))
			}
			return
		case args := <-p.commands:
			if len(args) > 0 && args[0] == "loadfile" {
				p.prepareLoad()
			}
			if err = enc.Encode(map[string]any{"command": args}); err != nil {
				p.fail(err)
				return
			}
		}
	}
}
func (p *player) event(data []byte) {
	var e struct {
		Event, Name, Error, Reason string
		Data                       json.RawMessage
		Args                       []string `json:"args"`
	}
	if json.Unmarshal(data, &e) != nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if e.Event == "client-message" && len(e.Args) == 2 && e.Args[0] == "mswplayr" {
		select {
		case p.input <- e.Args[1]:
			p.state.InputSerial++
		default:
		}
		return
	}
	if e.Event == "end-file" && e.Reason == "error" {
		p.state.Error = "This video could not be played. Check that the file is readable and complete."
		p.state.Loaded = false
	}
	if e.Event == "file-loaded" {
		p.state.Loaded = true
		p.state.Error = ""
	}
	if e.Event != "property-change" {
		return
	}
	switch e.Name {
	case "aid":
		p.state.AudioID = trackID(e.Data)
	case "chapter-list":
		p.state.ChapterJSON = string(e.Data)
	case "chapter":
		p.state.Chapter = -1
		json.Unmarshal(e.Data, &p.state.Chapter)
	case "container-fps":
		json.Unmarshal(e.Data, &p.state.FPS)
	case "track-list":
		p.state.TrackJSON = string(e.Data)
	case "sid":
		p.state.SubtitleID = trackID(e.Data)
	case "time-pos":
		json.Unmarshal(e.Data, &p.state.Position)
	case "duration":
		json.Unmarshal(e.Data, &p.state.Duration)
	case "pause":
		json.Unmarshal(e.Data, &p.state.Paused)
	case "eof-reached":
		json.Unmarshal(e.Data, &p.state.Ended)
	case "video-codec":
		json.Unmarshal(e.Data, &p.state.Codec)
	case "hwdec-current":
		json.Unmarshal(e.Data, &p.state.Decoder)
	case "width":
		json.Unmarshal(e.Data, &p.state.Width)
	case "height":
		json.Unmarshal(e.Data, &p.state.Height)
	}
}

var user32 = syscall.NewLazyDLL("user32.dll")
var createWindow = user32.NewProc("CreateWindowExW")
var moveWindow = user32.NewProc("MoveWindow")
var setWindowPos = user32.NewProc("SetWindowPos")
var showWindow = user32.NewProc("ShowWindow")
var getClient = user32.NewProc("GetClientRect")
var getDPI = user32.NewProc("GetDpiForWindow")

func makeHost(parent uintptr) uintptr {
	class, _ := syscall.UTF16PtrFromString("STATIC")
	h, _, _ := createWindow.Call(0, uintptr(unsafe.Pointer(class)), 0, 0x40000000|0x04000000|0x02000000, 0, 0, 1, 1, parent, 0, 0, 0)
	return h
}
func (a *app) layoutHost() {
	if a.window == nil || a.host == 0 {
		return
	}
	if !a.status.Loaded || a.notice != "" || a.page != "player" || a.transition != nil || (!a.fullscreen && (a.showQueue || a.captionsOpen || a.onlineOpen || a.statsOpen || a.audioOpen || a.chaptersOpen)) {
		if a.hostVisible {
			showWindow.Call(a.host, 0)
			a.hostVisible = false
		}
		return
	}
	parent := a.window.NativeHandle()
	var r struct{ Left, Top, Right, Bottom int32 }
	getClient.Call(parent, uintptr(unsafe.Pointer(&r)))
	dpi, _, _ := getDPI.Call(parent)
	s := float64(dpi) / 96
	if s == 0 {
		s = 1
	}
	px := func(v float64) uintptr { return uintptr(max(1, int(v*s))) }
	geometry := videoGeometry(int(r.Right), int(r.Bottom), s, a.fullscreen)
	if geometry == a.hostGeometry && a.hostVisible {
		return
	}
	a.hostGeometry = geometry
	moveWindow.Call(a.host, uintptr(geometry[0]), uintptr(geometry[1]), uintptr(geometry[2]), uintptr(geometry[3]), 1)
	if a.fullscreen {
		user32.NewProc("SetWindowRgn").Call(a.host, 0, 1)
	} else {
		region, _, _ := syscall.NewLazyDLL("gdi32.dll").NewProc("CreateRoundRectRgn").Call(0, 0, uintptr(geometry[2]+1), uintptr(geometry[3]+1), px(24), px(24))
		if region != 0 {
			user32.NewProc("SetWindowRgn").Call(a.host, region, 1)
		}
	}
	showWindow.Call(a.host, 5)
	a.hostVisible = true
	// MyGo owns a sibling rendering surface. Keep the video host above it.
	setWindowPos.Call(a.host, 0, 0, 0, 0, 0, 0x0001|0x0002|0x0010)
}
