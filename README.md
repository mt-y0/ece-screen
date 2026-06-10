# ECE-screen

A fullscreen 60 fps display board for a university Electrical & Computer Engineering department. It shows today's lecture schedule across four stages, a live-updating scrolling ticker, professor boards, media cards, and a notice board — all with first-class Arabic text support.

Built with **Go**, **raylib-go**, and an embedded **PocketBase** SQLite database.

---

## Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [Dependencies](#dependencies)
- [Setup & Running](#setup--running)
- [Environment Variables](#environment-variables)
- [Data Model (PocketBase Collections)](#data-model-pocketbase-collections)
- [Project Structure](#project-structure)
- [Arabic Text Pipeline](#arabic-text-pipeline)
- [Live Data Updates](#live-data-updates)

---

## Features

- Fullscreen raylib window targeting 60 fps (configurable)
- **Table layout** — four stage cards showing today's schedule sections (A, B, مسائي) with start/end times, subject, and hall
- **Ticker bar** — horizontally scrolling footer fed from the database
- **Professors board** — dedicated board for professor information
- **Media cards** — secondary and main media slots (images / video via MPV)
- **Notice board** — Arabic body text with word-wrap and bidi rendering
- **Real-time updates** — PocketBase event hooks push changes to all components without restarting
- **Arabic shaping** — custom Unicode → FE-range presentation-form converter with lam-alef ligatures and bidi reversal
- **Emoji support** — Twemoji PNG atlas rendered inline alongside text
- **Rounded-corner clipping** — SDF shader via a `MaskContainer` primitive

---

## Architecture

```
PocketBase (goroutine, :8090)
    │
    │  initial load + live OnRecord* hooks
    ▼
components (each guarded by sync.RWMutex)
    │
    │  every frame
    ▼
Raylib render loop (main goroutine)
```

PocketBase starts as a goroutine inside the same process and serves its admin UI on port 8090. Each UI component holds its own `sync.RWMutex`; PocketBase event hooks write under a write-lock while `Draw()` reads under a read-lock, keeping the render loop wait-free in the common case.

---

## Dependencies

| Dependency | Purpose |
|---|---|
| [`raylib-go`](https://github.com/gen2brain/raylib-go) | 2D/3D rendering, window management |
| [`pocketbase`](https://github.com/pocketbase/pocketbase) | Embedded SQLite database + REST/realtime API |
| [`go-mpv`](https://github.com/gen2brain/go-mpv) | MPV video player bindings (initialised; playback reserved for future use) |

**System requirements**

- A connected display (raylib requires a real window)
- `libmpv` shared library installed
- Go 1.22+

On Arch Linux / Manjaro:
```bash
sudo pacman -S mpv
```

On Ubuntu/Debian:
```bash
sudo apt install libmpv-dev
```

---

## Setup & Running

```bash
# Clone
git clone https://github.com/mt-y0/ece-screen.git
cd ece-screen

# Build
go build ./...

# Run (requires display + libmpv)
go run .
```

The PocketBase admin UI is available at **http://localhost:8090/_/** while the app is running. Use it to seed and manage all data.

Press **X** to exit the application.

---

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `FPS` | `60` | Target render frame rate |
| `ZOOM` | `1.0` | Global UI scale factor |

Example:
```bash
FPS=30 ZOOM=1.25 go run .
```

---

## Data Model (PocketBase Collections)

### `Schedules`

Holds the per-stage lecture schedule.

| Field | Type | Notes |
|---|---|---|
| `day` | string | Arabic weekday name (e.g. `الأحد`) |
| `stage` | number | 1–4 |
| `section` | string | `"A"`, `"B"`, or `"مسائي"` |
| `subject` | string | Subject name (Arabic or Latin) |
| `hall` | string | Room/hall identifier |
| `start_time` | string | 12-hour format, e.g. `"08:00 AM"` |
| `end_time` | string | 12-hour format, e.g. `"10:00 AM"` |

### `TickerBar`

Each record contributes one message to the scrolling ticker.

| Field | Type | Notes |
|---|---|---|
| `text` | string | Plain or Arabic text to scroll |

### Other collections

Additional collections (professors, notices, media) are managed via the PocketBase admin UI at `http://localhost:8090/_/`.

---

## Project Structure

```
.
├── main.go                        # Entry point: init PocketBase, raylib window, render loop
├── go.mod / go.sum
│
├── internal/
│   ├── db/
│   │   └── client.go              # Starts embedded PocketBase server
│   └── components/
│       ├── tableArea.go           # 4-stage schedule card grid
│       ├── tickerBar.go           # Scrolling footer
│       ├── noticeBoard.go         # Arabic notice board
│       ├── professorsBoard.go     # Professor information board
│       ├── mainMediaCard.go       # Primary media slot (image/video)
│       └── secMediaCard.go        # Secondary media slot
│
├── lib/
│   ├── fonts/
│   │   └── arabic.go              # Arabic shaper (Reshape, LoadFont)
│   ├── emoji/
│   │   └── emoji.go               # Twemoji atlas loader, Split, DrawRun, MeasureRun
│   ├── raycss/
│   │   └── ray.go                 # CSS-like primitives: palette, shadows, MaskContainer, FitText
│   └── video/
│       └── video.go               # go-mpv thin wrapper (reserved for future use)
│
└── resources/
    └── fonts/
        ├── notokufi.ttf           # Noto Kufi Arabic font (embedded via go:embed)
        └── twemoji/               # 72×72 Twemoji PNG files for emoji rendering
```

---

## Arabic Text Pipeline

Raw Arabic Unicode must never be passed directly to raylib — glyphs will not connect. The correct pipeline is:

```
raw Arabic string
      │
      ▼
lib/fonts.Reshape()          ← converts to FE-range presentation forms,
      │                         applies lam-alef ligatures, reverses bidi runs
      ▼
rl.DrawTextEx(font, shaped, ...)
```

**Important:** Do **not** pre-reshape text before passing it to `raycss.FitText` — that function calls `Reshape` internally on each word so cross-word contextual forms are computed correctly.

---

## Live Data Updates

Every component registers PocketBase event hooks (`OnRecordAfterCreateSuccess`, `OnRecordAfterUpdateSuccess`, `OnRecordAfterDeleteSuccess`) via its `Watch(pb)` method. When a record changes, the hook acquires a write-lock, refreshes the component's internal state, and releases the lock. The render loop only ever takes read-locks, so updates are non-blocking from the frame's perspective.

---

## License

MIT
