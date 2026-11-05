package main

import "testing"

func pausedBy(commands chan []any) bool {
	for {
		select {
		case command := <-commands:
			if command[0] == "set_property" && command[1] == "pause" && command[2] == true {
				return true
			}
		default:
			return false
		}
	}
}

func TestLibraryKeepsPlayingVideoInPip(t *testing.T) {
	a := newApp()
	a.engine = &player{commands: make(chan []any, 64), cancel: func() {}}
	a.page, a.queue, a.current = "player", []string{"A.mkv", "B.mkv"}, 0
	a.status = playback{Loaded: true}
	a.navigate("library")
	if !a.pip || a.page != "library" || pausedBy(a.engine.commands) {
		t.Fatal("a playing video must continue in picture-in-picture")
	}
	a.advance()
	if !a.pip || a.page != "library" || a.current != 1 {
		t.Fatal("the next video must play in picture-in-picture")
	}
	a.status = playback{Loaded: true}
	a.navigate("player")
	if a.pip || a.page != "player" {
		t.Fatal("opening the player must end picture-in-picture")
	}
	a.status.Paused = true
	a.navigate("library")
	if a.pip || !pausedBy(a.engine.commands) {
		t.Fatal("a paused video must stay paused without picture-in-picture")
	}
}

func TestPipToggle(t *testing.T) {
	a := newApp()
	a.engine = &player{commands: make(chan []any, 64), cancel: func() {}}
	a.page, a.queue, a.current = "player", []string{"A.mkv"}, 0
	a.status = playback{Loaded: true, Paused: true}
	a.togglePip()
	if !a.pip || a.page != "library" || pausedBy(a.engine.commands) {
		t.Fatal("the toggle must move even a paused video into picture-in-picture")
	}
	a.closePip()
	if a.pip || !pausedBy(a.engine.commands) {
		t.Fatal("closing picture-in-picture must pause")
	}
	a.togglePip()
	a.play(0)
	if a.pip || a.page != "player" {
		t.Fatal("choosing a video must open it in the player")
	}
}
