# Conventions & invariants

_Last updated: 2026-10-24_

## Critical invariants

1. **Arabic pipeline:** raw Arabic → `lib.Reshape()` → `rl.DrawTextEx`. Never pass raw Arabic to raylib.
2. **`FitText` reshapes internally:** do not pre-`Reshape` text passed to it.
3. **`MaskContainer` callbacks use local coordinates** `(0,0)` = card origin (SDF clips in texture space).
4. **Spacing consistency:** `TextStyle.Spacing` must equal the spacing later passed to `rl.DrawTextEx`. Arabic uses `0`. ⚠ The emoji path (`DrawRun`/`MeasureRun`) hard-codes spacing `1` — see [suggestions.md](suggestions.md) S-07.
5. **Emoji sizing:** `MeasureRun` takes an explicit `emojiSize`; callers pass `fontSize*0.85`. ⚠ `DrawRun` actually uses `fontSize*0.65` (the older docs said 0.85) — S-07.
6. **Threading:** raylib/GL calls only on the main goroutine. PocketBase hooks may only touch component state under `mu`/atomics, never GL.
7. **Nested masks:** pre-render inner `MaskContainer`s *before* the outer `PreRender`, then `Blit` inside it.
8. `Reshape`d strings are for display only — never store or compare them.

## Do not touch

- `lib/video` and `video.MpvInit()` in `main.go`.
- The commented-out/odd blend-mode block in `main.go` (`BeginBlendMode`/`SetBlendFactorsSeparate`/`EndBlendMode`) is a deliberate reference; don't remove or "clean up".

## Code style

Match surrounding code: `gofmt`, section banner comments (`// ── name ───`), `zoom`-scaled sizes via `math.Round(N * zoom)` or `raycss.Spacing.Scale`, one file per component, constructor `Get<Name>Layout`. Errors from `GetData` are logged, not fatal.

## Documentation discipline

- Update the relevant `docs/*.md` in the same change as the code.
- Bump `_Last updated: YYYY-MM-DD_` in every doc touched, and append to `docs/CHANGELOG.md`.
- Commit messages: imperative, what+why.
