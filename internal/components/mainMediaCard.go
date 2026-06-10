package components

import (
	"fmt"
	"log"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"mt-y0/ECE-screen/lib/emoji"
	"mt-y0/ECE-screen/lib/raycss"
	"mt-y0/ECE-screen/lib/video"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type MainMediaItem struct {
	FilePath string
	IsVideo  bool
	Duration int
}

type MainMediaCard struct {
	BgRect     rl.Rectangle
	Mask       raycss.MaskContainer
	screenSize [2]int32
	mu         sync.RWMutex
	dirty      atomic.Bool
	data       []MainMediaItem
	currentIdx int
	lastSwitch time.Time
	Alpha      float32
	loadedPath string
	currentTex rl.Texture2D
	hasTexture bool
	IsVideo    bool
}

func GetMainMediaLayout(screenSize [2]int32, zoom float64) *MainMediaCard {
	margin := raycss.S(15).Scale(zoom)

	bgHeight := int32(math.Round(float64(screenSize[1])*0.65)) - int32(margin.V())
	bgWidth := int32(math.Round(float64(bgHeight) * 1.454))

	secH := int32(math.Round(float64(screenSize[1]) * 0.65 * 0.55))
	bgX := float32(math.Round(float64(secH)*1.454)) + margin.H()
	bgY := float32(math.Round(float64(screenSize[1])*0.3)) + margin.Top

	mc := &MainMediaCard{
		BgRect: rl.Rectangle{
			X:      bgX,
			Y:      bgY,
			Width:  float32(bgWidth),
			Height: float32(bgHeight),
		},
		Mask:       raycss.NewMaskContainer(bgWidth, bgHeight, cardRadius*float32(zoom)),
		screenSize: screenSize,
		lastSwitch: time.Now(),
		Alpha:      255,
	}
	mc.dirty.Store(true)
	return mc
}

func (l *MainMediaCard) playVideo(path string) {
	cardW := int32(l.BgRect.Width)
	cardH := int32(l.BgRect.Height)
	x := int32(l.BgRect.X)
	y := int32(l.BgRect.Y)
	sw := l.screenSize[0]
	sh := l.screenSize[1]

	drawW, drawH := cardW, cardH
	offX, offY := int32(0), int32(0)

	if vw, vh, err := video.Resolution(path); err == nil && vw > 0 && vh > 0 {
		scale := float64(cardW) / float64(vw)
		if scaledH := float64(vh) * scale; scaledH > float64(cardH) {
			scale = float64(cardH) / float64(vh)
		}
		drawW = int32(float64(vw) * scale)
		drawH = int32(float64(vh) * scale)
		offX = (cardW - drawW) / 2
		offY = (cardH - drawH) / 2
	}

	vf := fmt.Sprintf("lavfi=[scale=%d:%d,pad=%d:%d:%d:%d:black@0]",
		drawW, drawH, sw, sh, x+offX, y+offY)
	_ = video.Player.SetPropertyString("vf", vf)
	_ = video.Player.SetPropertyString("fullscreen", "yes")
	_ = video.Player.Command([]string{"loadfile", path, "replace"})
	l.IsVideo = true
}

func (l *MainMediaCard) GetData(pb *pocketbase.PocketBase) error {
	col, err := pb.FindCollectionByNameOrId("Media")
	if err != nil {
		return err
	}

	records, err := pb.FindRecordsByFilter("Media", "", "+display_order", 500, 0)
	if err != nil {
		return err
	}

	data := make([]MainMediaItem, 0, len(records))
	for _, r := range records {
		file := r.GetString("file")
		if file == "" {
			continue
		}
		filePath := "pb_data/storage/" + col.Id + "/" + r.Id + "/" + file
		dur := r.GetInt("duration")
		if dur <= 0 {
			dur = 10
		}
		data = append(data, MainMediaItem{
			FilePath: filePath,
			IsVideo:  isVideo(file),
			Duration: dur,
		})
	}

	l.mu.Lock()
	l.data = data
	if l.currentIdx >= len(data) {
		l.currentIdx = 0
		l.lastSwitch = time.Now()
	}
	l.mu.Unlock()
	l.dirty.Store(true)

	return nil
}

func (l *MainMediaCard) Watch(pb *pocketbase.PocketBase) {
	reload := func(e *core.RecordEvent) error {
		if err := l.GetData(pb); err != nil {
			log.Printf("mainmedia reload: %v", err)
		}
		return e.Next()
	}
	pb.OnRecordAfterCreateSuccess("Media").BindFunc(reload)
	pb.OnRecordAfterUpdateSuccess("Media").BindFunc(reload)
	pb.OnRecordAfterDeleteSuccess("Media").BindFunc(reload)
}

func (l *MainMediaCard) cycle() {
	l.mu.Lock()
	if len(l.data) > 0 {
		l.currentIdx = (l.currentIdx + 1) % len(l.data)
	}
	l.lastSwitch = time.Now()
	l.mu.Unlock()
	l.Alpha = 0
}

func (l *MainMediaCard) CleanUp() {
	if l.hasTexture {
		rl.UnloadTexture(l.currentTex)
		l.hasTexture = false
	}
	l.Mask.Unload()
}

func (l *MainMediaCard) Draw(font *rl.Font, emojiAtlas *emoji.Atlas, zoom float64, blue, lblue rl.Color) {
	l.mu.RLock()
	data := l.data
	idx := l.currentIdx
	lastSwitch := l.lastSwitch
	l.mu.RUnlock()

	if len(data) > 0 {
		var cycleDur time.Duration
		if !data[idx].IsVideo {
			cycleDur = time.Duration(data[idx].Duration) * time.Second
		}
		if !l.IsVideo && (time.Since(lastSwitch) >= cycleDur) {
			l.cycle()
			l.mu.RLock()
			data = l.data
			idx = l.currentIdx
			l.mu.RUnlock()
		}
	}

	// load texture / trigger video when the item changes
	filePath := ""
	if len(data) > 0 {
		filePath = data[idx].FilePath
	}
	if filePath != l.loadedPath {
		if l.hasTexture {
			rl.UnloadTexture(l.currentTex)
			l.hasTexture = false
		}
		if filePath != "" {
			if data[idx].IsVideo {
				l.playVideo(filePath)
			} else {
				if tex, ok := loadTexture(filePath); ok {
					rl.SetTextureFilter(tex, rl.FilterBilinear)
					l.currentTex = tex
					l.hasTexture = true
				}
			}
		}
		l.loadedPath = filePath
	}

	needsRender := l.dirty.Swap(false)
	if l.Alpha < 255 {
		needsRender = true
		l.Alpha += 510 * rl.GetFrameTime()
		if l.Alpha > 255 {
			l.Alpha = 255
		}
	}

	if needsRender {
		l.Mask.PreRender(rl.White, func() {
			if len(data) > 0 && data[idx].IsVideo {
				rl.DrawRectangle(0, 0, int32(l.BgRect.Width), int32(l.BgRect.Height), rl.White)
			} else if l.hasTexture {
				rl.DrawRectangle(0, 0, int32(l.BgRect.Width), int32(l.BgRect.Height), rl.White)
				// contain-fit: scale to fit, preserve ratio, center
				srcW := float32(l.currentTex.Width)
				srcH := float32(l.currentTex.Height)
				scale := l.BgRect.Width / srcW
				if l.BgRect.Height/srcH < scale {
					scale = l.BgRect.Height / srcH
				}
				drawW := srcW * scale
				drawH := srcH * scale
				src := rl.NewRectangle(0, 0, srcW, srcH)
				dst := rl.NewRectangle(
					(l.BgRect.Width-drawW)/2,
					(l.BgRect.Height-drawH)/2,
					drawW, drawH,
				)
				rl.DrawTexturePro(l.currentTex, src, dst, rl.Vector2{}, 0, rl.White)
			} else {
				rl.DrawRectangle(0, 0, int32(l.BgRect.Width), int32(l.BgRect.Height), rl.White)
			}

			if l.Alpha < 255 {
				overlayAlpha := uint8(255 - l.Alpha)
				rl.DrawRectangle(0, 0, int32(l.BgRect.Width), int32(l.BgRect.Height),
					rl.NewColor(255, 255, 255, overlayAlpha))
			}
		})
	}

	raycss.DrawSoftShadow(l.BgRect, shadowRound*float32(zoom), shadowSegs)
	l.Mask.Blit(l.BgRect.X, l.BgRect.Y)
}
