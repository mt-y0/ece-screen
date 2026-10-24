package components

import (
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"mt-y0/ECE-screen/lib/emoji"
	lib "mt-y0/ECE-screen/lib/fonts"
	"mt-y0/ECE-screen/lib/raycss"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// loadTexture decodes any format supported by Go's image package (JPEG, PNG, GIF…)
// and uploads the pixels to the GPU. This avoids raylib's compile-time format flags.
func loadTexture(path string) (rl.Texture2D, bool) {
	f, err := os.Open(path)
	if err != nil {
		return rl.Texture2D{}, false
	}
	defer f.Close()

	src, _, err := image.Decode(f)
	if err != nil {
		return rl.Texture2D{}, false
	}

	b := src.Bounds()
	w, h := int32(b.Dx()), int32(b.Dy())

	rgba := image.NewRGBA(image.Rect(0, 0, int(w), int(h)))
	draw.Draw(rgba, rgba.Bounds(), src, b.Min, draw.Src)

	rlImg := &rl.Image{
		Data:    unsafe.Pointer(&rgba.Pix[0]),
		Width:   w,
		Height:  h,
		Mipmaps: 1,
		Format:  rl.UncompressedR8g8b8a8,
	}
	tex := rl.LoadTextureFromImage(rlImg)
	runtime.KeepAlive(rgba) // hold until C upload is done

	if tex.Width == 0 {
		return rl.Texture2D{}, false
	}
	return tex, true
}

type NoticeItem struct {
	Id       string
	Text     string
	ImgPath  string
	Duration int // seconds
}

type NoticeBoard struct {
	BgRect rl.Rectangle
	Mask   raycss.MaskContainer
	mu     sync.RWMutex
	dirty  atomic.Bool
	// protected by mu:
	data       []NoticeItem
	currentIdx int
	lastSwitch time.Time
	// render-goroutine only (no lock needed):
	Alpha      float32 // 0→255: content fade-in / white overlay fade-out
	loadedPath string
	currentTex rl.Texture2D
	hasTexture bool
}

func GetNoticeBoardLayout(screenSize [2]int32, zoom float64) *NoticeBoard {
	pad := int32(math.Round(15 * zoom))

	secH := int32(math.Round(float64(screenSize[1]) * 0.65 * 0.55))
	leftColW := int32(math.Round(float64(secH) * 1.454))
	mainH := int32(math.Round(float64(screenSize[1])*0.65)) - pad*2
	mainW := int32(math.Round(float64(mainH) * 1.454))

	// left-margin + leftCol + gap + mainMedia + right-margin = start of notice board
	bgX := float32(pad + leftColW + pad + mainW + pad)
	bgWidth := float32(screenSize[0]) - bgX - float32(pad)
	bgHeight := mainH

	nb := &NoticeBoard{
		BgRect: rl.Rectangle{
			X:      bgX,
			Y:      float32(math.Round(float64(screenSize[1])*0.3)) + float32(pad),
			Width:  bgWidth,
			Height: float32(bgHeight),
		},
		Mask:       raycss.NewMaskContainer(int32(bgWidth), bgHeight, cardRadius*float32(zoom)),
		lastSwitch: time.Now(),
		Alpha:      255,
	}
	nb.dirty.Store(true)
	return nb
}

func (l *NoticeBoard) GetData(pb *pocketbase.PocketBase) error {
	col, err := pb.FindCollectionByNameOrId("Notices")
	if err != nil {
		return err
	}

	records, err := pb.FindRecordsByFilter("Notices", "", "+display_order", 500, 0)
	if err != nil {
		return err
	}

	data := make([]NoticeItem, 0, len(records))
	for _, r := range records {
		dur := r.GetInt("duration")
		if dur <= 0 {
			dur = 10
		}
		imgPath := ""
		if imgFile := r.GetString("image"); imgFile != "" {
			imgPath = storagePath(pb, col.Id, r.Id, imgFile)
		}
		data = append(data, NoticeItem{
			Id:       r.Id,
			Text:     r.GetString("text"),
			ImgPath:  imgPath,
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

func (l *NoticeBoard) Watch(pb *pocketbase.PocketBase) {
	reload := func(e *core.RecordEvent) error {
		if err := l.GetData(pb); err != nil {
			log.Printf("notice reload: %v", err)
		}
		return e.Next()
	}
	pb.OnRecordAfterCreateSuccess("Notices").BindFunc(reload)
	pb.OnRecordAfterUpdateSuccess("Notices").BindFunc(reload)
	pb.OnRecordAfterDeleteSuccess("Notices").BindFunc(reload)
}

// cycle advances to the next notice and triggers a fade-in.
// Must only be called from the render goroutine.
func (l *NoticeBoard) cycle() {
	l.mu.Lock()
	if len(l.data) > 0 {
		l.currentIdx = (l.currentIdx + 1) % len(l.data)
	}
	l.lastSwitch = time.Now()
	l.mu.Unlock()
	l.Alpha = 0
}

// CleanUp unloads any GPU texture. Call before rl.CloseWindow.
func (l *NoticeBoard) CleanUp() {
	if l.hasTexture {
		rl.UnloadTexture(l.currentTex)
		l.hasTexture = false
	}
	l.Mask.Unload()
}

func (l *NoticeBoard) Draw(font *rl.Font, emojiAtlas *emoji.Atlas, zoom float64, blue, lblue rl.Color) {
	// ── cycling check ─────────────────────────────────────────────────────────
	l.mu.RLock()
	data := l.data
	idx := l.currentIdx
	lastSwitch := l.lastSwitch
	l.mu.RUnlock()

	if len(data) > 1 {
		dur := time.Duration(data[idx].Duration) * time.Second
		if time.Since(lastSwitch) >= dur {
			l.cycle()
			l.mu.RLock()
			data = l.data
			idx = l.currentIdx
			l.mu.RUnlock()
		}
	}

	// ── texture: load/swap when the current item's image changes ─────────────
	imgPath := ""
	if len(data) > 0 {
		imgPath = data[idx].ImgPath
	}
	if imgPath != l.loadedPath {
		if l.hasTexture {
			rl.UnloadTexture(l.currentTex)
			l.hasTexture = false
		}
		if imgPath != "" {
			if tex, ok := loadTexture(imgPath); ok {
				rl.SetTextureFilter(tex, rl.FilterBilinear)
				l.currentTex = tex
				l.hasTexture = true
			}
		}
		l.loadedPath = imgPath
	}

	// ── alpha fade-in: 0→255 over ~0.5 s ─────────────────────────────────────
	needsRender := l.dirty.Swap(false)
	if l.Alpha < 255 {
		needsRender = true
		l.Alpha += 510 * rl.GetFrameTime()
		if l.Alpha > 255 {
			l.Alpha = 255
		}
	}

	// ── shadow + card ─────────────────────────────────────────────────────────
	raycss.DrawSoftShadow(l.BgRect, cardRadius*float32(zoom), shadowSegs)
	if needsRender {
		l.Mask.PreRender(rl.White, func() {
			headerH := float32(math.Round(50.0 * zoom))
			titleFS := float32(math.Round(float64(headerH) * 0.8))
			bodyPad := raycss.S(10).Scale(zoom)
			imgMargin := raycss.SXY(0, 8).Scale(zoom)

			// ── header ────────────────────────────────────────────────────────────
			rl.DrawRectangle(0, 0, int32(l.BgRect.Width), int32(headerH), lblue)

			title := lib.Reshape(" 📢اعلانات القسم")
			seg := emoji.Split(title)
			ts := rl.MeasureTextEx(*font, title, titleFS, 1)
			emoji.DrawRun(seg, emojiAtlas, font, l.BgRect.Width/2-ts.X/2, headerH/2, titleFS, blue)

			rl.DrawRectangle(0, int32(headerH), int32(l.BgRect.Width), 5, blue)

			// ── body area ─────────────────────────────────────────────────────────
			bodyY := headerH + 5
			bodyH := l.BgRect.Height - headerH - 5

			rl.DrawRectangle(0, int32(bodyY), int32(l.BgRect.Width), int32(bodyH), rl.White)

			bodyFS := titleFS * 0.8

			if l.hasTexture {
				imgW := l.BgRect.Width - bodyPad.H()
				imgH := (bodyH-bodyPad.Top)/2 - imgMargin.V()
				imgX := bodyPad.Left
				imgY := bodyY + bodyPad.Top + imgMargin.Top

				// contain-fit: scale to fit within bounds, preserve ratio, centre
				srcW := float32(l.currentTex.Width)
				srcH := float32(l.currentTex.Height)
				scale := imgW / srcW
				if imgH/srcH < scale {
					scale = imgH / srcH
				}
				drawW := srcW * scale
				drawH := srcH * scale
				src := rl.NewRectangle(0, 0, srcW, srcH)
				dst := rl.NewRectangle(imgX+(imgW-drawW)/2, imgY+(imgH-drawH)/2, drawW, drawH)
				rl.DrawTexturePro(l.currentTex, src, dst, rl.Vector2{}, 0, rl.White)

				// text: lower half, VAlign start
				textY := imgY + imgH + imgMargin.Bottom + bodyPad.Top/2
				textH := bodyH - (imgH + imgMargin.V() + bodyPad.Top + bodyPad.Top/2)
				if len(data) > 0 {
					style := raycss.TextStyle{RTL: true, HAlign: raycss.AlignCenter, VAlign: raycss.AlignStart, Spacing: 0, UseEmoji: true}
					lines := raycss.FitText("📌"+data[idx].Text, font, 0, textY, l.BgRect.Width, textH, bodyFS, bodyPad.Top, style)
					for _, line := range lines {
						emoji.DrawRun(line.Segs, emojiAtlas, font, line.Rect.X, line.Rect.Y, line.FontSize, raycss.TextDark)
					}
				}
			} else {
				// no image: text fills full body, VAlign center
				if len(data) > 0 {
					style := raycss.TextStyle{RTL: true, HAlign: raycss.AlignCenter, VAlign: raycss.AlignCenter, Spacing: 0, UseEmoji: true}
					lines := raycss.FitText("📌"+data[idx].Text, font, 0, bodyY, l.BgRect.Width, bodyH, bodyFS, bodyPad.Top, style)
					for _, line := range lines {
						emoji.DrawRun(line.Segs, emojiAtlas, font, line.Rect.X, line.Rect.Y, line.FontSize, raycss.TextDark)
					}
				}
			}

			// white overlay fades out as Alpha rises → content fades in
			overlayAlpha := uint8(255 - l.Alpha)
			rl.DrawRectangle(0, int32(bodyY), int32(l.BgRect.Width), int32(bodyH),
				rl.NewColor(255, 255, 255, overlayAlpha))
		})
	}
	l.Mask.Blit(l.BgRect.X, l.BgRect.Y)
}
