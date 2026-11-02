package main

import (
	"fmt"
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"math"
)

type subtitleAppearance struct {
	Font        string  `json:"font"`
	Size        float64 `json:"size"`
	Color       string  `json:"color"`
	Outline     float64 `json:"outline"`
	Shadow      float64 `json:"shadow"`
	Position    float64 `json:"position"`
	Bold        bool    `json:"bold"`
	Background  bool    `json:"background"`
	OverrideASS bool    `json:"overrideASS"`
}

var subtitleColors = map[string]string{"White": "#ffffff", "Warm white": "#f4ecd8", "Yellow": "#ffe36a", "Cyan": "#9de8ff"}

func defaultSubtitleAppearance() subtitleAppearance {
	return subtitleAppearance{Font: "Segoe UI", Size: 38, Color: "White", Outline: 1.65, Position: 100}
}
func normalizedSubtitleAppearance(s subtitleAppearance) subtitleAppearance {
	d := defaultSubtitleAppearance()
	known := false
	for _, font := range []string{"Segoe UI", "Arial", "Verdana", "Tahoma"} {
		if s.Font == font {
			known = true
		}
	}
	if !known {
		s.Font = d.Font
	}
	if _, ok := subtitleColors[s.Color]; !ok {
		s.Color = d.Color
	}
	clamp := func(value, low, high, fallback float64) float64 {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return fallback
		}
		return min(high, max(low, value))
	}
	if s.Size == 0 {
		s.Size = d.Size
	}
	s.Size = clamp(s.Size, 16, 72, d.Size)
	s.Outline = clamp(s.Outline, 0, 6, d.Outline)
	s.Shadow = clamp(s.Shadow, 0, 6, d.Shadow)
	s.Position = clamp(s.Position, 0, 100, d.Position)
	return s
}

type styleProperty struct {
	name  string
	value any
}

