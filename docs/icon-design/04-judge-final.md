# Icon judge — final round (5 specimens → 1)

Judge: Claude Opus 5.5. This is an independent final review. I made no application edits and used no subagents.

**What I inspected:**
- `03-refined-finalists.png`
- Each `finalists/*.png` at 512 px
- The native 16, 20, 24 and 32 px PNGs, enlarged with nearest-neighbour scaling so each pixel can be seen
- The SVG sources and the small-size drawings in `scripts/icon-design/refine.go`

Shapes are judged as drawn. Any resemblance I note is a reading risk, not a trademark finding.

## Ranked comparison

Scores run from 1 to 5. **Sil** is how distinctive the silhouette is. **Video** is how clearly it relates to video. **Bal** is optical balance. **Small** is how well it works at 16–32 px.

| Rank | Specimen | Sil | Video | Bal | Small | Total | Verdict |
|---|---|---|---|---|---|---|---|
| **1** | **16 Cadence B (forward)** | 5 | 4 | 4 | 4 | **17** | **Winner** |
| 2 | 19 Cutout | 3 | 5 | 4 | 3 | 15 | Safe fallback; relies on the category cliché |
| 3 | 16 Cadence A (folded) | 4 | 2 | 3 | 3 | 12 | The refinement made it worse (see below) |
| 4 | 07 Keystone | 3 | 3 | 4 | 2 | 12 | A generic window; drifts toward Windows UI |
| 5 | 13 Focus | 3 | 2 | 2 | 4 | 11 | Mark is off-centre; "C + dot" monogram |

### 1. Cadence B (forward), the winner
- **Does it read as motion? Yes.** Three slices taper along one shared envelope, so the eye finishes the play arrow and reads it moving to the right. The arrow is implied rather than drawn, which keeps it clear of the play-triangle cliché. At 48 and 128 px it is the most distinctive and most "video" mark on the board.
- **Objection: it is left-heavy.** The tallest slice is on the left, so the visual centroid sits near x ≈ 56 even though the bounding box is centred on 64. Play shapes need to be nudged right to look centred, and this one isn't.
- **Objection: the small frames don't match the vector.**
  - At 16, 20 and 24 px, the third slice is drawn 4 → 2 px tall. At that size it collapses into a blob of 2–3 pixels, visible in the enlarged 16, 20 and 24 px renders.
  - That taper is also much steeper than the vector, where the third slice is half the height of the first. So the third slice jumps from a nub at 24 px to a full panel at 32 px.
  - The 20 px frame is the 24-unit drawing scaled to 0.83×. Every vertical edge lands on a fractional pixel, so the whole glyph is blurry.
  - The 32 px frame is rasterised straight from the vector. Its edges fall at 6.5, 11, 13.75, 18.25, 21 and 25.5 px, which leaves grey columns on every slice.
- **Other readings:** a perspective corridor, or a "next" glyph. Both are weaker than the play reading and neither is a conflicting app meaning.

### 2. Cutout
- It is the most unambiguous video mark, and the refinement improved it: the notch is shallower and the triangle is better centred.
- It is still a play triangle on a pennant. That is the most expected answer in the category, and it is the least ownable option here.
- The first judge's instruction to drop the notch at 16 px was not implemented. At 16 px the notch is a grey smear and the triangle is about 3 px of mud.
- This is the right fallback only if user testing shows Cadence B isn't read as video.

### 3. Cadence A (folded): the refinement made it worse
- **Does it look like a folded map? Yes, clearly.** The panels have equal width, parallel shear, and alternate up/down/up. That is exactly the folded-map pictogram, the same reading that eliminated concept 08 in round one. Unifying the shear made the zig-zag *more* regular, so it is now more map-like than the original draft.
- The decreasing heights then add a second reading: a descending bar chart, or a signal-strength glyph. It doesn't read as an arrow, and the "M" hint is faint.
- At 16 px, a 1 px shear across a 3 px slice is only grey corner pixels. The fold disappears and what's left is three descending bars.
- Its craft is fine. The idea simply doesn't communicate video.

