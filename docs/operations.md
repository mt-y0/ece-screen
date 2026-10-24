# Operations

_Last updated: 2026-10-24_

## Build / run / lint

```bash
go build ./...   # build
go vet ./...     # lint (clean as of 2026-10-24)
gofmt -l .       # formatting (noticeBoard.go currently unformatted)
go run .         # needs a display + libmpv (cannot run headless)
```

No automated tests exist. Press **X** to quit.

## System dependencies

`libmpv` (Arch: `mpv`; Debian/Ubuntu: `libmpv-dev`), `ffprobe` (from ffmpeg; used for video sizing), OpenGL 3.3+ (shader is `#version 330`), a real display.

## Environment variables (read in `main.go`)

| Var | Default | Notes |
|---|---|---|
| `FPS` | 60 | positive int |
| `ZOOM` | 1.0 | float; **not validated** (0/negative breaks layout) |

## CLI behaviour

`db.InitPocketBase`: with no CLI args it appends `serve --http=0.0.0.0:8090`. With any args they pass straight to PocketBase (so `go run . serve --http=127.0.0.1:8090` works; other args may make PB exit).

## First-time setup

1. `go run .` once → PocketBase creates `pb_data/` and prints a one-time superuser install link.
2. Open `http://localhost:8090/_/`, create the superuser and the collections in [data-model.md](data-model.md).
3. Add data; the screen updates live.

## Secrets / git

`.env` and `.env.*` are git-ignored. A GitHub token for pushes is stored in the local `.env` as `GITHUB_TOKEN` (never commit it; never paste it in docs or commits). Push with:

```bash
set -a; . ./.env; set +a
git push "https://x-access-token:${GITHUB_TOKEN}@github.com/<owner>/ece-screen.git" main
```
(or configure `gh auth`/a credential helper from the token). The app itself does not read `.env`.

## Working directory matters

`pb_data/` (images/videos) and `resources/fonts/twemoji` are resolved relative to CWD. The Arabic font is embedded in the binary via `go:embed`; the Twemoji PNGs are not.
