package main

import (
	"embed"
	"time"

	"github.com/egoist/mygo/plugins/glass"
	"github.com/egoist/mygo/ui"
)

//go:embed resources/fonts/*.ttf
var fonts embed.FS

//go:embed resources/mark.svg
var brandMarkSource []byte
var brandMark = ui.MustParseSVG(brandMarkSource)

func init() {
	for _, w := range []string{"Regular", "Medium", "Bold"} {
		b, err := fonts.ReadFile("resources/fonts/MartianGrotesk-" + w + ".ttf")
		if err == nil {
			err = ui.RegisterFont(b, "Martian Grotesk")
		}
		if err != nil {
			panic(err)
		}
	}
}
func (a *app) theme() *ui.Theme {
	t := ui.LightTheme()
	t.Font = "Martian Grotesk"
	t.FontSize = 13
	t.Radius = 8
	t.Spacing = 4
	t.Background = ui.Hex("#f1f3f8")
	t.Surface = ui.Hex("#ffffff")
	t.Text = ui.Hex("#20293e")
	t.TextMuted = ui.Hex("#667189")
	t.Accent = ui.Hex("#5260d8")
	t.Selection = ui.Hex("#e8ebfb")
	t.Border = ui.Hex("#dce2ed")
	if a.dark {
		t = ui.DarkTheme()
		t.Font = "Martian Grotesk"
		t.FontSize = 13
		t.Radius = 8
		t.Background = ui.Hex("#050505")
		t.Surface = ui.Hex("#111111")
		t.SurfaceHover = ui.Hex("#1c1c1c")
		t.SurfacePressed = ui.Hex("#262626")
		t.Border = ui.Hex("#2b2b2b")
		t.Text = ui.Hex("#eeeeee")
		t.TextMuted = ui.Hex("#a0a0a0")
		t.Accent = ui.Hex("#dedede")
		t.AccentText = ui.Hex("#111111")
		t.AccentHover = ui.Hex("#ffffff")
		t.AccentPressed = ui.Hex("#bdbdbd")
		t.Selection = ui.Hex("#252525")
		t.Focus = ui.Hex("#909090")
		t.Scrollbar = ui.Hex("#606060")
	}
	t.Spacing = a.densityMetrics().unit
	return t
}

var icons = func() map[string]*ui.SVG {
	paths := map[string]string{
		"chapter-back": `<path d="M5 5v14m8-14-7 7 7 7m8-14-7 7 7 7"/>`,
		"chapter-next": `<path d="M19 5v14M11 5l7 7-7 7M3 5l7 7-7 7"/>`,
		"chapters":     `<path d="M9 5h12M9 12h12M9 19h12"/><path d="m2 3 4 2-4 2zm0 7 4 2-4 2zm0 7 4 2-4 2z"/>`,
		"settings":     `<path d="M4 6h16M4 12h16M4 18h16"/><circle cx="9" cy="6" r="2" fill="currentColor"/><circle cx="15" cy="12" r="2" fill="currentColor"/><circle cx="8" cy="18" r="2" fill="currentColor"/>`,
		"refresh":      `<path d="M20 7v5h-5M4 17v-5h5"/><path d="M5.1 7a8 8 0 0 1 13.2-2L20 7M4 17l1.7 2A8 8 0 0 0 19 17"/>`,
		"more":         `<circle cx="5" cy="12" r="1"/><circle cx="12" cy="12" r="1"/><circle cx="19" cy="12" r="1"/>`,
		"play":         `<path d="m8 5 11 7-11 7Z" fill="currentColor" stroke="none"/>`,
		"pause":        `<path d="M8 5v14M16 5v14" stroke-width="4"/>`,
		"open":         `<path d="M3 7V5h7l3 3h8v12H3Z"/><path d="M9 14h6m-3-3v6"/>`,
		"queue":        `<path d="M4 6h16M4 12h16M4 18h10m4-2 3 2-3 2"/>`,
		"back":         `<path d="M5 5v14m14-14L8 12l11 7Z"/>`,
		"next":         `<path d="M19 5v14M5 5l11 7-11 7Z"/>`,
		"volume":       `<path d="m11 5-6 4H2v6h3l6 4ZM15 8a6 6 0 0 1 0 8m3-11a10 10 0 0 1 0 14"/>`,
		"expand":       `<path d="M8 3H3v5m13-5h5v5M3 16v5h5m13-5v5h-5"/>`,
		"moon":         `<path d="M21 13a9 9 0 0 1-10-10 9 9 0 1 0 10 10Z"/>`,
		"close":        `<path d="m6 6 12 12M6 18 18 6"/>`,
		"check":        `<path d="m5 12.5 4.5 4.5L19 7.5"/>`,
		"min":          `<path d="M5 12h14"/>`,
		"max":          `<rect x="5" y="5" width="14" height="14" rx="1"/>`,
		"video":        `<rect x="3" y="4" width="18" height="16" rx="4"/><path d="m10 8 6 4-6 4Z"/>`,
		"sub":          `<rect x="3" y="5" width="18" height="14" rx="2"/><path d="M6 12h5m2 0h5M6 15h3m2 0h7"/>`,
		"audio":        `<path d="M9 18V5l12-2v13M9 8l12-2"/><circle cx="6" cy="18" r="3"/><circle cx="18" cy="16" r="3"/>`,
	}
	out := map[string]*ui.SVG{}
	for k, p := range paths {
		out[k] = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">` + p + `</svg>`))
	}
	return out
}()

