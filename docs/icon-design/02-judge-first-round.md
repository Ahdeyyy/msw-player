# Icon judge — first round (20 → 4 → 1)

Judge: Claude Opus 5.5 (review only, no application code touched).
Inputs inspected: `01-twenty-concepts.png` contact sheet (128 px + 16/24/32 px previews), individual 256 px PNGs for 03, 04, 07, 11, 13, 16, 18, 19, and the raw SVG geometry in `concepts.json`.
Brief: a minimal, distinctive video-player mark for a native Windows app. Neutral black and dark tones. No purple, no gradients, no excess detail, nothing that looks like generic AI output.

Forms were judged as drawn, not by their names. Similarity notes record what a viewer is likely to read into a shape. They are **not** trademark searches or clearance.

## Scoring

Each criterion is scored 1–5:
**Sil** = distinctive silhouette · **Small** = legibility at 16/24/32 px · **Media** = reads as video/playback · **Bal** = optical balance in the tile · **Mem** = brand memorability.

| # | Concept | Sil | Small | Media | Bal | Mem | **Total** | What it actually reads as / objection |
|---|---|---|---|---|---|---|---|---|
| 01 | Splice | 2 | 2 | 2 | 3 | 2 | **11** | A stacked "Z" or hamburger menu. Its four fragments blur together at 16 px. The diagonal two-piece "Z" construction resembles Zendesk-style marks. |
| 02 | Open frame | 3 | 4 | 1 | 4 | 2 | **14** | It's simply a letter **G**. That has no link to the product name or to video, and it looks like a generic monogram. |
| 03 | Ribbon | 3 | 2 | 1 | 3 | 2 | **11** | An isometric house or box. The 55 % grey facet adds fake depth, which breaks the flat, no-gradient brief. Too much detail at 16 px. |
| 04 | Eclipse | 4 | 3 | 1 | 4 | 3 | **15** | Reads as a **moon**. On Windows that clashes with the Night light, Dark mode and Do Not Disturb glyphs. The detached square turns into a speck at 16 px. |
| 05 | Corner cut | 3 | 2 | 1 | 4 | 1 | **11** | This is the **fullscreen/expand** UI glyph. It would look like a button, not an app. The corners break apart at 16 px. |
| 06 | Loop | 2 | 2 | 2 | 3 | 2 | **11** | A rounded "O" or stop-button ring. The corner notch that carries the idea is invisible below 32 px. |
| 07 | Keystone | 4 | 4 | 4 | 4 | 3 | **19** | Clearly a screen or projection. The frame is too heavy and the aperture too small. The skewed-perspective window loosely echoes the Windows 10/11 logo, which matters on this platform. Without refinement it risks looking like a generic frame. |
| 08 | Flipbook | 3 | 3 | 1 | 3 | 2 | **12** | Reads instantly as the classic **folded-map** glyph (maps/navigation apps). |
| 09 | Tangent | 3 | 3 | 1 | 3 | 2 | **12** | Reads as "€" or "e". The bar overlapping the ring looks like a fintech or crypto mark. Nothing in it suggests media. |
| 10 | Folded w | 4 | 4 | 1 | 4 | 2 | **15** | A plain **W**. No filmstrip is visible, and W is not the app's initial. |
| 11 | Beam | 4 | 4 | 2 | 3 | 2 | **15** | Reads as a **speaker/volume** glyph, or a mirrored video-call camera (the Zoom/Meet camera family). For a video player it says audio. |
| 12 | Shutter | 3 | 3 | 1 | 3 | 2 | **12** | A stylised F or R, or a maze corner. There's no sense of motion or a screen. |
| 13 | Focus | 4 | 5 | 3 | 4 | 3 | **19** | The sturdiest shape at 16 px. The dot reads as record or lens. But "C + dot" monograms are very common (fintech, consultancies), and C is not the initial. The dot nearly touches the arms. |
| 14 | Current | 3 | 3 | 1 | 3 | 2 | **12** | Water, waves or the "≈" sign. Points to weather, audio or utility apps. |
| 15 | Between frames | 2 | 2 | 2 | 3 | 2 | **11** | Four facets is too busy. It reads as an hourglass or "Z" built from triangle pairs (Zendesk-adjacent). |
| 16 | Cadence | 4 | 4 | 4 | 3 | 5 | **20** | Three slices whose outline tapers into an *implied* play arrow, and the fold hints at an "M" for mswplayr. It's ownable without using the cliché triangle. The objection: the shear is inconsistent (slice 1 tilts up, slices 2 and 3 tilt down), so it currently drifts toward "bar chart". The 10-unit gaps are about 1.25 px at 16 px. |
| 17 | Ticket | 4 | 3 | 2 | 4 | 2 | **15** | Reads as a **%** or coupon (discounts). The slash gap closes up at 16 px. |
| 18 | Afterimage | 3 | 4 | 1 | 4 | 1 | **13** | This is the Windows **copy / restore-down** glyph. Users would read it as a system control. |
| 19 | Cutout | 4 | 4 | 5 | 4 | 3 | **20** | Unmistakably video. But a play triangle is the category cliché, and the notched banner reads as a bookmark ("watch later"). The notch and the triangle compete at 16 px. |
| 20 | Comet | 2 | 3 | 1 | 2 | 2 | **10** | A leaf, guitar pick or speech bubble, plus an unexplained square. It doesn't read as media and the balance is lopsided. |

