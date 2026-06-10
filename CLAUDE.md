# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
go build ./...          # build
go run .                # run (requires a display and MPV installed)
go vet ./...            # lint
```

There are no tests. The application requires a connected display and `libmpv` — it cannot run headlessly.

PocketBase admin UI is served at `http://localhost:8090/_/` while the app is running.

## Architecture

**ECE-screen** is a university department display board: a fullscreen 60 fps raylib window that shows today's lecture schedule, a scrolling ticker, and a notice board. All data lives in an embedded PocketBase SQLite database (`pb_data/`); PocketBase starts as a goroutine inside the process.

### Data flow

```
PocketBase (goroutine, :8090)
    ↓ initial load + live hooks
components (sync.RWMutex)
    ↓ each frame
Raylib render loop (main goroutine)
```

Each component owns a `sync.RWMutex`: PocketBase event hooks (`OnRecordAfterCreateSuccess` etc.) write under a write-lock; `Draw()` reads under a read-lock.

### Key packages

| Package | Purpose |
|---|---|
| `lib/fonts` | Custom Arabic shaper — `Reshape()` converts raw Unicode Arabic to FE-range presentation forms, applies lam-alef ligatures, and reverses bidi runs so raylib's LTR renderer displays it correctly. `LoadFont()` bakes NotoKufi TTF with all Arabic codepoint ranges. |
| `lib/emoji` | Twemoji PNG atlas loader. `Split()` tokenises already-shaped text into `[]Segment` (plain text runs or single emoji). `DrawRun()` renders a segment slice. `MeasureRun()` measures without drawing. |
| `lib/raycss` | CSS-like primitives: palette colours, `DrawHLine`/`DrawVLine`, `DrawSoftShadow`, `MaskContainer` (SDF shader rounded-corner clipping), and `FitText` (Arabic-aware word-wrap returning `[]Line`). |
| `lib/video` | Thin `go-mpv` wrapper; MPV instance is created at startup but video playback is currently unused — ignore it for now. |
| `internal/db` | Starts the embedded PocketBase server. If no CLI args are given, defaults to `serve --http=0.0.0.0:8090`. |
| `internal/components` | Three UI components: `TableLayout` (4 stage cards), `TickerBarLayout` (scrolling footer), `NoticeBoard` (Arabic body text). |

### Critical invariants

1. **Arabic text pipeline**: raw Arabic → `lib.Reshape()` → `rl.DrawTextEx`. Never pass raw Arabic directly to raylib — glyphs will not connect.

2. **`FitText` reshapes internally**: do _not_ pre-`Reshape` text passed to `raycss.FitText`; it calls `lib.Reshape` on each word internally so cross-word contextual forms are correct.

3. **`MaskContainer` local coordinates**: the `drawChildren` callback passed to `mask.Draw()` must use `(0, 0)` as the card origin — the SDF shader clips in texture space, not screen space.

4. **Spacing consistency**: `TextStyle.Spacing` in `FitText` must equal the `spacing` argument later passed to `rl.DrawTextEx`. Arabic text uses `0`; mismatching causes line widths to be wrong and text to overflow.

5. **Emoji sizing**: `emoji.DrawRun` internally sizes emoji at `fontSize * 0.85`. `emoji.MeasureRun` takes an explicit `emojiSize` — pass `fontSize * 0.85` to match.

### Ignored / dead code

- `lib/video` and the `m := video.MpvInit()` call in `main.go` are not in active use — do not modify or remove them, just leave them alone.
- The commented-out blend mode block in `main.go` (`BeginBlendMode` / `SetBlendFactorsSeparate` / `EndBlendMode`) is intentionally left as a reference — do not uncomment or delete it.

### PocketBase collections (inferred from code)

- `Schedules` — fields: `day` (Arabic weekday string), `stage` (1–4), `section` (`"A"`, `"B"`, `"مسائي"`), `subject`, `hall`, `start_time`, `end_time` (12-hour format `"03:04 PM"`)
- `TickerBar` — field: `text`