func icon(c *ui.Context, name, label string) ui.Element { return keyedIcon(c, label, name, label) }

// keyedIcon is icon under a key that outlives its label, so a button whose
// icon and label change together, as play/pause, crosses over its icon.
func keyedIcon(c *ui.Context, key any, name, label string) ui.Element {
	t := c.Theme()
	b := ui.ButtonBase(c.Key(key)).Label(label).Tooltip(label).Size(36, 36).Padding(0).Center().Radius(8).Background(ui.Transparent).TextColor(t.Text).Transition(colorFade)
	if b.Hovered() {
		b.Background(t.SurfaceHover)
	}
	press(b, .94)
	b.Children(func() { ui.Icon(c, icons[name]).Size(18, 18).ContentTransition(iconSwap) })
	return b
}

// Strong curves; the built-in EaseOut is too soft for UI answering a click.
var (
	easeOut    = ui.CubicBezier(.23, 1, .32, 1)
	easeDrawer = ui.CubicBezier(.32, .72, 0, 1)
	easeHover  = ui.CubicBezier(.25, .1, .25, 1)
)

// Hover and selection backgrounds fade instead of snapping.
var colorFade = ui.ElementTransition{Colors: true, Duration: 120 * time.Millisecond, Ease: easeHover}

// Thumbnails that finish loading fade in over their placeholder.
var thumbnailFade = ui.ElementTransition{Colors: true, Duration: 180 * time.Millisecond, Ease: easeOut, Enter: &ui.Motion{}}

// Icons that change meaning (play/pause, queued) cross over quickly.
var iconSwap = ui.ContentTransition{Animation: &ui.Animation{Duration: 160 * time.Millisecond, Ease: easeOut}}

// press shrinks a control slightly while it is held. Poses only repaint,
// so holding a button rebuilds nothing.
var pressMotion = ui.Animation{Duration: 120 * time.Millisecond, Ease: easeOut}

func press(e ui.Element, scale float32) ui.Element {
	pose := ui.Pose{}
	if e.Pressed() {
		pose.Scale = scale
	}
	return e.AnimatePose(pose, pressMotion)
}

// The selection pill slides between sibling choices instead of blinking.
type selectionPill struct{ group string }

var pillSlide = ui.Animation{Duration: 220 * time.Millisecond, Ease: easeOut}

