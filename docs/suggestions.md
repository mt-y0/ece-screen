# Suggestions: known issues & proposed fixes

_Last updated: 2026-10-24_

Found by code reading + `go vet` on 2026-10-24. Status: `open` / `fixed (date)` / `wontfix`. Severity: H/M/L.

| ID | Sev | Status | Where | Problem → Proposed fix |
|---|---|---|---|---|
| S-01 | H | wontfix (intentional, 2026-10-24) | `internal/db/client.go` | Binds `0.0.0.0:8090`: admin UI + API reachable from the whole LAN on a public display. → default to `127.0.0.1:8090` (env `PB_ADDR` to override); keep collection API rules locked. |
| S-02 | H | fixed (2026-10-24) | `internal/db/client.go`, `main.go` | `time.Sleep(50ms)` after `go app.Start()` is a race: `GetData` may run before PocketBase is bootstrapped, so the first load fails (only logged) and screens stay empty until a record changes. → fixed: `InitPocketBase` now waits on a ready channel closed from `OnServe` after `EnsureSchema`. |
| S-03 | H | fixed (2026-10-24) | `tableArea.go` `GetData` | Day filter is evaluated once at load: after midnight the board keeps showing yesterday's lectures until a Schedules record changes. → reload when the weekday changes (check in `Draw` alongside `lastMinute`). |
| S-04 | M | fixed (2026-10-24) | `tableArea.go` Draw (subject) | When the subject font is shrunk to fit, X is still computed from the *unscaled* width `subSz.X`, so shrunk text is left-shifted/overflows. → re-measure at `fs` and use that width. |
| S-05 | M | fixed (2026-10-24) | `tickerBar.go` Draw | (a) Wrap resets at `TextX >= screenW`, but the loop period is the text width → visible jump unless `ts == screenW`; (b) text shorter than the screen leaves a gap; (c) `MeasureRun` called every frame and reads `l.segments` **without the lock** (data race with hooks). → cache measured width when segments change (under lock); wrap at `ts` (pad/repeat short text). |
| S-06 | M | open | `mainMediaCard.go`, `main.go` | Single video item: after it ends, `cycleDur=0` makes `cycle()` run every frame (Alpha reset each frame, video never replays since `loadedPath` is unchanged). Single image cycles to itself and re-fades every N s. Also `IsIdle()` may report idle right after `loadfile` before playback begins. → skip cycle when `len(data)==1`; replay video explicitly; use mpv `file-loaded`/`end-file` events instead of polling idle. |
| S-07 | M | open | `lib/emoji/emoji.go`, `raycss/ray.go` | Measure ≠ draw: `DrawRun` emoji size 0.65 vs callers' 0.85; `MeasureRun` adds `0.25*emojiSize` for text segments too; both hard-code spacing 1 while Arabic callers use 0 (breaks invariant 4). Also `FitText` lines carry a *top* `Y` but `DrawRun` expects a *centre* `Y` (notice text is drawn half a line high). → make one `emojiSize` helper shared by both, take a `spacing` param, document/convert Y. Verify visually. |
| S-08 | M | open | all components | Textures decoded (`image.Decode`) and uploaded synchronously on the render thread when an item changes → frame hitch on large photos; decode failures are silent and not retried (`loadedPath` is set anyway). → pre-decode in a goroutine, upload on main thread; log failures. |
| S-09 | M | fixed (2026-10-24) | all components | Image paths hard-code `"pb_data/storage/…"` relative to CWD. → use `pb.DataDir()` (or `app.Dir`) and `filepath.Join`; log missing files. |
| S-10 | M | fixed (2026-10-24; hall-label edge case still open) | `tableArea.go` | `Parse12HourTime` ignores errors → a mistyped time becomes `0000-01-01 00:00`, the lecture is silently treated as already finished. `"8:00 AM"` (no leading zero) also fails the `03:04 PM` layout. → try `3:04 PM` too and log bad rows. When the first lecture doesn't fit the card height the "لا يوجد" (none) label is shown wrongly. |
| S-11 | M | partly fixed (ZOOM validated 2026-10-24; resize/layout open) | `main.go` | `ZOOM` unvalidated (0/negative); window size read once (no resize handling); layout fractions/aspect `1.454` hard-coded and overlap on non-16:9 screens. → validate; recompute layout on resize or document supported ratio. |
| S-12 | L | open | `raycss.NewMaskContainer` | Every container compiles its own identical shader (≥8 compiles). Shader is `#version 330` (fails on GLES/Raspberry Pi). → share one shader; provide `#version 100/300 es` fallback if targeting ARM boards. |
| S-13 | L | open | `lib/fonts/arabic.go` | No Persian/Kurdish letters or tatweel; alef maqsura Initial/Medial reuse the *final* glyph FEF0. → extend `Registry` if the department uses those letters. `Reshape` is pure → ideal for unit tests. |
| S-14 | L | partly fixed (Mask unload 2026-10-24; goroutine `log.Fatalf` left on purpose) | `noticeBoard.go` | `CleanUp` doesn't unload `Mask`. `db/client.go` uses `log.Fatalf` from a goroutine. |
| S-15 | L | partly fixed (tests, `.env.example`, systemd unit, README 2026-10-24; CI/Dockerfile/PB migrations open) | repo | No tests, CI, `.env.example`, Dockerfile/systemd unit (kiosk autostart). (Schema bootstrap now exists in `internal/db/schema.go`, 2026-10-24.) → add `Reshape`/`FitText`/time-parse tests, commit PB migrations, add a systemd unit. README says "Go 1.22+" but go.mod requires 1.26.2. |
| S-16 | L | fixed (2026-10-24) | `main.go` | raylib logged ~3000 INFO lines at startup (one pair per emoji PNG). → log level Warning by default; `RL_LOG=info` restores it. |
| S-17 | L | open | `lib/emoji/emoji.go` | The atlas is packed from 1514 PNGs on every start. → bake once into a single atlas PNG + index and fall back to packing. |
| S-18 | L | fixed (2026-10-24) | `lib/emoji/emoji.go` | A PNG that failed to load left a blank slot silently. → failures logged and counted. |
| S-19 | L | fixed (2026-10-24) | `main.go` | Twemoji dir was CWD-relative only, silently missing from other directories. → falls back to the executable dir; missing dir is logged. |

## Suggested order

Remaining: S-06 → S-07 → S-08 → S-11 (layout) → S-12 → S-13 → S-17 (each changes logic or needs a visual check).