### 4. Keystone
- The frame weight is better and the aperture is larger.
- At 16 and 24 px the perspective vanishes, and it becomes a hollow rectangle that resembles the Windows "maximize/restore" glyph.
- At large sizes, a single skewed pane still points toward the Windows logo. On Windows that is a real reading risk.

### 5. Focus
- The geometry is off-centre. The glyph spans x 31–88, so its centre is at 59.5, about 4.5 units left of the tile centre. The large render shows the empty space on the right.
- The dot now clears the arms properly, and it is the crispest shape at 16 px.
- "C + dot" still reads as a generic monogram or a record button. It's not the app's initial and it doesn't suggest moving pictures.

## Do the fine edges collapse? Do the small grids need adjusting?
- **Yes, for Cadence B at 16, 20 and 24 px:** the third slice collapses. Correction 1 fixes this.
- **The 20 px and 32 px frames are blurry, because neither is hand-drawn:** 20 px is a scaled 24 px frame and 32 px is the raw vector. Corrections 1 and 3 add native drawings.
- **The tile edges are soft at 16 and 24 px.** The `rect x=".5"` and `x=".75"` insets give every edge pixel partial coverage. The result is a grey ring on light backgrounds and a mushy edge on the dark taskbar. Correction 2 fixes this.
- Cadence A doesn't need its grids fixed, because it should not ship.

## Final polish corrections (at most 3, all for Cadence B)

1. **Redraw the small frames so the taper matches the vector and no slice collapses.** Every slice keeps a 1-pixel-per-slice-width shear and its centre on the tile's middle row.
   - **16 px:** change only the third slice.
     `M3 2 6 3v10l-3 1zM7 4l3 1v6l-3 1zM11 5l3 1v4l-3 1z`
     The slices step 12/10 → 8/6 → 6/4 px tall, from each slice's left edge to its right edge.
   - **24 px:** change slices 2 and 3.
     `M4 4 8 5v14l-4 1zM10 6l4 1v10l-4 1zM16 8l4 1v6l-4 1z`
     The slices step 16/14 → 12/10 → 8/6 px.
   - **20 px:** use a native 20-unit drawing (`viewBox="0 0 20 20"`) instead of scaling the 24 px one.
     `M4 3 7 4v12l-3 1zM9 5l3 1v8l-3 1zM14 7l3 1v4l-3 1z`
     The slices step 14/12 → 10/8 → 6/4 px.

2. **Align the small tiles to whole pixels.**
   - 16 px: use `x="0" y="0" width="16" height="16" rx="3.5"`.
   - 20 px: use `x="0" y="0" width="20" height="20" rx="4.5"`.
   - 24 px: use `x="0" y="0" width="24" height="24" rx="5.5"`.
   - With these, only the corner arcs are antialiased. The 32 px and larger frames keep the hairline edge from round one.

3. **Nudge B right to balance it optically, and add a hand-snapped 32 px frame.**
   - In the vector, wrap the mark in `<g transform="translate(2 0)">`. That moves the glyph to x 28–104, which is still well inside the tile padding.
   - For the 32 px ICO frame, add this native drawing, with 5 px slices and 3 px gaps:
     `M6 7 11 8v16l-5 1zM14 9l5 1v12l-5 1zM22 11l5 1v8l-5 1z`
     It sits at x 6–27, which is the same +0.5 px nudge. The slices step 18/16 → 14/12 → 10/8 px.
   - The 48, 64 and 256 px frames can stay vector-rendered.

## Verdict

**The final winner is 16 Cadence B (forward).** It has the most distinctive silhouette in the set, and it reads as playback and motion without drawing the stock triangle. Once the three corrections above are applied, its small sizes are sound.

Cadence A is retired: the refinement turned it into a folded map.

Cutout stays as the documented fallback if users don't read Cadence B as video.
