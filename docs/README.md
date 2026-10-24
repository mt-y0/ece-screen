# ECE-screen documentation index

_Last updated: 2026-10-24_

Start here. Every file below is a focused reference; read only what the task needs.

| File | Read it when you need… |
|---|---|
| [architecture.md](architecture.md) | process model, data flow, threading, render loop, screen layout geometry |
| [components.md](components.md) | per-widget behaviour (`internal/components/*`), cycling/fade logic, DB collections each one reads |
| [libraries.md](libraries.md) | `lib/fonts` (Arabic shaper), `lib/emoji`, `lib/raycss` (FitText, MaskContainer), `lib/video` |
| [data-model.md](data-model.md) | PocketBase collections and fields, file storage paths |
| [operations.md](operations.md) | build, run, env vars, system deps, PocketBase admin, deployment notes |
| [conventions.md](conventions.md) | critical invariants (Arabic pipeline, spacing, mask coords), code idioms, do-not-touch list |
| [suggestions.md](suggestions.md) | known bugs and proposed fixes, with status and dates |
| [CHANGELOG.md](CHANGELOG.md) | dated log of every change to code or docs |

## Documentation rules (apply to humans and LLMs)

1. **Any code change must update the affected doc file(s) in the same commit.**
2. **Every doc edit must bump the `_Last updated:_` date at the top of that file** (format `YYYY-MM-DD`) and add a line to [CHANGELOG.md](CHANGELOG.md).
3. When a suggestion in `suggestions.md` is implemented, mark it `fixed` with the date; don't delete it.
4. If code and docs disagree, the code wins — fix the docs and note it in the changelog.
