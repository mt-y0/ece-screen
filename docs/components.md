# Components (`internal/components`)

_Last updated: 2026-10-24_

All components live in package `components`. Shared constants in `tableArea.go`: `cardRadius=10`, `shadowRound=0.05`, `shadowSegs=8`. Shared helpers: `loadTexture(path)` in `noticeBoard.go` (decodes via Go `image` — jpeg/png/gif — and uploads RGBA to GPU), `isVideo(path)` in `secMediaCard.go` (extension check).

Common fields: `BgRect`, `Mask raycss.MaskContainer`, `mu sync.RWMutex`, `dirty atomic.Bool`, `data`, `currentIdx`, `lastSwitch`, `Alpha` (0→255 fade-in, driven at 510/s ≈ 0.5 s), `loadedPath/currentTex/hasTexture` (single cached GPU texture).

| File | Type | Collection | Behaviour |
|---|---|---|---|
| `tableArea.go` | `TableLayout` | `Schedules` | 4 stage cards (stage = `4 - i`). Header, section sub-header, per-lecture: start time, subject (shrunk to fit width), hall. Past lectures hidden; "لا يوجد" when empty. Sort by start time. |
| `tickerBar.go` | `TickerBarLayout` | `TickerBar` | Concatenates all `text` fields with ` ~ ` separators, reshaped + emoji-split. Scrolls rightwards at `screenW/30` px/s, drawing two copies for wrap-around. Blue bar, white text. |
| `professorsBoard.go` | `ProfessorsBoard` | `Professors` (sorted `display_order`) | Cycles every fixed 20 s (`profCycleDuration`). Cropped (cover-fit) photo in its own `imgMask`, name / uppercase role / justified bio. Latin text path (no Reshape of whole card; `FitText` reshapes internally). |
| `secMediaCard.go` | `SecondaryMediaCard` | `Media` where `main_only = false`, images only | Contain-fit image, per-item `duration` (default 10 s). Starts at index 1 so it is out of phase with main media. |
| `mainMediaCard.go` | `MainMediaCard` | `Media` (all) | Images (contain-fit) or video. Video: `playVideo()` computes size via `ffprobe`, builds an mpv `lavfi` scale+pad filter to position the video inside `BgRect` on the fullscreen mpv window, sets `IsVideo`. While a video plays the card is white and `main.go` punches a hole for it. Cycle advances when video goes idle. |
| `noticeBoard.go` | `NoticeBoard` | `Notices` (sorted `display_order`) | Header "📢اعلانات القسم", body = optional image (top half, contain-fit) + RTL centred text via `FitText` with `UseEmoji`. Text prefixed with 📌. Per-item `duration` (default 10 s). |

## Cycling & fade (identical in all rotating cards)

`Draw` → read-locked snapshot → if elapsed ≥ item duration, `cycle()` (index+1 mod n, reset `lastSwitch`, `Alpha=0`) → if item path changed, unload old texture and load new one **synchronously on the render thread** → if dirty or fading, `Mask.PreRender(...)` (fade = white overlay with alpha `255-Alpha`) → shadow + `Mask.Blit`.

Index is reset to 0 in `GetData` only when it falls out of range.

## Cleanup

Each component with GPU resources has `CleanUp()` (called with `defer` in `main.go`). `NoticeBoard.CleanUp` unloads the texture but not its `Mask`; `TickerBarLayout` has none (no GPU resources).

## Adding a new component

1. File `internal/components/<name>.go`, constructor `Get<Name>Layout(screenSize, zoom)`.
2. Implement `GetData(pb)`, `Watch(pb)`, `Draw(...)`, `CleanUp()` following the pattern above (copy `noticeBoard.go`).
3. Register in `main.go` (init, `GetData`, `Watch`, `defer CleanUp`, `Draw` in loop).
4. Update this file, `architecture.md` (layout table), `data-model.md`, and `CHANGELOG.md`.
