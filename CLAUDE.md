# CLAUDE.md

_Last updated: 2026-10-24_

Entry point for Claude Code. **This file is an index; the details live in [`docs/`](docs/README.md).** (`AGENTS.md` is the same index for other agents.)

## Project in one paragraph

ECE-screen: fullscreen 60 fps raylib-go display board for a university ECE department — 4-stage lecture table, professors board, main/secondary media (images + mpv video), notice board, scrolling ticker; Arabic-first (custom shaper). Data in embedded PocketBase (`pb_data/`, admin at `:8090/_/`) with live hooks into the render loop. Needs a display + libmpv; cannot run headless.

```bash
make test   # vet + gofmt (no Go tests exist); also: make build / make run
make run                        # auto-fetches deps on fresh clone; needs display + libmpv
```

## Where to look

| Need | Doc |
|---|---|
| Overview, index, doc rules | [docs/README.md](docs/README.md) |
| Process model, data flow, screen geometry | [docs/architecture.md](docs/architecture.md) |
| Each UI component | [docs/components.md](docs/components.md) |
| `lib/fonts`, `lib/emoji`, `lib/raycss`, `lib/video` | [docs/libraries.md](docs/libraries.md) |
| PocketBase collections/fields | [docs/data-model.md](docs/data-model.md) |
| Build/run/env/setup/secrets | [docs/operations.md](docs/operations.md) |
| Invariants, do-not-touch, style | [docs/conventions.md](docs/conventions.md) |
| Known bugs & proposed fixes | [docs/suggestions.md](docs/suggestions.md) |
| History | [docs/CHANGELOG.md](docs/CHANGELOG.md) |

## Must-know invariants (full list in docs/conventions.md)

- Arabic: raw text → `lib.Reshape()` → `rl.DrawTextEx`; **but** never pre-`Reshape` text given to `raycss.FitText`.
- `MaskContainer` callbacks draw in local coords (0,0); `TextStyle.Spacing` must match `DrawTextEx` spacing.
- GL calls on the main goroutine only; hooks mutate state under the component `mu`.
- Do not modify `lib/video`, `video.MpvInit()`, or the commented blend-mode block in `main.go`.

## Mandatory documentation rules

1. **Any code change → update the affected `docs/*.md` in the same change.**
2. **Every update must carry a date:** bump `_Last updated: YYYY-MM-DD_` in each touched doc (and in this file if edited) and add a dated entry to `docs/CHANGELOG.md`.
3. Mark implemented items in `docs/suggestions.md` as `fixed (date)`; don't delete them.
4. Never commit secrets. `.env` is git-ignored and holds `GITHUB_TOKEN`; don't print it or put it in docs/commits.
