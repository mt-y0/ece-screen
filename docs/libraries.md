# Libraries (`lib/*`)

_Last updated: 2026-10-24_

## `lib/fonts` (package `lib`, file `arabic.go`)

- `Registry map[rune]GlyphGroup` — Arabic letter → Isolated/Initial/Medial/Final presentation forms (FE-range) + `ConnectsToNext`. Covers the 28 base letters, hamza forms, ta marbuta, alef maqsura (ى), alef wasla. **Not covered:** Persian/Kurdish letters (پ چ ژ گ ڤ ێ ۆ ڕ ڵ), tatweel (U+0640).
- `Reshape(string) string` — 3 steps: (1) contextual form selection skipping tashkeel; (2) lam-alef ligatures (FEF5–FEFC); (3) bidi: split into Arabic / non-Arabic runs, reverse each Arabic run, then reverse the run order. Output is ready for raylib's left-to-right renderer. Spaces and Latin/digits are "LTR runs".
- `GetArabicCodepoints()` / `LoadFont(ttf []byte) *rl.Font` — bakes NotoKufi at size 64 with ASCII, Latin-1, general punctuation, Arabic (0600–06FF, 08A0–08FF), presentation forms A (FB50–FDFF) and B (FE70–FEFF); bilinear filter.

## `lib/emoji`

- `Load(dir)` — packs `*.png` (Twemoji naming `1f600.png`, `1f1f8-1f1e6.png`) into one 64-px-cell atlas texture. Inert (no-op, nil-safe) atlas if dir missing (a warning is logged; failed PNGs are logged and counted). Must run after `InitWindow`. `main.go` loads `resources/fonts/twemoji` relative to the CWD, falling back to the executable's directory.
- `Split(shapedText) []Segment` — alternating text / emoji segments; folds FE0F, skin tones, ZWJ chains, flag pairs into one key. Base emoji ranges: 1F300–1FAFF and 2600–27BF only.
- `DrawRun(segs, atlas, font, x, y, fontSize, tint) float32` — **`y` is the vertical centre of the line**. Text drawn with spacing `1`; emoji drawn at `fontSize*0.65` (not 0.85) with `0.25*size` left pad and +2 px after.
- `MeasureRun(segs, font, fontSize, emojiSize) float32` — adds `emojiSize*0.25` for **every** segment (text too) and text spacing `1`.
- `Atlas.Draw/Unload/find` (FE0F-tolerant lookup).

## `lib/raycss`

- Palette: `BorderColor`, `TextDark`, `TextMuted`.
- `Spacing` box model: `S`, `SXY(x,y)`, `STRBL`, `.Scale(zoom)`, `.H()`, `.V()`, `.Inner(rect)`, `.Outer(rect)`.
- `DrawHLine/DrawVLine`, `DrawSoftShadow(rect, roundness, segs)` (5 stacked translucent rounded rects, 4 px y-offset).
- `FitText(text, font, x, y, w, h, fontSize, padding, TextStyle) []Line` — word-wrap + vertical fit. Tries `fontSize` down to 70 % in 1-px steps; if still too tall, clips and truncates last line with `...`. Reshapes internally (per whole line, per word for measuring). `TextStyle`: `RTL`, `HAlign`/`VAlign` (`AlignStart|Center|End|Justify`), `Spacing`, `UseEmoji`, `IgnoreSpace`, `IgnoreNewLine`. `Line` = `Rect`, `FontSize`, shaped `Text`, `Segs` (emoji mode), `JustWords` (justify mode: draw each word at its own X).
- `MaskContainer` — render texture + GLSL 330 SDF rounded-rect shader (one shader compile per container). `NewMaskContainer(w,h,radius)`, `PreRender(bg, drawChildren)`, `Blit(x,y)`, `Draw` (= both), `Unload`. Children draw in **local coordinates**; can nest by pre-rendering inner masks before the outer `BeginTextureMode` and `Blit`ting inside (see professors board).

## `lib/video` (inactive in practice — do not modify)

`Player *mpv.Mpv` global, `MpvInit()` (idle, force-window, fullscreen, transparent-black bg, 400 ms sleep), `Resolution(path)` via `ffprobe`, `Unload()` (`stop`), `IsIdle()` (`idle-active`). Note: despite the older note, `mainMediaCard.go` and `main.go` **do** call it when a `Media` record is a video.
