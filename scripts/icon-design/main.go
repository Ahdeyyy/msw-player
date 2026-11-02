// Icon exploration and deterministic exports. Run from the repository root.
package main

import (
	"encoding/json"
	"fmt"
	"github.com/egoist/mygo/ui"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
)

type concept struct{ ID, Name, Idea, Shape string }

var concepts = []concept{
	{"01", "Splice", "Two pieces of film joined by a diagonal cut", `<path d="M34 33h51L67 51H34zM94 77v18H43l18-18zM34 57h37L53 75H34zM94 39v30H76V57z"/>`},
	{"02", "Open frame", "A screening frame with a single forward opening", `<path fill-rule="evenodd" d="M43 29h42a14 14 0 0 1 14 14v42a14 14 0 0 1-14 14H43a14 14 0 0 1-14-14V43a14 14 0 0 1 14-14zm2 17v36h37V70H66V56h33V43H45z"/>`},
	{"03", "Ribbon", "A strip folded once around an angled central aperture", `<path d="M29 88V39l25-14 45 27v37L74 103V65L54 53v20z"/><path d="m29 88 25-15 20 12v18z" opacity=".55"/>`},
	{"04", "Eclipse", "A cinema lens reduced to an off-center crescent and square", `<path d="M79 29a37 37 0 1 0 0 70 32 32 0 0 1 0-70z"/><rect x="79" y="55" width="18" height="18" rx="3"/>`},
	{"05", "Corner cut", "Four frame corners, one transformed into a diagonal cut", `<path d="M30 53V30h28v14H44v9zm40-23h28v28H84V44H70zM30 70h14v14h14v14H30zm40 28 28-28v28z"/>`},
	{"06", "Loop", "A continuous film loop interrupted by an offset join", `<path fill-rule="evenodd" d="M51 29h26a25 25 0 0 1 25 25v20a25 25 0 0 1-25 25H51a25 25 0 0 1-25-25V54a25 25 0 0 1 25-25zm1 17a9 9 0 0 0-9 9v18a9 9 0 0 0 9 9h24a9 9 0 0 0 9-9V55a9 9 0 0 0-9-9z"/><path d="m83 27 19 19v19L83 46z" fill="#0b0b0b"/>`},
	{"07", "Keystone", "A projection aperture with subtly converging sides", `<path fill-rule="evenodd" d="m24 38 80-12v76L24 90zm18 15v22l44 7V46z"/>`},
	{"08", "Flipbook", "Three asymmetric pages opening into a screen", `<path d="m28 38 19-8v63l-19 9zm25-8 21 6v62l-21-5zm27 9 21 10v61L80 99z"/>`},
	{"09", "Tangent", "A viewing orbit extending into a single horizontal beam", `<path d="M74 30a35 35 0 1 0 25 59L87 77a18 18 0 1 1-13-30h27V30z"/><rect x="61" y="57" width="40" height="15" rx="7.5"/>`},
	{"10", "Folded w", "A compact custom w constructed from a filmstrip", `<path d="M26 35h17l8 39 13-27 13 27 8-39h17L88 96H73l-9-20-9 20H40z"/>`},
	{"11", "Beam", "A projection beam with a clean staggered end", `<path d="M29 47 81 25v28l19-8v39l-19-8v27L29 81z"/>`},
	{"12", "Shutter", "Three broad nesting turns that suggest a moving aperture", `<path d="M28 28h72v17H45v55H28zM55 55h45v45H83V72H55z"/>`},
	{"13", "Focus", "A circular lens cut into an unusually tall open bracket", `<path d="M47 27h41v17H49v40h39v17H47a16 16 0 0 1-16-16V43a16 16 0 0 1 16-16z"/><circle cx="75" cy="64" r="13"/>`},
	{"14", "Current", "Two opposing ribbons with a quiet, horizontal rhythm", `<path d="M27 43q18-25 37 0t37 0v22q-18 25-37 0T27 65zM27 76q18-16 37 0t37 0v15q-18 16-37 0T27 91z"/>`},
	{"15", "Between frames", "Two offset wedges meeting across a diagonal splice", `<path d="M30 29h67L70 57H30zM98 99H31l27-28h40z"/><path d="m30 40 28 17-28 18zm68 13L70 71l28 17z"/>`},
	{"16", "Cadence", "Three advancing film slices with unequal heights", `<path d="m27 42 18-10v64l-18 10zm28-5 18 10v49l-18-10zm28 15 18 10v24l-18 10z"/>`},
	{"17", "Ticket", "A screening ticket reduced to a split diagonal silhouette", `<path d="M28 38h72v15a11 11 0 0 0 0 22v15H28V75a11 11 0 0 0 0-22z"/><path d="m73 35-25 58h10l25-58z" fill="#0b0b0b"/>`},
	{"18", "Afterimage", "A detached frame corner beside a rounded image tile", `<rect x="29" y="48" width="52" height="52" rx="12"/><path d="M54 28h46v46H85V43H54z"/>`},
	{"19", "Cutout", "A single deep notch transforms an angled sheet", `<path d="m30 35 68-9v76l-34-20-34 20zm20 12v30l28-16z" fill-rule="evenodd"/>`},
	{"20", "Comet", "A single swept capsule and its square afterimage", `<path d="M32 78a18 18 0 0 1 0-28l54-22v48L55 98a18 18 0 0 1-23-20z"/><rect x="88" y="85" width="15" height="15" rx="3"/>`},
}

