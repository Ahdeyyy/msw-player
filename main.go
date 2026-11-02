package main

import (
	"fmt"
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"log"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

type app struct {
	window                                          *mygo.Window
	engine                                          *player
	status                                          playback
	queue                                           []string
	current                                         int
	speedIndex                                      int
	showQueue, dark                                 bool
	volume, seek                                    float64
	notice                                          string
	host                                            uintptr
	done                                            chan struct{}
	pollEngine                                      atomic.Pointer[player]
	hostGeometry                                    [4]int
	hostVisible                                     bool
	queueList                                       ui.ListState
	page                                            string
	fullscreen, captionsOpen, onlineOpen, statsOpen bool
	folders                                         []string
	library, filtered                               []videoFile
	folderFilter, search                            string
	libraryList                                     ui.ListState
	scanning                                        bool
	scanGeneration                                  uint64
	scanCancel                                      func()
	libraryNotice                                   string
	store                                           *settingsStore
	subtitleID                                      string
	subtitleQuery, subtitleLanguage                 string
	density                                         string
	selectedVideo                                   string
	history                                         map[string]bookmark
	lastCheckpoint                                  time.Time
	thumbs                                          map[string]*thumbnail
	thumbJobs                                       chan thumbnailJob
	thumbCancel                                     func()
	thumbClock                                      uint64
	cacheDir                                        string
	transition                                      *openTransition
	minPlaybackGeneration                           uint64
	autoSubtitles                                   bool
	online                                          subtitleDownloadState
	subtitleService                                 *subtitleService
	subtitleRequest                                 uint64
	subtitlesDir                                    string
}

func newApp() *app {
	return &app{current: -1, volume: 80, page: "library", subtitleID: "no", subtitleLanguage: "English", autoSubtitles: true, density: "Comfortable", history: map[string]bookmark{}, thumbs: map[string]*thumbnail{}, done: make(chan struct{})}
}
func (a *app) send(args ...any) {
	if a.engine != nil {
		a.engine.command(args...)
	}
}
func (a *app) toggle() {
	if a.transition != nil {
		return
	}
	if a.status.Ended {
		a.send("seek", 0, "absolute")
		a.send("set_property", "pause", false)
	} else {
		a.send("cycle", "pause")
	}
}
func (a *app) jump(seconds float64) {
	if a.transition != nil {
		return
	}
	a.send("seek", seconds, "relative")
}
func (a *app) play(index int) {
	if index < 0 || index >= len(a.queue) {
		return
	}
	a.checkpoint(true)
	a.cancelSubtitleJob()
	a.online = subtitleDownloadState{}
	a.current = index
	a.subtitleQuery = videoSearchTitle(a.queue[index])
	a.page = "player"
	a.showQueue = false
	a.captionsOpen = false
	a.notice = ""
	if a.engine == nil || a.engine.snapshot().Error != "" {
		if a.engine != nil {
			a.engine.close()
		}
		a.engine = newPlayer(a.host)
		a.pollEngine.Store(a.engine)
	}
	a.status = playback{}
	a.subtitleID = "no"
	a.engine.prepareLoad()
	a.minPlaybackGeneration = a.engine.snapshot().Generation
	start := a.resumePosition(a.queue[index])
	a.engine.command("loadfile", a.queue[index], "replace", -1, map[string]string{"start": fmt.Sprintf("%.3f", start), "pause": boolOption(a.transition != nil)})
	if a.transition == nil {
		a.engine.command("set_property", "pause", false)
	}
	a.engine.command("set_property", "volume", a.volume)
	a.engine.command("set_property", "speed", a.speed())
	a.layoutHost()
}
func (a *app) add(paths []string) {
	first := len(a.queue)
	for _, path := range paths {
		if supported(path) {
			if abs, err := filepath.Abs(path); err == nil {
				a.queue = append(a.queue, abs)
			}
		}
	}
	if len(a.queue) > first {
		a.play(first)
	} else if len(paths) > 0 {
		a.notice = "Choose a video file such as MKV, MP4, MOV, or WebM."
	}
}
func (a *app) browse() {
	if a.window == nil {
		return
	}
	go func() {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{Parent: a.window, Title: "Open videos", Multiple: true, Filters: []mygo.FileFilter{{Name: "Video", Extensions: videoExtensions}}})
		a.window.Update(func() {
			if err != nil {
				a.notice = err.Error()
			} else {
				a.add(paths)
			}
		})
	}()
}
func main() {
	a := newApp()
	base, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}
	if custom := os.Getenv("MSWPLAYR_DATA"); custom != "" {
		base = custom
	} else {
		base = filepath.Join(base, "mswplayr")
	}
	prefs, err := loadSettings(filepath.Join(base, "settings.json"))
	if err != nil {
		a.libraryNotice = "Settings could not be read: " + err.Error()
	}
	a.folders, a.dark = prefs.Folders, prefs.Dark
	if prefs.Density == "Compact" || prefs.Density == "Spacious" {
		a.density = prefs.Density
	}
	if prefs.History != nil {
		a.history = prefs.History
	}
	a.cacheDir = filepath.Join(base, "thumbnails")
	a.subtitlesDir = filepath.Join(base, "subtitles")
	if prefs.AutoSubtitles != nil {
		a.autoSubtitles = *prefs.AutoSubtitles
	}
	if _, ok := languageCodes[prefs.SubtitleLanguage]; ok {
		a.subtitleLanguage = prefs.SubtitleLanguage
	}
	a.store = newSettingsStore(filepath.Join(base, "settings.json"))
	mygo.App.WhenReady(func() {
		a.window = mygo.NewWindow(mygo.WindowOptions{Title: "mswplayr", Width: 1120, Height: 760, MinWidth: 760, MinHeight: 540, StateKey: "main", Frameless: true, BackgroundColor: "#f1f3f8", Content: ui.View(a.view)})
		a.host = makeHost(a.window.NativeHandle())
		a.startThumbnailWorker()
		a.window.OnResize(a.layoutHost)
		a.window.OnEnterFullScreen(func() { a.fullscreen = true; a.window.Invalidate(); a.layoutHost() })
		a.window.OnLeaveFullScreen(func() { a.fullscreen = false; a.window.Invalidate(); a.layoutHost() })
		a.window.OnFileDrop(func(e *mygo.FileDropEvent) { a.importPaths(e.Paths) })
		a.window.OnClosed(func() {
			a.cancelSubtitleJob()
			a.checkpoint(true)
			close(a.done)
			if a.thumbCancel != nil {
				a.thumbCancel()
			}
			if a.engine != nil {
				a.engine.close()
			}
			if a.scanCancel != nil {
				a.scanCancel()
			}
			if a.store != nil {
				a.store.close()
			}
		})
		go func() {
			ticker := time.NewTicker(250 * time.Millisecond)
			defer ticker.Stop()
			var last playback
			var previous *player
			for {
				select {
				case <-a.done:
					return
				case <-ticker.C:
					engine := a.pollEngine.Load()
					if engine == nil {
						continue
					}
					s := engine.snapshot()
					if engine == previous && s == last {
						continue
					}
					previous, last = engine, s
					a.window.Update(func() {
						if a.engine != engine || s.Generation < a.minPlaybackGeneration {
							return
						}
						a.status = s
						a.subtitleTick()
						a.seek = s.Position
						a.checkpoint(false)
						a.subtitleID = s.SubtitleID
						if a.subtitleID == "" {
							a.subtitleID = "no"
						}
						if s.Error != "" {
							a.notice = s.Error
						}
						if s.Ended && a.current+1 < len(a.queue) {
							a.play(a.current + 1)
						}
						a.layoutHost()
					})
				}
			}
		}()
		if len(os.Args) > 1 {
			a.importPaths(os.Args[1:])
		}
		a.rescan()
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
