# Final choice: Cadence / forward

The final mark is three solid slices tapering toward the right: an implied play arrow constructed like consecutive film frames. It uses off-white `#f4f3ef` on a near-black `#0b0b0b` tile, with no gradients or simulated lighting.

## Selection record

1. Twenty distinct rough constructions: [contact sheet](01-twenty-concepts.png), editable SVGs and PNGs in `concepts/`.
2. Claude Opus 5.5, Claude provider: [first critique](02-judge-first-round.md). Cut twenty to four families: Cadence, Cutout, Keystone, Focus.
3. Refined those families and tested two Cadence cuts: [finalist sheet](03-refined-finalists.png).
4. A second independent Claude Opus 5.5 judgement chose **Cadence / forward**: [final critique](04-judge-final.md). The folded version was rejected as map-like. The winning specimen scored 17/20 before final polish.
5. Applied the judge's three corrections: stronger taper at small sizes, native 16/20/24/32 pixel drawings with aligned tile edges, and a two-unit optical shift of the large mark. [Final proof](05-final-icon.png).

The judge's score is a design critique, not trademark clearance or a user-recognition study. The final polish follows the verdict; no additional score is claimed for the corrected exports.

## Assets

- `resources/icon.svg`: editable master.
- `resources/icon.png`: transparent 1024 × 1024 export.
- `resources/icon-{size}.png`: 16, 20, 24, 32, 40, 48, 64, 128, 256, and 512 pixel exports.
- `resources/icon.ico`: nine PNG frames at 16–256 pixels, including the four hand-drawn small sizes.
- `resources/mark.svg`: theme-colored, unboxed mark for the app title bar.

The build uses MyGo's production packaging for executable metadata and manifest, then embeds the ICO's exact frames through the Windows resource API. Both executable copies receive those frames. The icon is also shipped as a standalone ICO and PNG.

## Reproduce

```powershell
go run ./scripts/icon-design
go run ./scripts/icon-design refine
go run ./scripts/icon-design export cadence-b
./scripts/build.ps1
```

The first two commands recreate the reviewed rough sheets. Export applies the final critique separately so the earlier comparison remains an honest record of what was judged.