func svg(shape string) string {
	return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 128 128"><rect x="4" y="4" width="120" height="120" rx="28" fill="#0b0b0b"/><g fill="#f2eee6">` + shape + `</g></svg>`
}
func savePNG(path string, img image.Image) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err = png.Encode(f, img); err != nil {
		panic(err)
	}
}
func raster(source string, size int) *image.RGBA {
	s := ui.MustParseSVG([]byte(source))
	return ui.Render(func(c *ui.Context) {
		c.Root().Background(ui.Transparent)
		ui.Image(c, s).Size(float32(size), float32(size))
	}, size, size, 1)
}
func main() {
	if len(os.Args) > 2 && os.Args[1] == "verify" {
		if err := verifyExport(os.Args[2]); err != nil {
			panic(err)
		}
		fmt.Println("Verified transparent master and all nine embedded icon frames")
		return
	}
	if len(os.Args) > 3 && os.Args[1] == "embed" {
		if err := embedIcon(os.Args[2], os.Args[3]); err != nil {
			panic(err)
		}
		fmt.Println("Embedded native-size icon frames in " + os.Args[2])
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "refine" {
		refine()
		return
	}
	if len(os.Args) > 2 && os.Args[1] == "export" {
		exportWinner(os.Args[2])
		return
	}
	root := filepath.Join("docs", "icon-design")
	os.MkdirAll(filepath.Join(root, "concepts"), 0755)
	board := image.NewRGBA(image.Rect(0, 0, 1200, 1080))
	draw.Draw(board, board.Bounds(), image.NewUniform(color.RGBA{238, 238, 234, 255}), image.Point{}, draw.Src)
	for i, c := range concepts {
		source := svg(c.Shape)
		base := filepath.Join(root, "concepts", c.ID+"-"+c.Name)
		os.WriteFile(base+".svg", []byte(source), 0644)
		savePNG(base+".png", raster(source, 256))
		tile := ui.Render(func(ctx *ui.Context) {
			t := ui.LightTheme()
			t.Background = ui.Hex("#eeeeea")
			ctx.SetTheme(t)
			ui.Column(ctx).Fill().Center().Gap(12).Children(func() {
				ui.Image(ctx, ui.MustParseSVG([]byte(source))).Size(144, 144)
				ui.Text(ctx, c.ID+"  "+c.Name).FontSize(14)
				ui.Row(ctx).Gap(16).AlignItems(ui.Center).Children(func() {
					ui.Image(ctx, ui.MustParseSVG([]byte(source))).Size(16, 16)
					ui.Image(ctx, ui.MustParseSVG([]byte(source))).Size(24, 24)
					ui.Image(ctx, ui.MustParseSVG([]byte(source))).Size(32, 32)
				})
			})
		}, 240, 270, 1)
		draw.Draw(board, image.Rect(i%5*240, i/5*270, (i%5+1)*240, (i/5+1)*270), tile, image.Point{}, draw.Src)
	}
	savePNG(filepath.Join(root, "01-twenty-concepts.png"), board)
	data, _ := json.MarshalIndent(concepts, "", "  ")
	os.WriteFile(filepath.Join(root, "concepts.json"), data, 0644)
	fmt.Println("Created 20 SVG/PNG concepts and contact sheet in " + root)
}
