# Architecture

_Last updated: 2026-10-24_

**ECE-screen** is a fullscreen (60 fps, raylib) department display board: today's lecture schedule, professors, rotating media (images + video), a notice board and a scrolling ticker. Go + raylib-go + embedded PocketBase (SQLite). Module path: `mt-y0/ECE-screen`, Go `1.26.2` (go.mod).

## Process model

```
main goroutine (OS thread, required by raylib/OpenGL)
 ├─ db.InitPocketBase()  → goroutine runs app.Start() (serve 0.0.0.0:8090); blocks until OnServe has run EnsureSchema (creates missing collections + examples), 30 s timeout
 ├─ video.MpvInit()      → libmpv instance (own fullscreen window behind raylib's transparent one)
 ├─ rl.InitWindow(0,0)   → transparent, monitor-sized window
 ├─ build components, GetData(pb) once, Watch(pb) hooks
 └─ render loop (until window close or key X)

PocketBase goroutines
 └─ OnRecordAfter{Create,Update,Delete}Success hooks → component.GetData(pb)
```

## Data flow

```
PocketBase SQLite (pb_data/)
   │  initial GetData + live hooks (write-lock)
   ▼
component state (sync.RWMutex + atomic dirty flag)
   │  Draw() each frame (read-lock, copy slice header)
   ▼
MaskContainer off-screen texture (re-rendered only when dirty / fading)
   │  Blit with rounded-corner SDF shader
   ▼
screen
```

Pattern shared by all data components: `GetXLayout(screenSize, zoom)` → `GetData(pb)` → `Watch(pb)` → `Draw(...)` per frame → `CleanUp()` at exit. Hooks replace `data` wholesale (slice is never mutated in place), so readers can safely keep the slice header after unlocking.

Rendering optimisation: each card draws into its own `MaskContainer` only when `dirty` is set or a fade (`Alpha < 255`) is in progress; otherwise the cached texture is just blitted.

## Render loop order (main.go)

1. exit on `X` key; if `mainMedia.IsVideo && video.IsIdle()` → clear flag, `video.Unload()`
2. clear to `bgColor` (248,249,250)
3. draw: table → professors → secMedia → mainMedia → ticker → notice board
4. if a video is playing, punch a transparent hole over `mainMedia.BgRect` using a custom blend mode (`BlendCustomSeparate`) so the mpv window underneath shows through.

## Screen layout (fractions of screen H, `pad = 15*zoom`)

| Region | Position / size |
|---|---|
| Table (4 stage cards) | y=pad, h=0.30H−pad; 4 equal cards across width. Stage 4 is leftmost, stage 1 rightmost (RTL) |
| Professors board | x=pad, y=0.30H+pad, w=`secH*1.454`, h=`0.65H − secH − 3pad` where `secH = 0.65*0.55*H` |
| Secondary media | x=pad, bottom-aligned at 0.95H−pad, size `secH*1.454 × secH` |
| Main media | x=`secH*1.454 + pad`, y=0.30H+pad, h=`0.65H−2pad`, w=`h*1.454` |
| Notice board | x=`pad+leftColW+pad+mainW+pad` to screen right−pad, same y/h as main media |
| Ticker | bottom 5% of H, full width |

`1.454` is a hard-coded media aspect ratio. Layout is computed once at start; window resize is not handled.

## Time logic

- Table loads only today's schedules (`day` = Arabic weekday of load time, see data-model).
- Before 14:00 local → morning sections (B, A); from 14:00 → evening section (`مسائي`).
- Lectures whose `end_time` ≤ now are hidden; the table is re-rendered on every minute boundary.
