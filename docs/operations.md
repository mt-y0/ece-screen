# Operations

_Last updated: 2026-10-24_

## Makefile (preferred)

`make` shows a menu (TUI helpers in `scripts/tui.sh`; honours `NO_COLOR`). `make build` / `make run` first detect a fresh clone (missing `.build/deps.stamp`, or `go.mod`/`go.sum` newer), run a system-dependency check (`make doctor`) and `go mod download`, then build `./ECE-screen`. `make test` = `go vet` + `gofmt -l` check + `go test ./...`. Also: `setup`, `vet`, `fmt`, `clean`, `distclean`.

## Build / run / lint

```bash
go build ./...   # build
go vet ./...     # lint (clean as of 2026-10-24)
gofmt -l .       # formatting (clean)
go run .         # needs a display + libmpv (cannot run headless)
```

Unit tests exist for `lib/fonts` (`Reshape`), `lib/emoji` (`Split`, nil atlas) and `Parse12HourTime`; nothing renders, so `go test ./...` runs headless. Press **X** to quit.

## System dependencies

`libmpv` (Arch: `mpv`; Debian/Ubuntu: `libmpv-dev`), `ffprobe` (from ffmpeg; used for video sizing), OpenGL 3.3+ (shader is `#version 330`), a real display.

## Environment variables (read in `main.go`)

| Var | Default | Notes |
|---|---|---|
| `FPS` | 60 | positive int |
| `ZOOM` | 1.0 | float > 0; invalid values are logged and ignored |
| `RL_LOG` | (warnings only) | `info` / `debug` restore raylib's verbose logging (default hides ~3000 emoji-load lines) |

`.env.example` lists these. For kiosk autostart see `deploy/ece-screen.service` (systemd user unit; install steps in its header).

## CLI behaviour

`db.InitPocketBase`: with no CLI args it appends `serve --http=0.0.0.0:8090`. With any args they pass straight to PocketBase (so `go run . serve --http=127.0.0.1:8090` works; other args may make PB exit).

## First-time setup

1. `go run .` with no `pb_data/` → PocketBase creates it, `EnsureSchema` creates all collections with example data ([data-model.md](data-model.md)), and PocketBase prints (and tries to open) a one-time registration URL `http://localhost:8090/_/#/pbinstall/<token>` (valid 30 min). Open it to create the first superuser. Alternative: `go run . superuser upsert EMAIL PASS`.
2. The screen already works with the example data; after registering, edit/replace it at `http://localhost:8090/_/` and it updates live.
3. The URL is only printed while no superuser exists; restart the app to get a fresh one.

## Secrets / git

`.env` and `.env.*` are git-ignored. A GitHub token for pushes is stored in the local `.env` as `GITHUB_TOKEN` (never commit it; never paste it in docs or commits). Push with:

```bash
set -a; . ./.env; set +a
git push "https://x-access-token:${GITHUB_TOKEN}@github.com/<owner>/ece-screen.git" main
```
(or configure `gh auth`/a credential helper from the token). The app itself does not read `.env`.

## Working directory matters

`pb_data/` (images/videos) is resolved relative to CWD (via `pb.DataDir()`); `resources/fonts/twemoji` is tried relative to CWD, then to the executable's directory. The Arabic font is embedded in the binary via `go:embed`; the Twemoji PNGs are not.