// choice is a flat button whose selected state is a pill shared by its group.
func choice(c *ui.Context, group, label string, active bool) ui.Element {
	t := c.Theme()
	b := ui.ButtonBase(c.Key(label)).Label(label).Padding(10, 12).Radius(t.Radius).Background(ui.Transparent).Transition(colorFade)
	if b.Hovered() && !active {
		b.Background(t.SurfaceHover)
	}
	press(b, .98)
	b.Children(func() {
		// The label in the flow sizes the button. Absolute children paint
		// over the flow, so while the pill shows, a copy of the label
		// above it does the showing.
		sizer := ui.Text(c.Key("label"), label).SingleLine().MinWidth(0).TextColor(t.Text)
		if !active {
			return
		}
		sizer.Opacity(0)
		ui.Box(c.Key("pill")).Absolute().Left(0).Top(0).Right(0).Bottom(0).Radius(t.Radius).Background(t.Selection).
			Transition(ui.ElementTransition{Animation: &pillSlide, Match: selectionPill{group}})
		ui.Row(c.Key("face")).Absolute().Left(0).Top(0).Right(0).Bottom(0).Padding(10, 12).Center().PassThrough().Transition(ui.ElementTransition{Duration: 160 * time.Millisecond, Ease: easeHover, Enter: &ui.Motion{}}).Children(func() {
			ui.Text(c, label).SingleLine().MinWidth(0).TextColor(t.Accent)
		})
	})
	return b
}

// Dialogs and popovers rise from 96% and fade in; a dialog's backdrop
// frosts over the same beat.
var panelEnter = []ui.Keyframe{{At: 0, Value: 0}, {At: 200 * time.Millisecond, Value: 1, Ease: easeOut}}

// enter fades a newly opened dialog in from slightly smaller,
// and returns it for its Dismissed check.
func enter(panel ui.Element) ui.Element {
	if panel == (ui.Element{}) {
		return panel
	}
	k := panel.Keyframes("enter", panelEnter)
	s := .96 + .04*k
	return panel.Opacity(k).Scale(s, s)
}

// enterFrom is enter for a popover: it grows out of its trigger, from
// where the last frame placed the two.
func enterFrom(panel, trigger ui.Element) ui.Element {
	if panel == (ui.Element{}) {
		return panel
	}
	p, a := panel.Bounds(), trigger.Bounds()
	if p.W > 0 && p.H > 0 {
		oy := float32(0)
		if p.Y < a.Y {
			oy = 1
		}
		panel.TransformOrigin(max(0, min(1, (a.X+a.W/2-p.X)/p.W)), oy)
	}
	return enter(panel)
}

// frost swaps a popover's flat fill for frosted glass. Popovers mostly sit
// over the black viewport, which untinted glass turns a muddy gray in the
// light theme, so the theme's background tints it.
func frost(c *ui.Context, panel ui.Element) ui.Element {
	if panel == (ui.Element{}) {
		return panel
	}
	tint := c.Theme().Background
	tint.A = 217 // 85%
	return panel.Material(glass.Glass{Tint: tint})
}

// frostedModal is ui.Modal with a backdrop that blurs what is behind it
// and a glass panel.
func frostedModal(c *ui.Context, open *bool, fn func()) ui.Element {
	if !*open {
		return ui.Element{}
	}
	t := c.Theme()
	w, h := c.Size()
	return ui.DialogBase(c, open, func(back, panel ui.Element) {
		k := back.Keyframes("frost", panelEnter)
		back.Material(glass.Blur{Radius: 10 * k})
		panel.Padding(t.Space(5)).Gap(t.Space(3)).Radius(t.Space(2.5)).MaxWidth(w-t.Space(10)).MaxHeight(h-t.Space(10)).
			Material(glass.Glass{}).Shadow(0, 10, 30, 0, ui.RGBA(0, 0, 0, 0.3))
		fn()
	})
}
