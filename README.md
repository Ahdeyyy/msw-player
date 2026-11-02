# mswplayr

A native Windows video player built with Go, MyGo, and mpv. The interface takes its typography and quiet light/dark palette from MiniMuse.

The app icon is **Cadence / forward**, selected from twenty distinct concepts through two critiques by Claude Opus 5.5. The [design record](docs/icon-design/README.md) includes the drafts, shortlist, critique, and final assets. The SVG master, 1024px PNG, native-size PNGs, and Windows ICO are in `resources/`.

Add video folders to build your library, search or filter by source folder, then play a video. Settings manages saved sources and appearance. Subfolders are scanned automatically; unavailable sources are reported without removing them. Refresh rescans for new files. You can also drop folders or individual videos from Explorer, or pass paths on the command line.

The player keeps only the timeline and transport controls visible. The queue opens as a modal, captions open as an exact track selection menu, and codec/resolution details live in **Stats for nerds**. Fullscreen shows only the video; use the keyboard to control playback and Escape to return.

Click a library video once to select it; click it again to open it. Its thumbnail expands into the player. Watched videos resume at their saved position, while unwatched videos start at the beginning after the thumbnail fades. The transition respects the system's reduced-motion preference.

Thumbnails use the opening frame for unwatched videos and the saved playback position for watched videos. A single background mpv worker extracts frames for visible rows, with a bounded in-memory cache and one disk thumbnail per source file. Unreadable or unsupported files retain a simple video placeholder. Returning to the library pauses playback and saves the current position.

## Run

With Go 1.27.3 and 7-Zip installed:

```powershell
./scripts/setup-mpv.ps1
go run .
# Or build a portable folder:
./scripts/build.ps1
./build/mswplayr.exe
```

Keep `mpv.exe` and `d3dcompiler_43.dll` beside the packaged player. During development, the player also finds `runtime/mpv.exe`, an executable on PATH, or the path in `MSWPLAYR_MPV`.

## Controls

| Key | Action |
| --- | --- |
| Ctrl+O | Add video folder |
| Space | Play / pause |
| Left / Right | Seek five seconds |
| M | Mute |
| F | Fullscreen |
| Escape | Leave fullscreen |

The audio button cycles audio tracks. The caption menu lists each subtitle track by title, language, and ID, including external/forced/SDH indicators where available. Choose **Off** to disable captions or **Load subtitle file** for SRT, ASS, SSA, VTT, SUB, and IDX files. mpv also discovers matching external subtitle files. The speed button cycles 1, 1.25, 1.5, 2, and 0.75 times normal speed.

Subtitles are found, downloaded, cached, and loaded entirely inside the app. Automatic loading is enabled by default: the player uses an existing track in your preferred language first, then a previously downloaded subtitle, then searches online. Settings lets you change the language or turn automatic downloads off. Choosing **Off** or a local track stops any pending automatic request for that video.

**Find subtitles online** opens an in-app results dialog. The title is derived from the filename, including year or season/episode. **Find and load best** downloads and selects the strongest release match; each result can also be downloaded individually. Ambiguous titles are shown for selection instead of guessing a different film. Weak release matches remain available for manual selection. There is no browser download step.

The default provider is the public [OpenSubtitles v3 add-on](https://opensubtitles-v3.strem.io/manifest.json), with [Cinemeta](https://v3-cinemeta.strem.io/manifest.json) for title identification, through the documented [Stremio subtitle protocol](https://github.com/Stremio/stremio-addon-sdk/blob/master/docs/api/requests/defineSubtitlesHandler.md). No API key is required for this public service at present. Availability and rate limits depend on those services. Errors and empty results are reported in the app; no repeated automatic retry loop runs.

Matching sends the search title, the video's basename, file size, and an OpenSubtitles file hash when available; video contents and full local paths are not uploaded. Cached downloads live in `%APPDATA%/mswplayr/subtitles` (or under `MSWPLAYR_DATA`), so reopening the same unchanged file and language can work offline. Downloads have size limits, support raw/ZIP/gzip subtitle data, are checked for actual subtitle cues, and are written using app-generated filenames. Changing videos, changing language, closing the dialog, or cancelling a request prevents stale results from loading into the player.

## Playback and performance

MKV, MOV, and MP4 are containers. Their actual video and audio codecs determine decoding support. mpv's FFmpeg backend handles common H.264, HEVC, VP9, AV1, AAC, and Opus media, among others. Encrypted/DRM media is outside this player's scope.

Video renders directly into an embedded native window through mpv's GPU renderer, without copying decoded frames through Go or the UI. Hardware decoding is requested with software fallback. Demux buffers are bounded to 64 MiB forward and 16 MiB backward. The UI observes state at most four times per second and redraws only when state changes; the queue builds visible rows only. These are implementation choices, not a measured guarantee of 4K performance on every GPU.

This version targets Windows x64. Folder sources, appearance, density, and per-video playback history are saved to `%APPDATA%/mswplayr/settings.json`; set `MSWPLAYR_DATA` to override that directory. Settings offers Compact, Comfortable, and Spacious layouts. Scrollable lists reserve a 16-DIP gutter beside their contents. The dark theme uses neutral black and gray surfaces and controls.

Positions are checkpointed every ten seconds during playback and on pause, navigation, video changes, and normal exit. Finished videos restart from the beginning. Thumbnails are stored in the `thumbnails` subdirectory of the settings directory. Scans and settings writes run in background workers. Removing a source changes the library only. The playback queue lasts for the session.

## Verification

```powershell
$env:CGO_ENABLED='0'
go test ./...
go tool mygo vet .
go vet ./...
# Requires ffmpeg and the mpv runtime:
$env:MSWPLAYR_INTEGRATION='1'
go test -run Integration -v
# Optional live subtitle search/download (also loads it into mpv if integration is enabled):
$env:MSWPLAYR_SUBTITLE_LIVE='1'
go test -run TestSubtitleProviderLive -v
# Optional UI renders in build/:
$env:MSWPLAYR_RENDER='1'
go test -run TestRender
```

Integration tests generate H.264 MP4/MKV/MOV, HEVC MKV, VP9 WebM, and AV1 MP4 samples with audio and verify loading, timeline progression, pause, seek, and resume. A separate multitrack MKV test verifies exact subtitle selection, disabling captions, and external subtitle loading. UI and filesystem tests cover fullscreen, modal transitions, caption selection, source persistence, filtering, duplicate sources, and cancelled scans. These tests do not measure frame drops, subjective audio synchronization, or HDR correctness.

## Dependencies

The setup script pins the [20261102 Windows mpv build](https://github.com/shinchiro/mpv-winbuild-cmake/releases/tag/20261102) and checks the release asset SHA-256 digest when supplied by GitHub. mpv is a separate process controlled through its local named-pipe JSON IPC. See the [mpv manual](https://mpv.io/manual/stable/) for rendering, decoding, and IPC options.

Martian Grotesk is distributed under the bundled [SIL Open Font License](resources/fonts/OFL.txt). MyGo and go-winio retain their upstream licenses. The downloaded mpv/FFmpeg build has its own GPL and third-party licensing obligations; consult its [build repository](https://github.com/shinchiro/mpv-winbuild-cmake), [mpv sources](https://github.com/mpv-player/mpv), and [FFmpeg sources](https://ffmpeg.org/) before redistribution. The generated portable folder is a local development build, not a completed redistribution package.
