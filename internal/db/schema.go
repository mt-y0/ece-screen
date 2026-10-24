package db

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"time"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
)

// Collection names — must match the names used by internal/components.
const (
	colSchedules  = "Schedules"
	colTickerBar  = "TickerBar"
	colProfessors = "Professors"
	colNotices    = "Notices"
	colMedia      = "Media"
)

// arabicDays mirrors components.arabicDays (indexed by time.Weekday).
// Keep both in sync: the Schedules `day` filter compares against these exact strings.
var arabicDays = [7]string{
	"الاحد", "الاثنين", "الثلاثاء", "الاربعاء", "الخميس", "الجمعة", "السبت",
}

type collectionSpec struct {
	name  string
	build func(c *core.Collection)
	seed  func(app core.App, c *core.Collection) error
}

// EnsureSchema creates every collection the display needs that does not exist
// yet and fills each newly created one with example records. Existing
// collections (and their data) are never touched, so it is safe on every start.
func EnsureSchema(app core.App) error {
	specs := []collectionSpec{
		{colSchedules, buildSchedules, seedSchedules},
		{colTickerBar, buildTickerBar, seedTickerBar},
		{colProfessors, buildProfessors, seedProfessors},
		{colNotices, buildNotices, seedNotices},
		{colMedia, buildMedia, seedMedia},
	}

	for _, s := range specs {
		if existing, _ := app.FindCollectionByNameOrId(s.name); existing != nil {
			continue
		}
		c := core.NewBaseCollection(s.name)
		s.build(c)
		if err := app.Save(c); err != nil {
			return fmt.Errorf("create collection %s: %w", s.name, err)
		}
		if err := s.seed(app, c); err != nil {
			return fmt.Errorf("seed collection %s: %w", s.name, err)
		}
		log.Printf("db: created collection %q with example data", s.name)
	}
	return nil
}

func ptr[T any](v T) *T { return &v }

func insert(app core.App, c *core.Collection, rows []map[string]any) error {
	for _, row := range rows {
		r := core.NewRecord(c)
		for k, v := range row {
			r.Set(k, v)
		}
		if err := app.Save(r); err != nil {
			return err
		}
	}
	return nil
}

// ── Schedules ────────────────────────────────────────────────────────────────

func buildSchedules(c *core.Collection) {
	c.Fields.Add(
		&core.TextField{Name: "day", Required: true},
		&core.NumberField{Name: "stage", Required: true, OnlyInt: true, Min: ptr(1.0), Max: ptr(4.0)},
		&core.TextField{Name: "section", Required: true}, // "A", "B", "مسائي"
		&core.TextField{Name: "subject", Required: true},
		&core.TextField{Name: "hall"},
		&core.TextField{Name: "start_time", Required: true}, // "03:04 PM"
		&core.TextField{Name: "end_time", Required: true},
	)
	c.AddIndex("idx_schedules_day", false, "`day`", "")
}

func seedSchedules(app core.App, c *core.Collection) error {
	today := arabicDays[int(time.Now().Weekday())]
	// End times are late so the examples stay visible all day (finished lectures are hidden).
	row := func(stage int, section, subject, hall, start, end string) map[string]any {
		return map[string]any{
			"day": today, "stage": stage, "section": section,
			"subject": subject, "hall": hall, "start_time": start, "end_time": end,
		}
	}
	return insert(app, c, []map[string]any{
		row(1, "A", "الرياضيات", "1", "08:30 AM", "11:30 PM"),
		row(1, "B", "الفيزياء", "2", "09:30 AM", "11:30 PM"),
		row(2, "A", "الدوائر الكهربائية", "3", "10:30 AM", "11:30 PM"),
		row(2, "B", "Digital Logic", "4", "11:30 AM", "11:30 PM"),
		row(3, "A", "الالكترونيات", "5", "08:30 AM", "11:30 PM"),
		row(3, "B", "الاشارات والانظمة", "6", "09:30 AM", "11:30 PM"),
		row(4, "A", "الاتصالات", "7", "10:30 AM", "11:30 PM"),
		row(4, "B", "السيطرة", "8", "11:30 AM", "11:30 PM"),
		row(1, "مسائي", "البرمجة", "9", "03:00 PM", "11:30 PM"),
		row(2, "مسائي", "الحاسبات", "10", "04:00 PM", "11:30 PM"),
		row(3, "مسائي", "الشبكات", "11", "05:00 PM", "11:30 PM"),
		row(4, "مسائي", "مشروع التخرج", "12", "06:00 PM", "11:30 PM"),
	})
}