func subtitleStyleProperties(s subtitleAppearance) []styleProperty {
	border, override := "outline-and-shadow", "scale"
	if s.Background {
		border = "background-box"
	}
	if s.OverrideASS {
		override = "force"
	}
	return []styleProperty{{"sub-font", s.Font}, {"sub-font-size", s.Size}, {"sub-color", subtitleColors[s.Color]}, {"sub-outline-size", s.Outline}, {"sub-outline-color", "#000000"}, {"sub-shadow-offset", s.Shadow}, {"sub-back-color", "#b3000000"}, {"sub-border-style", border}, {"sub-pos", s.Position}, {"sub-bold", s.Bold}, {"sub-ass-override", override}}
}
func (a *app) applySubtitleAppearance() {
	a.subtitleStyle = normalizedSubtitleAppearance(a.subtitleStyle)
	if a.engine == nil {
		return
	}
	next, old := subtitleStyleProperties(a.subtitleStyle), subtitleStyleProperties(a.appliedSubtitleStyle)
	for i, property := range next {
		if !a.styleApplied || property.value != old[i].value {
			a.send("set_property", property.name, property.value)
		}
	}
	a.appliedSubtitleStyle = a.subtitleStyle
	a.styleApplied = true
}
func (a *app) subtitleAppearanceChanged() { a.applySubtitleAppearance(); a.persist() }
func (a *app) invalidateWindows() {
	if a.window != nil {
		a.window.Invalidate()
	}
	if a.settingsWindow != nil {
		a.settingsWindow.Invalidate()
	}
}
func (a *app) openSettings() {
	a.settingsOpen = true
	a.captionsOpen = false
	a.audioOpen = false
	a.chaptersOpen = false
	if a.settingsWindow != nil {
		a.settingsWindow.Show()
		a.settingsWindow.Focus()
		a.settingsWindow.Invalidate()
		a.layoutHost()
		return
	}
	if a.window == nil {
		return
	}
	a.settingsWindow = mygo.NewWindow(mygo.WindowOptions{Title: "Settings — mswplayr", Width: 720, Height: 760, MinWidth: 600, MinHeight: 440, Parent: a.window, Content: ui.View(a.settingsWindowView), BackgroundColor: "#111111"})
	a.settingsWindow.OnClosed(func() { a.settingsWindow = nil; a.settingsOpen = false; a.invalidateWindows() })
	a.settingsWindow.Focus()
	a.layoutHost()
}
func (a *app) closeSettings() {
	a.settingsOpen = false
	if a.settingsWindow != nil {
		a.settingsWindow.Close()
	}
	a.invalidateWindows()
	a.layoutHost()
}
func (a *app) openSubtitleAppearance() { a.settingsTarget = "subtitle-appearance"; a.openSettings() }
func (a *app) settingsWindowView(c *ui.Context) {
	c.SetTheme(a.theme())
	c.Root().Background(c.Theme().Background)
	c.OnShortcut(0, ui.KeyEscape, a.closeSettings)
	c.OnShortcut(ui.Ctrl, ui.KeyO, a.browseFolder)
	ui.Column(c).Fill().Children(func() { a.settingsView(c) })
}
func (a *app) settingsFallback(c *ui.Context) {
	if a.settingsWindow != nil {
		return
	}
	w, h := c.Size()
	ui.Modal(c, &a.settingsOpen, func() {
		ui.Column(c).Width(min(680, w-80)).Height(min(680, h-100)).Children(func() { a.settingsView(c) })
	})
}
func (a *app) subtitleAppearanceView(c *ui.Context) {
	t := c.Theme()
	d := a.densityMetrics()
	panel := ui.Column(c.Key("subtitle-appearance")).FillWidth().Background(t.Surface).Radius(12).Padding(d.panel).Gap(16)
	if a.settingsTarget == "subtitle-appearance" {
		panel.ScrollIntoView()
		a.settingsTarget = ""
	}
	panel.Children(func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Children(func() {
			ui.Text(c, "Subtitle appearance").FontSize(16).Bold()
			ui.Spacer(c)
			ui.Button(c, "Reset").Label("Reset subtitle appearance").OnClick(func() { a.subtitleStyle = defaultSubtitleAppearance(); a.subtitleAppearanceChanged() })
		})
		ui.Row(c).FillWidth().Gap(12).Children(func() {
			ui.Column(c).Grow(1).Gap(6).Children(func() {
				ui.Text(c, "Font")
				ui.Select(c, &a.subtitleStyle.Font, []string{"Segoe UI", "Arial", "Verdana", "Tahoma"}).Label("Subtitle font").FillWidth().OnChange(a.subtitleAppearanceChanged)
			})
			ui.Column(c).Width(160).Gap(6).Children(func() {
				ui.Text(c, "Text color")
				ui.Select(c, &a.subtitleStyle.Color, []string{"White", "Warm white", "Yellow", "Cyan"}).Label("Subtitle color").FillWidth().OnChange(a.subtitleAppearanceChanged)
			})
		})
		slider := func(label string, value *float64, low, high, step float64) {
			ui.Row(c).FillWidth().Gap(16).AlignItems(ui.Center).Children(func() {
				ui.Text(c, label).Width(140)
				ui.Slider(c.Key(label), value, low, high).Step(step).Label(label).Grow(1).OnChange(a.subtitleAppearanceChanged)
				ui.Text(c, fmt.Sprintf("%g", *value)).Width(36).TextColor(t.TextMuted)
			})
		}
		slider("Subtitle size", &a.subtitleStyle.Size, 16, 72, 1)
		slider("Outline", &a.subtitleStyle.Outline, 0, 6, .25)
		slider("Shadow", &a.subtitleStyle.Shadow, 0, 6, .25)
		slider("Vertical position", &a.subtitleStyle.Position, 0, 100, 1)
		ui.Text(c, "Position: 0 is top, 100 is bottom.").FontSize(12).TextColor(t.TextMuted)
		ui.Row(c).FillWidth().Gap(20).Children(func() {
			ui.Checkbox(c, &a.subtitleStyle.Bold, "Bold text").OnChange(a.subtitleAppearanceChanged)
			ui.Checkbox(c, &a.subtitleStyle.Background, "Dark background").OnChange(a.subtitleAppearanceChanged)
		})
		ui.Checkbox(c, &a.subtitleStyle.OverrideASS, "Apply to styled ASS/SSA captions").OnChange(a.subtitleAppearanceChanged)
		ui.Box(c).FillWidth().Height(96).Radius(8).Background(ui.Hex("#080808")).Center().Children(func() {
			s := a.subtitleStyle
			text := ui.Text(c, "Subtitle preview").Font(s.Font).FontSize(float32(s.Size/1.5)).TextColor(ui.Hex(subtitleColors[s.Color])).Padding(6, 12)
			if s.Background {
				text.Background(ui.Hex("#303030"))
			}
			if s.Bold {
				text.Bold()
			}
		})
		ui.Text(c, "Font and color settings apply to text subtitles.").FontSize(12).TextColor(t.TextMuted)
	})
}
