package components

import (
	"log"
	"math"
	"strings"
	"sync"

	"mt-y0/ECE-screen/lib/emoji"
	lib "mt-y0/ECE-screen/lib/fonts"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type TickerBarLayout struct {
	BgRect   rl.Rectangle
	mu       sync.RWMutex
	segments []emoji.Segment
	textW    float32 // cached MeasureRun(segments); valid when widthOK
	widthOK  bool
	TextX    float32
	fontSize float32
	emojiSz  float32
}

func GetTickerBarLayout(screenSize [2]int32, zoom float64) *TickerBarLayout {
	barH := float32(math.Round(float64(screenSize[1]) * 0.05))
	fs := float32(math.Round(float64(barH) * 0.70))
	return &TickerBarLayout{
		BgRect: rl.Rectangle{
			X:      0,
			Y:      float32(screenSize[1]) - barH,
			Width:  float32(screenSize[0]),
			Height: barH,
		},
		fontSize: fs,
		emojiSz:  fs * 0.85,
	}
}

func (l *TickerBarLayout) GetData(pb *pocketbase.PocketBase) error {
	records, err := pb.FindAllRecords("TickerBar")
	if err != nil {
		return err
	}

	var b strings.Builder
	b.WriteString(" ~ ")
	for _, r := range records {
		b.WriteString(r.GetString("text"))
		b.WriteString(" ~ ")
	}
	segs := emoji.Split(lib.Reshape(b.String()))

	l.mu.Lock()
	l.segments = segs
	l.widthOK = false
	l.mu.Unlock()
	return nil
}

func (l *TickerBarLayout) Watch(pb *pocketbase.PocketBase) {
	reload := func(e *core.RecordEvent) error {
		if err := l.GetData(pb); err != nil {
			log.Printf("ticker reload: %v", err)
		}
		return e.Next()
	}
	pb.OnRecordAfterCreateSuccess("TickerBar").BindFunc(reload)
	pb.OnRecordAfterUpdateSuccess("TickerBar").BindFunc(reload)
	pb.OnRecordAfterDeleteSuccess("TickerBar").BindFunc(reload)
}

func (l *TickerBarLayout) Draw(screenSize [2]int32, font *rl.Font, atlas *emoji.Atlas, blue rl.Color) {
	rl.DrawRectangleRec(l.BgRect, blue)

	// vertical centre of the bar
	cy := l.BgRect.Y + l.BgRect.Height/2

	// Measure once per data change (not every frame); all reads under the lock.
	l.mu.Lock()
	segs := l.segments
	if !l.widthOK {
		l.textW = emoji.MeasureRun(segs, font, l.fontSize, l.emojiSz)
		l.widthOK = true
	}
	ts := l.textW
	l.mu.Unlock()

	if ts <= 0 {
		return
	}

	sw := float32(screenSize[0])

	// Two copies scroll right, exactly one screen width apart (start to start).
	// When a copy's start reaches the right corner it wraps to the left corner
	// while the other keeps the same distance. If the text is wider than the
	// screen the spacing grows to the text width so copies never overlap.
	period := sw
	if ts > period {
		period = ts
	}
	emoji.DrawRun(segs, atlas, font, l.TextX, cy, l.fontSize, rl.White)
	emoji.DrawRun(segs, atlas, font, l.TextX-period, cy, l.fontSize, rl.White)

	l.TextX += sw / 30.0 * rl.GetFrameTime()
	if l.TextX >= period {
		l.TextX -= period
	}
}