// ── TickerBar ────────────────────────────────────────────────────────────────

func buildTickerBar(c *core.Collection) {
	c.Fields.Add(&core.TextField{Name: "text", Required: true})
}

func seedTickerBar(app core.App, c *core.Collection) error {
	return insert(app, c, []map[string]any{
		{"text": "مرحبا بكم في قسم الهندسة الكهربائية والحاسوب 👋"},
		{"text": "Example ticker message — edit it in the PocketBase admin UI 📢"},
	})
}

// ── Professors ───────────────────────────────────────────────────────────────

func buildProfessors(c *core.Collection) {
	c.Fields.Add(
		&core.TextField{Name: "name", Required: true},
		&core.TextField{Name: "role"},
		&core.TextField{Name: "bio"},
		&core.FileField{Name: "image", MaxSelect: 1, MaxSize: 10 << 20,
			MimeTypes: []string{"image/jpeg", "image/png", "image/gif"}},
		&core.NumberField{Name: "display_order", OnlyInt: true},
	)
}

func seedProfessors(app core.App, c *core.Collection) error {
	return insert(app, c, []map[string]any{
		{"name": "Dr. Example One", "role": "Head of Department", "display_order": 1,
			"bio": "Example biography. Replace this record from the admin UI and upload a portrait (ratio about 118x148)."},
		{"name": "Dr. Example Two", "role": "Lecturer", "display_order": 2,
			"bio": "Another example record showing how the professors board rotates every 20 seconds."},
	})
}

// ── Notices ──────────────────────────────────────────────────────────────────

func buildNotices(c *core.Collection) {
	c.Fields.Add(
		&core.TextField{Name: "text", Required: true},
		&core.FileField{Name: "image", MaxSelect: 1, MaxSize: 10 << 20,
			MimeTypes: []string{"image/jpeg", "image/png", "image/gif"}},
		&core.NumberField{Name: "duration", OnlyInt: true, Min: ptr(0.0)}, // seconds, <=0 → 10
		&core.NumberField{Name: "display_order", OnlyInt: true},
	)
}

func seedNotices(app core.App, c *core.Collection) error {
	return insert(app, c, []map[string]any{
		{"text": "هذا اعلان تجريبي. يمكنك تعديله من لوحة التحكم.", "duration": 10, "display_order": 1},
		{"text": "الامتحانات النصفية تبدأ الاسبوع القادم، يرجى مراجعة الجدول 📚", "duration": 10, "display_order": 2},
	})
}

// ── Media ────────────────────────────────────────────────────────────────────

func buildMedia(c *core.Collection) {
	c.Fields.Add(
		&core.FileField{Name: "file", Required: true, MaxSelect: 1, MaxSize: 1 << 30}, // image or video
		&core.NumberField{Name: "duration", OnlyInt: true, Min: ptr(0.0)},             // seconds (images), <=0 → 10
		&core.BoolField{Name: "main_only"},                                            // true → not shown on the secondary card
		&core.NumberField{Name: "display_order", OnlyInt: true},
	)
}

func seedMedia(app core.App, c *core.Collection) error {
	for i, col := range []color.RGBA{{0, 86, 179, 255}, {231, 241, 255, 255}} {
		data, err := placeholderPNG(1454, 1000, col)
		if err != nil {
			return err
		}
		f, err := filesystem.NewFileFromBytes(data, fmt.Sprintf("example-%d.png", i+1))
		if err != nil {
			return err
		}
		r := core.NewRecord(c)
		r.Set("file", f)
		r.Set("duration", 10)
		r.Set("main_only", false)
		r.Set("display_order", i+1)
		if err := app.Save(r); err != nil {
			return err
		}
	}
	return nil
}

// placeholderPNG returns a solid-colour PNG (1.454 aspect matches the media cards).
func placeholderPNG(w, h int, c color.RGBA) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = c.R, c.G, c.B, c.A
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
