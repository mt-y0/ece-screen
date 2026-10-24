package components

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	lib "mt-y0/ECE-screen/lib/fonts"
	"mt-y0/ECE-screen/lib/raycss"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// ── constants ──────────────────────────────────────────────────────────────

const (
	cardRadius  = float32(10)
	shadowRound = float32(0.05)
	shadowSegs  = int32(8)
)

var arabicDays = [7]string{
	"الاحد", "الاثنين", "الثلاثاء", "الاربعاء", "الخميس", "الجمعة", "السبت",
}

// ── static labels ─────────────────────────────────────────────────────────

var (
	shapedSectionA   = lib.Reshape("شعبةA ")
	shapedSectionB   = lib.Reshape("شعبةB ")
	shapedEvening    = lib.Reshape("الدراسة المسائية")
	shapedNoLectures = lib.Reshape("لا يوجد")
)

type secDef struct {
	key   string
	label string
}

var (
	morningSecs = []secDef{
		{key: "B", label: shapedSectionB},
		{key: "A", label: shapedSectionA},
	}
	eveningSecs = []secDef{
		{key: "مسائي", label: shapedEvening},
	}
)

// ── data types ─────────────────────────────────────────────────────────────

type classStruct struct {
	Id            string
	Stage         int
	Section       string
	ShapedSubject string
	ShapedHall    string
	StartTimeStr  string
	StartTime     time.Time
	EndTime       time.Time
}

type TableLayout struct {
	Stages     [4]rl.Rectangle
	Masks      [4]raycss.MaskContainer
	titles     [4]string
	mu         sync.RWMutex
	data       []classStruct
	dirty      atomic.Bool
	lastMinute int

	pb        *pocketbase.PocketBase // set by Watch; used for the midnight reload
	loadedDay atomic.Int32           // weekday the current data was loaded for
	reloading atomic.Bool
}

// ── helpers ────────────────────────────────────────────────────────────────

// Parse12HourTime parses "03:04 PM" as well as "3:04 PM" (case-insensitive).
func Parse12HourTime(s string) (time.Time, error) {
	s = strings.ToUpper(strings.TrimSpace(s))
	if t, err := time.Parse("03:04 PM", s); err == nil {
		return t, nil
	}
	return time.Parse("3:04 PM", s)
}

// ── init ───────────────────────────────────────────────────────────────────

func GetTableLayout(screenSize [2]int32, zoom float64) *TableLayout {
	l := &TableLayout{}

	padding := int32(math.Round(15 * zoom))
	stageW := int32(math.Round(float64(screenSize[0]-padding*5) / 4))
	stageH := int32(math.Round(float64(screenSize[1])*0.3)) - padding

	for i := range 4 {
		x := padding + (padding+stageW)*int32(i)
		l.Stages[i] = rl.Rectangle{
			X: float32(x), Y: float32(padding),
			Width: float32(stageW), Height: float32(stageH),
		}
		l.Masks[i] = raycss.NewMaskContainer(stageW, stageH, cardRadius)
		l.titles[i] = lib.Reshape(fmt.Sprintf("المرحلة%d  ", 4-i))
	}

	l.dirty.Store(true)
	l.lastMinute = -1
	l.loadedDay.Store(-1)
	return l
}

// ── data ───────────────────────────────────────────────────────────────────

func (l *TableLayout) GetData(pb *pocketbase.PocketBase) error {
	// Record the day first so a failing query is retried tomorrow / on the
	// next record change, not on every frame.
	weekday := int(time.Now().Weekday())
	l.loadedDay.Store(int32(weekday))
	records, err := pb.FindRecordsByFilter(
		"Schedules",
		fmt.Sprintf("day='%s'", arabicDays[weekday]),
		"", 300, 0,
	)
	if err != nil {
		return err
	}

	data := make([]classStruct, 0, len(records))
	for _, r := range records {
		start, errS := Parse12HourTime(r.GetString("start_time"))
		end, errE := Parse12HourTime(r.GetString("end_time"))
		if errS != nil || errE != nil {
			log.Printf("schedule %s: bad time (start=%q end=%q), row skipped",
				r.Id, r.GetString("start_time"), r.GetString("end_time"))
			continue
		}
		data = append(data, classStruct{
			Id:            r.Id,
			Stage:         r.GetInt("stage"),
			Section:       r.GetString("section"),
			ShapedSubject: lib.Reshape(r.GetString("subject")),
			ShapedHall:    lib.Reshape("قاعة " + r.GetString("hall") + " "),
			StartTimeStr:  r.GetString("start_time"),
			StartTime:     start,
			EndTime:       end,
		})
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].StartTime.Before(data[j].StartTime)
	})

	l.mu.Lock()
	l.data = data
	l.mu.Unlock()
	l.dirty.Store(true)

	return nil
}

func (l *TableLayout) Watch(pb *pocketbase.PocketBase) {
	l.pb = pb
	reload := func(e *core.RecordEvent) error {
		if err := l.GetData(pb); err != nil {
			log.Printf("table reload: %v", err)
		}
		return e.Next()
	}
	pb.OnRecordAfterCreateSuccess("Schedules").BindFunc(reload)
	pb.OnRecordAfterUpdateSuccess("Schedules").BindFunc(reload)
	pb.OnRecordAfterDeleteSuccess("Schedules").BindFunc(reload)
}

// ── draw ───────────────────────────────────────────────────────────────────

