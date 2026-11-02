package main

import (
	"embed"
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
		"chapters": `<path d="M9 5h12M9 12h12M9 19h12"/><path d="m2 3 4 2-4 2zm0 7 4 2-4 2zm0 7 4 2-4 2z"/>`,
		"settings": `<path d="M4 6h16M4 12h16M4 18h16"/><circle cx="9" cy="6" r="2" fill="currentColor"/><circle cx="15" cy="12" r="2" fill="currentColor"/><circle cx="8" cy="18" r="2" fill="currentColor"/>`,
		"refresh":  `<path d="M20 7v5h-5M4 17v-5h5"/><path d="M5.1 7a8 8 0 0 1 13.2-2L20 7M4 17l1.7 2A8 8 0 0 0 19 17"/>`,
		"more":     `<circle cx="5" cy="12" r="1"/><circle cx="12" cy="12" r="1"/><circle cx="19" cy="12" r="1"/>`,
		"play":     `<path d="m8 5 11 7-11 7Z" fill="currentColor" stroke="none"/>`,
		"pause":    `<path d="M8 5v14M16 5v14" stroke-width="4"/>`,
		"open":     `<path d="M3 7V5h7l3 3h8v12H3Z"/><path d="M9 14h6m-3-3v6"/>`,
		"queue":    `<path d="M4 6h16M4 12h16M4 18h10m4-2 3 2-3 2"/>`,
		"back":     `<path d="M5 5v14m14-14L8 12l11 7Z"/>`,
		"next":     `<path d="M19 5v14M5 5l11 7-11 7Z"/>`,
		"volume":   `<path d="m11 5-6 4H2v6h3l6 4ZM15 8a6 6 0 0 1 0 8m3-11a10 10 0 0 1 0 14"/>`,
		"expand":   `<path d="M8 3H3v5m13-5h5v5M3 16v5h5m13-5v5h-5"/>`,
		"moon":     `<path d="M21 13a9 9 0 0 1-10-10 9 9 0 1 0 10 10Z"/>`,
		"close":    `<path d="m6 6 12 12M6 18 18 6"/>`,
		"min":      `<path d="M5 12h14"/>`,
		"max":      `<rect x="5" y="5" width="14" height="14" rx="1"/>`,
		"video":    `<rect x="3" y="4" width="18" height="16" rx="4"/><path d="m10 8 6 4-6 4Z"/>`,
		"sub":      `<rect x="3" y="5" width="18" height="14" rx="2"/><path d="M6 12h5m2 0h5M6 15h3m2 0h7"/>`,
		"audio":    `<path d="M9 18V5l12-2v13M9 8l12-2"/><circle cx="6" cy="18" r="3"/><circle cx="18" cy="16" r="3"/>`,
	}
	out := map[string]*ui.SVG{}
	for k, p := range paths {
		out[k] = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">` + p + `</svg>`))
	}
	return out
}()

func icon(c *ui.Context, name, label string) ui.Element {
	t := c.Theme()
	b := ui.ButtonBase(c.Key(label)).Label(label).Tooltip(label).Size(36, 36).Padding(0).Center().Radius(8).Background(ui.Transparent).TextColor(t.Text)
	if b.Hovered() {
		b.Background(t.SurfaceHover)
	}
	b.Children(func() { ui.Icon(c, icons[name]).Size(18, 18) })
	return b
}
