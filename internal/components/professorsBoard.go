package components

import (
	"log"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"mt-y0/ECE-screen/lib/emoji"
	"mt-y0/ECE-screen/lib/raycss"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

const profCycleDuration = 20 * time.Second

type ProfessorItem struct {
	Id      string
	Name    string
	Role    string
	Bio     string
	ImgPath string
}

type ProfessorsBoard struct {
	BgRect     rl.Rectangle
	Mask       raycss.MaskContainer
	imgMask    raycss.MaskContainer
	imgW       float32
	imgH       float32
	mu         sync.RWMutex
	dirty      atomic.Bool
	data       []ProfessorItem
	currentIdx int
	lastSwitch time.Time
	Alpha      float32
	loadedPath string
	currentTex rl.Texture2D
	hasTexture bool
}

func GetProfessorsBoardLayout(screenSize [2]int32, zoom float64) *ProfessorsBoard {
	margin := raycss.S(15).Scale(zoom)

	secH := int32(math.Round(float64(screenSize[1]) * 0.65 * 0.55))
	totalH := int32(math.Round(float64(screenSize[1]) * 0.65))
	bgHeight := totalH - secH - int32(margin.Top)*3
	bgWidth := int32(math.Round(float64(secH) * 1.454))

	imgH := float32(math.Round(float64(bgHeight) * 0.42))
	imgW := float32(math.Round(float64(imgH) * (118.0 / 148.0)))
	imgRadius := float32(math.Round(7 * zoom))

	bgX := margin.Left
	bgY := float32(math.Round(float64(screenSize[1])*0.3)) + margin.Top

	pb := &ProfessorsBoard{
		BgRect: rl.Rectangle{
			X:      bgX,
			Y:      bgY,
			Width:  float32(bgWidth),
			Height: float32(bgHeight),
		},
		Mask:       raycss.NewMaskContainer(bgWidth, bgHeight, cardRadius*float32(zoom)),
		imgMask:    raycss.NewMaskContainer(int32(imgW), int32(imgH), imgRadius),
		imgW:       imgW,
		imgH:       imgH,
		lastSwitch: time.Now(),
		Alpha:      255,
	}
	pb.dirty.Store(true)
	return pb
}

func (l *ProfessorsBoard) GetData(pb *pocketbase.PocketBase) error {
	col, err := pb.FindCollectionByNameOrId("Professors")
	if err != nil {
		return err
	}

	records, err := pb.FindRecordsByFilter("Professors", "", "+display_order", 500, 0)
	if err != nil {
		return err
	}

	data := make([]ProfessorItem, 0, len(records))
	for _, r := range records {
		imgPath := ""
		if imgFile := r.GetString("image"); imgFile != "" {
			imgPath = storagePath(pb, col.Id, r.Id, imgFile)
		}
		data = append(data, ProfessorItem{
			Id:      r.Id,
			Name:    r.GetString("name"),
			Role:    r.GetString("role"),
			Bio:     r.GetString("bio"),
			ImgPath: imgPath,
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

func (l *ProfessorsBoard) Watch(pb *pocketbase.PocketBase) {
	reload := func(e *core.RecordEvent) error {
		if err := l.GetData(pb); err != nil {
			log.Printf("professors reload: %v", err)
		}
		return e.Next()
	}
	pb.OnRecordAfterCreateSuccess("Professors").BindFunc(reload)
	pb.OnRecordAfterUpdateSuccess("Professors").BindFunc(reload)
	pb.OnRecordAfterDeleteSuccess("Professors").BindFunc(reload)
}

func (l *ProfessorsBoard) cycle() {
	l.mu.Lock()
	if len(l.data) > 0 {
		l.currentIdx = (l.currentIdx + 1) % len(l.data)
	}
	l.lastSwitch = time.Now()
	l.mu.Unlock()
	l.Alpha = 0
}

func (l *ProfessorsBoard) CleanUp() {
	if l.hasTexture {
		rl.UnloadTexture(l.currentTex)
		l.hasTexture = false
	}
	l.imgMask.Unload()
	l.Mask.Unload()
}

func (l *ProfessorsBoard) Draw(font *rl.Font, emojiAtlas *emoji.Atlas, zoom float64, blue, lblue rl.Color) {
	// cycling
	l.mu.RLock()
	data := l.data
	idx := l.currentIdx
	lastSwitch := l.lastSwitch
	l.mu.RUnlock()

	if len(data) > 1 && time.Since(lastSwitch) >= profCycleDuration {
		l.cycle()
		l.mu.RLock()
		data = l.data
		idx = l.currentIdx
		l.mu.RUnlock()
	}

	// texture swap when item changes
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

	// fade-in
	needsRender := l.dirty.Swap(false)
	if l.Alpha < 255 {
		needsRender = true
		l.Alpha += 510 * rl.GetFrameTime()
		if l.Alpha > 255 {
			l.Alpha = 255
		}
	}

	pad := float32(math.Round(20 * zoom))
	gap := float32(math.Round(15 * zoom))
	//gap5 := float32(math.Round(6 * zoom))

	if needsRender {
		// Pre-render image into imgMask BEFORE the card's BeginTextureMode is active.
		l.imgMask.PreRender(rl.White, func() {
			if l.hasTexture {
				srcW := float32(l.currentTex.Width)
				srcH := float32(l.currentTex.Height)
				scale := l.imgW / srcW
				if l.imgH/srcH > scale {
					scale = l.imgH / srcH
				}
				cropW := l.imgW / scale
				cropH := l.imgH / scale
				src := rl.NewRectangle((srcW-cropW)/2, (srcH-cropH)/2, cropW, cropH)
				dst := rl.NewRectangle(0, 0, l.imgW, l.imgH)
				rl.DrawTexturePro(l.currentTex, src, dst, rl.Vector2{}, 0, rl.White)
			} else {
				rl.DrawRectangle(0, 0, int32(l.imgW), int32(l.imgH), rl.NewColor(220, 228, 242, 255))
			}
		})

		l.Mask.PreRender(rl.White, func() {
			rl.DrawRectangle(0, 0, int32(l.BgRect.Width), int32(l.BgRect.Height), rl.White)

			if len(data) == 0 {
				return
			}
			item := data[idx]

			// ── image: blit pre-rendered imgMask at local (pad, pad) ──────────────
			l.imgMask.Blit(pad, pad)
			// subtle blue border over the image
			borderRoundness := float32(8*zoom) / (l.imgW / 2)
			rl.DrawRectangleRoundedLines(
				rl.NewRectangle(pad, pad, l.imgW, l.imgH),
				borderRoundness, 8,
				rl.NewColor(0, 86, 179, 38),
			)

			// ── info column ───────────────────────────────────────────────────────
			infoX := pad + l.imgW + gap
			infoW := l.BgRect.Width - infoX - pad

			nameFontSize := float32(math.Round(30 * zoom))
			roleFontSize := float32(math.Round(28 * zoom))
			bioFontSize := float32(math.Round(24 * zoom))

			nameLineH := rl.MeasureTextEx(*font, "|", nameFontSize, 0).Y + 2
			roleLineH := rl.MeasureTextEx(*font, "|", roleFontSize, 0).Y + 2

			nameAreaH := nameLineH
			roleAreaH := roleLineH
			bioAreaH := l.BgRect.Height - 2*pad - nameAreaH - roleAreaH

			nameStyle := raycss.TextStyle{RTL: false, HAlign: raycss.AlignStart, VAlign: raycss.AlignStart, Spacing: 0, IgnoreSpace: true, IgnoreNewLine: true}
			for _, line := range raycss.FitText(item.Name, font, infoX, pad, infoW, nameAreaH, nameFontSize, 0, nameStyle) {
				rl.DrawTextEx(*font, line.Text, rl.Vector2{X: line.Rect.X, Y: line.Rect.Y}, line.FontSize, 0, blue)
			}

			roleStyle := raycss.TextStyle{RTL: false, HAlign: raycss.AlignStart, VAlign: raycss.AlignStart, Spacing: 0, IgnoreSpace: true, IgnoreNewLine: true}
			for _, line := range raycss.FitText(strings.ToUpper(item.Role), font, infoX, pad+nameAreaH-float32(math.Round(8.0*zoom)), infoW, roleAreaH, roleFontSize, 0, roleStyle) {
				rl.DrawTextEx(*font, line.Text, rl.Vector2{X: line.Rect.X, Y: line.Rect.Y}, line.FontSize, 0, rl.NewColor(59, 76, 107, 255))
			}

			bioStyle := raycss.TextStyle{RTL: false, HAlign: raycss.AlignJustify, VAlign: raycss.AlignStart, Spacing: 0, IgnoreSpace: false, IgnoreNewLine: false}
			for _, line := range raycss.FitText(item.Bio, font, infoX, pad+nameAreaH+roleAreaH-float32(math.Round(16.0*zoom)), infoW, bioAreaH, bioFontSize, 0, bioStyle) {
				if len(line.JustWords) > 0 {
					for _, jw := range line.JustWords {
						rl.DrawTextEx(*font, jw.Text, rl.Vector2{X: jw.X, Y: line.Rect.Y}, line.FontSize, 0, raycss.TextDark)
					}
				} else {
					rl.DrawTextEx(*font, line.Text, rl.Vector2{X: line.Rect.X, Y: line.Rect.Y}, line.FontSize, 0, raycss.TextDark)
				}
			}

			// fade-in overlay
			if l.Alpha < 255 {
				overlayAlpha := uint8(255 - l.Alpha)
				rl.DrawRectangle(0, 0, int32(l.BgRect.Width), int32(l.BgRect.Height), rl.NewColor(255, 255, 255, overlayAlpha))
			}
		})
	}

	raycss.DrawSoftShadow(l.BgRect, shadowRound*float32(zoom), shadowSegs)
	l.Mask.Blit(l.BgRect.X, l.BgRect.Y)
}