func (l *TableLayout) Draw(font *rl.Font, zoom float64, blue, lightBlue rl.Color) {
	l.mu.RLock()
	data := l.data
	l.mu.RUnlock()

	now := time.Now()
	nowH, nowM, _ := now.Clock()

	// The day filter is applied at load time, so reload when the weekday rolls over.
	if l.pb != nil && int32(now.Weekday()) != l.loadedDay.Load() &&
		l.reloading.CompareAndSwap(false, true) {
		go func() {
			defer l.reloading.Store(false)
			if err := l.GetData(l.pb); err != nil {
				log.Printf("table day reload: %v", err)
			}
		}()
	}

	// Dirty on minute boundary so expired lectures are removed promptly.
	nowMin := now.Hour()*60 + now.Minute()
	if l.lastMinute != nowMin {
		l.lastMinute = nowMin
		l.dirty.Store(true)
	}

	sections := eveningSecs
	if nowH < 14 {
		sections = morningSecs
	}

	headerH := float32(math.Round(48.0 * zoom))
	titleFS := float32(math.Round(36.0 * zoom))
	secNameH := float32(math.Round(34.0 * zoom))
	sLblFS := float32(math.Round(28.0 * zoom))
	timeFS := float32(math.Round(20.0 * zoom))
	subFS := float32(math.Round(24.0 * zoom))
	hallFS := float32(math.Round(18.0 * zoom))
	noLecFS := float32(math.Round(18.0 * zoom))
	cardPad := raycss.SXY(8, 10).Scale(zoom)

	// Only re-render the off-screen textures when content has changed.
	if l.dirty.Swap(false) {
		for i := range 4 {
			stage := l.Stages[i]
			stageNum := 4 - i
			l.Masks[i].PreRender(rl.White, func() {
				W := stage.Width
				H := stage.Height
				secW := W / float32(len(sections))

				// ── stage header ──────────────────────────────────────────────
				rl.DrawRectangle(0, 0, int32(W), int32(headerH), blue)
				ts := rl.MeasureTextEx(*font, l.titles[i], titleFS, 0)
				rl.DrawTextEx(*font, l.titles[i],
					rl.Vector2{X: W/2 - ts.X/2, Y: headerH/2 - ts.Y/2},
					titleFS, 0, rl.White)

				// ── sections ──────────────────────────────────────────────────
				for j, sec := range sections {
					x0 := float32(j) * secW

					rl.DrawRectangle(int32(x0), int32(headerH), int32(secW), int32(secNameH), lightBlue)
					slSz := rl.MeasureTextEx(*font, sec.label, sLblFS, 0)
					rl.DrawTextEx(*font, sec.label,
						rl.Vector2{
							X: x0 + secW/2 - slSz.X/2,
							Y: headerH + secNameH/2 - slSz.Y/2,
						}, sLblFS, 0, blue)

					if j > 0 {
						raycss.DrawVLine(x0, headerH, H-headerH, raycss.BorderColor)
					}

					// ── lectures ──────────────────────────────────────────────
					y := headerH + secNameH + cardPad.Top
					hasAny := false

					for _, lec := range data {
						if lec.Stage != stageNum || lec.Section != sec.key {
							continue
						}
						// skip past lectures
						endH, endM, _ := lec.EndTime.Clock()
						if endH < nowH || (endH == nowH && endM <= nowM) {
							continue
						}
						// clip before drawing
						if y+subFS+hallFS+cardPad.V() > H {
							break
						}
						hasAny = true

						// time — right-aligned
						tSz := rl.MeasureTextEx(*font, lec.StartTimeStr, timeFS, 0)
						rl.DrawTextEx(*font, lec.StartTimeStr,
							rl.Vector2{X: x0 + secW - cardPad.Right - tSz.X, Y: y},
							timeFS, 0, blue)
						y += tSz.Y + 2

						// subject — scale down if too wide, right-aligned
						subSz := rl.MeasureTextEx(*font, lec.ShapedSubject, subFS, 0)
						fs := subFS
						if subSz.X > secW-cardPad.H() {
							fs = subFS * (secW - cardPad.H()) / subSz.X
							if fs < 10 {
								fs = 10
							}
						}
						if fs != subFS {
							subSz = rl.MeasureTextEx(*font, lec.ShapedSubject, fs, 0)
						}
						rl.DrawTextEx(*font, lec.ShapedSubject,
							rl.Vector2{X: x0 + secW - cardPad.Right - subSz.X, Y: y},
							fs, 0, raycss.TextDark)
						y += subFS + 2

						// hall — right-aligned
						hSz := rl.MeasureTextEx(*font, lec.ShapedHall, hallFS, 0)
						rl.DrawTextEx(*font, lec.ShapedHall,
							rl.Vector2{X: x0 + secW - cardPad.Right - hSz.X, Y: y},
							hallFS, 0, raycss.TextMuted)
						y += hallFS + 4

						raycss.DrawHLine(x0+cardPad.Left, y, secW-cardPad.H(), raycss.BorderColor)
						y += cardPad.Bottom
					}

					if !hasAny {
						nSz := rl.MeasureTextEx(*font, shapedNoLectures, noLecFS, 0)
						rl.DrawTextEx(*font, shapedNoLectures,
							rl.Vector2{
								X: x0 + secW/2 - nSz.X/2,
								Y: headerH + secNameH + cardPad.Top,
							}, noLecFS, 0, raycss.TextMuted)
					}
				}
			})
		}
	}

	// Always blit the cached textures to the screen.
	for i := range 4 {
		raycss.DrawSoftShadow(l.Stages[i], shadowRound*float32(zoom), shadowSegs)
		l.Masks[i].Blit(l.Stages[i].X, l.Stages[i].Y)
	}
}

// ── cleanup ────────────────────────────────────────────────────────────────

func (l *TableLayout) CleanUp() {
	for i := range 4 {
		l.Masks[i].Unload()
	}
}
