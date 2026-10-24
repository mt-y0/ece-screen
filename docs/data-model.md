# Data model (PocketBase)

_Last updated: 2026-10-24_

DB lives in `pb_data/` (git-ignored) next to the working directory. Admin UI: `http://localhost:8090/_/`. Collections are **created automatically** on startup by `internal/db/schema.go` (`EnsureSchema`): any missing collection is created with its fields and filled with example records (Schedules for today, 2 ticker messages, 2 professors, 2 notices, 2 generated placeholder PNGs in Media). Existing collections/data are never modified. Rules are left superuser-only (the app reads through the in-process API). When you change a field in code, update `schema.go` too.

| Collection | Field | Type | Notes |
|---|---|---|---|
| `Schedules` | `day` | text | Arabic weekday exactly as in `arabicDays` (`tableArea.go`): `الاحد الاثنين الثلاثاء الاربعاء الخميس الجمعة السبت` (**no hamza** on الاحد/الاربعاء) |
| | `stage` | number | 1–4 (4 is leftmost card) |
| | `section` | text | `A`, `B`, `مسائي` |
| | `subject`, `hall` | text | hall is rendered as `قاعة <hall>` |
| | `start_time`, `end_time` | text | 12-hour `03:04 PM` (Go layout `"03:04 PM"`, zero-padded hour). Parse failure silently yields zero time |
| `TickerBar` | `text` | text | one message per record; order = DB default |
| `Professors` | `name`, `role`, `bio` | text | role is upper-cased on display; Latin layout |
| | `image` | file | |
| | `display_order` | number | sort key |
| `Notices` | `text` | text | Arabic/emoji; prefixed with 📌 |
| | `image` | file | optional |
| | `duration` | number | seconds, ≤0 → 10 |
| | `display_order` | number | |
| `Media` | `file` | file | image or video (video by extension: mp4 avi mov mkv webm flv wmv m4v 3gp) |
| | `duration` | number | seconds for images, ≤0 → 10 |
| | `main_only` | bool | true → hidden from secondary card |
| | `display_order` | number | |

## File paths

Uploaded files are read straight from disk: `<pb.DataDir()>/storage/<collectionId>/<recordId>/<filename>` (via `storagePath` in `internal/components/paths.go`; a missing file is logged). `DataDir()` defaults to `pb_data/` in the process CWD, so run the binary from the directory containing `pb_data/`.

## Realtime

Hooks `OnRecordAfterCreate/Update/DeleteSuccess(<collection>)` call `GetData` again (full reload, not delta). `Media` is watched by two components (main + secondary).
