package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/egoist/mygo/ui"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
)

type finalist struct{ ID, Label, Shape string }

var finalists = []finalist{
	{"cadence-a", "16  Cadence / folded", `<path d="M26 35 44 27v66l-18 8zM55 35l18 8v50l-18-8zM84 51l18-8v34l-18 8z"/>`},
	{"cadence-b", "16  Cadence / forward", `<path d="M26 28 44 34v60l-18 6zM55 37l18 6v42l-18 6zM84 46l18 6v24l-18 6z"/>`},
	{"cutout", "19  Cutout", `<path fill-rule="evenodd" d="M30 34 98 26v74l-34-12-34 12zM55 46v32l27-16z"/>`},
	{"keystone", "07  Keystone", `<path fill-rule="evenodd" d="M26 36 102 28v72l-76-8zM40 50v28l48 6V44z"/>`},
	{"focus", "13  Focus", `<path d="M47 27h41v16H49v42h39v16H47a16 16 0 0 1-16-16V43a16 16 0 0 1 16-16z"/><circle cx="72" cy="64" r="11"/>`},
}

func finalistSVG(f finalist, size int) string {
	edge := ""
	if size >= 32 {
		edge = ` stroke="#303030" stroke-width=".8"`
	}
	if f.ID == "cadence-a" || f.ID == "cadence-b" {
		if size == 16 {
			shape := `M3 3 6 2v11l-3 1zM7 4l3 1v7l-3-1zM11 6l3-1v5l-3 1z`
			if f.ID == "cadence-b" {
				shape = `M3 2 6 3v10l-3 1zM7 4l3 1v6l-3 1zM11 6l3 1v2l-3 1z`
			}
			return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16"><rect x=".5" y=".5" width="15" height="15" rx="3.5" fill="#0b0b0b"/><path fill="#f4f3ef" d="` + shape + `"/></svg>`
		}
		if size == 20 || size == 24 {
			shape := `M4 6 8 4v14l-4 2zM10 6l4 2v10l-4-2zM16 10l4-2v6l-4 2z`
			if f.ID == "cadence-b" {
				shape = `M4 4 8 5v14l-4 1zM10 7l4 1v8l-4 1zM16 10l4 1v2l-4 1z`
			}
			return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><rect x=".75" y=".75" width="22.5" height="22.5" rx="5.25" fill="#0b0b0b"/><path fill="#f4f3ef" d="` + shape + `"/></svg>`
		}
	}
	return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 128 128"><rect x="4.4" y="4.4" width="119.2" height="119.2" rx="27.6" fill="#0b0b0b"` + edge + `/><g fill="#f4f3ef">` + f.Shape + `</g></svg>`
}
func refine() {
	root := filepath.Join("docs", "icon-design", "finalists")
	os.MkdirAll(root, 0755)
	board := image.NewRGBA(image.Rect(0, 0, 1200, 740))
	draw.Draw(board, board.Bounds(), image.NewUniform(color.RGBA{238, 238, 234, 255}), image.Point{}, draw.Src)
	for i, f := range finalists {
		source := finalistSVG(f, 256)
		os.WriteFile(filepath.Join(root, f.ID+".svg"), []byte(source), 0644)
		savePNG(filepath.Join(root, f.ID+".png"), raster(source, 512))
		for _, size := range []int{16, 20, 24, 32, 48, 64} {
			savePNG(filepath.Join(root, fmt.Sprintf("%s-%d.png", f.ID, size)), raster(finalistSVG(f, size), size))
		}
		tile := ui.Render(func(c *ui.Context) {
			t := ui.LightTheme()
			t.Background = ui.Hex("#eeeeea")
			c.SetTheme(t)
			ui.Column(c).Fill().Padding(20, 12).Gap(22).AlignItems(ui.Center).Children(func() {
				ui.Image(c, ui.MustParseSVG([]byte(source))).Size(176, 176)
				ui.Text(c, f.Label).FontSize(13).Bold()
				for _, dark := range []bool{false, true} {
					bg, fg := ui.Hex("#ffffff"), ui.Hex("#191919")
					if dark {
						bg, fg = ui.Hex("#202020"), ui.Hex("#eeeeee")
					}
					ui.Column(c).FillWidth().Padding(16).Radius(10).Gap(12).Background(bg).TextColor(fg).Children(func() {
						label := "Light surface"
						if dark {
							label = "Dark taskbar"
						}
						ui.Text(c, label).FontSize(12)
						ui.Row(c).FillWidth().Gap(14).AlignItems(ui.Center).Children(func() {
							for _, size := range []int{16, 24, 32, 48} {
								ui.Column(c).Gap(10).AlignItems(ui.Center).Children(func() {
									ui.Image(c, ui.MustParseSVG([]byte(finalistSVG(f, size)))).Size(float32(size), float32(size))
									ui.Text(c, fmt.Sprint(size)).FontSize(12)
								})
							}
						})
					})
				}
				ui.Text(c, "Monochrome mark").FontSize(12)
				mark := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 128 128"><g fill="#151515">` + f.Shape + `</g></svg>`
				ui.Image(c, ui.MustParseSVG([]byte(mark))).Size(112, 112)
			})
		}, 240, 740, 1)
		draw.Draw(board, image.Rect(i*240, 0, (i+1)*240, 740), tile, image.Point{}, draw.Src)
	}
	savePNG(filepath.Join("docs", "icon-design", "03-refined-finalists.png"), board)
	fmt.Println("Created refined finalist comparisons")
}
func exportWinner(id string) {
	var chosen *finalist
	for i := range finalists {
		if finalists[i].ID == id {
			chosen = &finalists[i]
			break
		}
	}
	if chosen == nil {
		panic("unknown finalist: " + id)
	}
	root := "resources"
	os.MkdirAll(root, 0755)
	source := polishedSVG(*chosen, 1024)
	os.WriteFile(filepath.Join(root, "icon.svg"), []byte(source), 0644)
	savePNG(filepath.Join(root, "icon.png"), raster(source, 1024))
	shape := chosen.Shape
	if chosen.ID == "cadence-b" {
		shape = `<g transform="translate(2 0)">` + shape + `</g>`
	}
	mark := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="24 22 84 84"><g fill="currentColor">` + shape + `</g></svg>`
	os.WriteFile(filepath.Join(root, "mark.svg"), []byte(mark), 0644)
	sizes := []int{16, 20, 24, 32, 40, 48, 64, 128, 256, 512}
	payloads := [][]byte{}
	for _, size := range sizes {
		img := raster(polishedSVG(*chosen, size), size)
		path := filepath.Join(root, fmt.Sprintf("icon-%d.png", size))
		savePNG(path, img)
		if size <= 256 {
			var p bytes.Buffer
			png.Encode(&p, img)
			payloads = append(payloads, p.Bytes())
		}
	}
	var ico bytes.Buffer
	binary.Write(&ico, binary.LittleEndian, [3]uint16{0, 1, uint16(len(payloads))})
	offset := 6 + 16*len(payloads)
	for i, p := range payloads {
		size := sizes[i]
		dimension := byte(size)
		if size == 256 {
			dimension = 0
		}
		ico.Write([]byte{dimension, dimension, 0, 0})
		binary.Write(&ico, binary.LittleEndian, [2]uint16{1, 32})
		binary.Write(&ico, binary.LittleEndian, [2]uint32{uint32(len(p)), uint32(offset)})
		offset += len(p)
	}
	for _, p := range payloads {
		ico.Write(p)
	}
	os.WriteFile(filepath.Join(root, "icon.ico"), ico.Bytes(), 0644)
	proof := ui.Render(func(c *ui.Context) {
		t := ui.LightTheme()
		t.Background = ui.Hex("#eeede9")
		c.SetTheme(t)
		ui.Column(c).Fill().Center().Gap(28).Children(func() {
			ui.Image(c, ui.MustParseSVG([]byte(source))).Size(256, 256)
			ui.Text(c, "mswplayr").FontSize(22).Bold()
			ui.Row(c).Gap(24).AlignItems(ui.Center).Children(func() {
				for _, size := range []int{16, 20, 24, 32, 48, 64} {
					ui.Image(c, ui.MustParseSVG([]byte(polishedSVG(*chosen, size)))).Size(float32(size), float32(size))
				}
			})
		})
	}, 720, 500, 1)
	savePNG(filepath.Join("docs", "icon-design", "05-final-icon.png"), proof)
	fmt.Println("Exported " + chosen.ID + " as SVG, transparent PNG sizes, mark, and nine-frame Windows ICO")
}

// Final judge corrections, separate from the reviewed second-round specimens.
func polishedSVG(f finalist, size int) string {
	if f.ID != "cadence-b" {
		return finalistSVG(f, size)
	}
	paths := map[int]string{
		16: `M3 2 6 3v10l-3 1zM7 4l3 1v6l-3 1zM11 5l3 1v4l-3 1z`,
		20: `M4 3 7 4v12l-3 1zM9 5l3 1v8l-3 1zM14 7l3 1v4l-3 1z`,
		24: `M4 4 8 5v14l-4 1zM10 6l4 1v10l-4 1zM16 8l4 1v6l-4 1z`,
		32: `M6 7 11 8v16l-5 1zM14 9l5 1v12l-5 1zM22 11l5 1v8l-5 1z`,
	}
	if path, ok := paths[size]; ok {
		radius := float64(size) * .23
		return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d"><rect width="%d" height="%d" rx="%.2f" fill="#0b0b0b"/><path fill="#f4f3ef" d="%s"/></svg>`, size, size, size, size, radius, path)
	}
	f.Shape = `<g transform="translate(2 0)">` + f.Shape + `</g>`
	return finalistSVG(f, size)
}