**Eliminated outright:** 03 (shaded facet breaks the flat brief), 05 and 18 (look like OS UI controls), 04, 08 and 14 (read as a different, existing glyph meaning), 02, 09, 10 and 12 (letterforms with no media signal), and 01, 06, 15 and 20 (fall apart at small sizes).

## Shortlist (4)

1. **16 Cadence** — 20
2. **19 Cutout** — 20
3. **07 Keystone** — 19
4. **13 Focus** — 19

## Provisional winner: 16 Cadence

Cadence and Cutout tie on points. Cadence wins on the user's top priority: it is the **distinctive, non-generic** option. It signals playback through rhythm and an implied arrow rather than drawing the triangle every video app uses. It also carries a quiet "M" that can belong to this product. Cutout is the safe fallback if testing shows Cadence isn't read as media.

## Exact refinements (128-unit grid, tile `rect x=4 y=4 w=120 h=120 rx=28`)

### 16 Cadence (winner)
- **Make the fold rhythmic.** Alternate the slopes up / down / up with the same shear (8 units) on every slice. Step the heights 66 / 50 / 34, center each slice on y = 64, and widen the gaps to 11 units (≈1.4 px at 16 px):
  ```svg
  <path d="M26 35 44 27v66l-18 8zM55 35l18 8v50l-18-8zM84 51l18-8v34l-18 8z"/>
  ```
  Bounding box: x 26–102, y 27–101, centered on (64, 64).
- If the zig-zag still reads as a chart in testing, use **variant B**. It drops the fold and slices one tapered arrow, which reads as play more strongly but is less ownable:
  ```svg
  <path d="M26 28 44 34v60l-18 6zM55 37l18 6v42l-18 6zM84 46l18 6v24l-18 6z"/>
  ```
- **16 px ICO frame:** hand-snap the slices to 3 px width with 1 px gaps (11 px total), heights 12 / 9 / 6 px, centered. Don't scale the vector down automatically. **24 px:** 4 px slices, 2 px gaps.

### 19 Cutout (runner-up)
- Make the notch shallower (20 → 12) so it reads less like a bookmark. Keep the slanted top as the distinctive cue: `M30 34 98 26v74l-34-12-34 12z`.
- Rebuild the triangle as near-equilateral and optically centered, with its centroid on x ≈ 64: `M55 46v32l27-16z`. Use even-odd fill or cut it from the sheet.
- At 16 px, drop the notch entirely and keep only the slanted sheet plus the triangle.

### 07 Keystone
- Soften the skew (left/right edges 56 / 72 instead of 52 / 76) so it echoes the Windows logo less. Thin the frame to about 14 units so the aperture grows. Center both edges on y = 64:
  `M26 36 102 28v72l-76-8zM40 50v28l48 6V44z` (even-odd).

### 13 Focus
- Thin the arms to 16 units and shrink the dot to r = 11 at (72, 64), so the dot has at least a 10-unit gap to each arm:
  `M47 27h41v16H49v42h39v16H47a16 16 0 0 1-16-16V43a16 16 0 0 1 16-16z` + `<circle cx="72" cy="64" r="11"/>`.

## Color and tile (all finalists)
- Keep it to two flat colors with no opacity tricks: tile `#0b0b0b`, glyph `#f2eee6`. If the warm cream looks beige next to other taskbar icons, try a cooler off-white like `#f4f3ef`. No purple, no gradient.
- The Windows 11 dark taskbar is about `#1c1c1c`–`#202020`, so a near-black tile disappears against it. Add a hairline inner edge: `stroke="#ffffff" stroke-opacity=".10"`, stroke-width 1.5 at 128 units, inset by half the stroke. Use it on 32 px frames and larger only.
- Keep the glyph inside a 76 × 76 live area (x/y 26–102). This leaves about 22 units of tile padding, so the mark keeps its presence at 24 px without crowding the rounded corners.
- Ship hand-tuned 16, 20, 24 and 32 px ICO frames with integer-snapped edges, plus 48/64/256 px frames from the vector.

## Verdict
**Shortlist: 16 Cadence, 19 Cutout, 07 Keystone, 13 Focus. Provisional winner: 16 Cadence**, on condition that the shear is unified and the gaps widened as specified. If testing shows people don't read Cadence as video, fall back to Cutout.
