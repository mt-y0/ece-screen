# Changelog

_Last updated: 2026-10-24_

Newest first. Every code or doc change gets a dated entry (`YYYY-MM-DD`).

## 2026-10-24 (second update)
- `internal/db`: new `schema.go` — `EnsureSchema` creates missing collections (Schedules, TickerBar, Professors, Notices, Media) with example records; `client.go` now runs it from `OnServe` and waits until ready (fixes S-02). PocketBase's built-in installer URL provides first-run superuser registration.
- Docs updated: data-model, operations, architecture, suggestions (S-02 fixed, S-15 amended).

## 2026-10-24
- Added `docs/` (architecture, components, libraries, data-model, operations, conventions, suggestions, this changelog).
- Rewrote `CLAUDE.md` and added `AGENTS.md` as index entries to `docs/`, with the rule "update docs + date on every change".
- Corrected stale notes: 6 components (not 3), `lib/video` is used by main media, emoji draw size is 0.65 (not 0.85).
- Recorded initial review findings S-01…S-15 in `suggestions.md` (no code changed).
