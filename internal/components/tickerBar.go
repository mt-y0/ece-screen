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

	l.mu.RLock()
	segs := l.segments
	l.mu.RUnlock()

	sw := float32(screenSize[0])

	drawAt := func(startX float32) {
		emoji.DrawRun(segs, atlas, font, startX, cy, l.fontSize, rl.White)
	}

	drawAt(l.TextX)
	ts := emoji.MeasureRun(l.segments, font, l.fontSize, l.emojiSz)
	if ts > sw {
		drawAt(l.TextX - ts)
	} else {
		drawAt(l.TextX - sw)
	}

	l.TextX += sw / 30.0 * rl.GetFrameTime()
	if l.TextX >= sw {
		l.TextX = 0
	}
}
