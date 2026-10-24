# Changelog

_Last updated: 2026-10-24_

Newest first. Every code or doc change gets a dated entry (`YYYY-MM-DD`).

## 2026-10-24 (fifth update)
- `suggestions.md` only (no code): recorded analysis notes for S-06/07/08/13/17, closed S-11 (always 16:9) and S-12 (negligible gain) as wontfix, added S-20 (downscale images on load) and S-21 (smaller emoji set), parked the rest until real-display benchmarking.

## 2026-10-24 (fourth update)
- Fixed S-03 (weekday rollover reloads Schedules), S-04 (shrunk subject re-measured), S-05 (ticker: width cached under lock, seamless loop, repeats short text), S-09 (`storagePath` via `pb.DataDir()`, logs missing files), S-10 (`Parse12HourTime` returns an error, accepts `8:00 AM`; bad rows are logged and skipped), S-14 partly (`NoticeBoard.CleanUp` unloads `Mask`), S-16/S-18/S-19 (raylib log level Warning by default + `RL_LOG`, emoji load failures logged, twemoji path falls back to the executable dir).
- `ZOOM` validated (part of S-11). `emoji.Atlas` methods are now really nil-safe (found by a new test).
- S-01 closed as intentional: the server stays bound to `0.0.0.0:8090` so LAN devices can reach the admin UI.
- Added tests (`lib/fonts`, `lib/emoji`, `internal/components`), `.env.example`, `deploy/ece-screen.service`; README Go version corrected.
- Docs updated: components, data-model, libraries, operations, suggestions, AGENTS.md.

## 2026-10-24 (third update)
- Added `Makefile` (help/setup/doctor/build/run/vet/test/fmt/clean/distclean) and `scripts/tui.sh` (playful spinner/banner UI); fresh-clone detection via `.build/deps.stamp` triggers dependency check + `go mod download`. `.build/` git-ignored.
- `gofmt` applied to `internal/components/noticeBoard.go`.
- Docs updated: operations.

## 2026-10-24 (second update)
- `internal/db`: new `schema.go` — `EnsureSchema` creates missing collections (Schedules, TickerBar, Professors, Notices, Media) with example records; `client.go` now runs it from `OnServe` and waits until ready (fixes S-02). PocketBase's built-in installer URL provides first-run superuser registration.
- Docs updated: data-model, operations, architecture, suggestions (S-02 fixed, S-15 amended).

## 2026-10-24
- Added `docs/` (architecture, components, libraries, data-model, operations, conventions, suggestions, this changelog).
- Rewrote `CLAUDE.md` and added `AGENTS.md` as index entries to `docs/`, with the rule "update docs + date on every change".
- Corrected stale notes: 6 components (not 3), `lib/video` is used by main media, emoji draw size is 0.65 (not 0.85).
- Recorded initial review findings S-01…S-15 in `suggestions.md` (no code changed).
