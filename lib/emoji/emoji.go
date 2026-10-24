package emoji

import (
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const cellSize = 64

// Atlas packs Twemoji 72×72 PNGs from a directory into one GPU sprite sheet.
// All public methods are safe to call on a nil or empty Atlas.
type Atlas struct {
	texture rl.Texture2D
	rects   map[string]rl.Rectangle
	loaded  bool
}

// Load scans dir for *.png files named by Twemoji convention ("1f600.png",
// "1f1f8-1f1e6.png", …), packs them into a single texture atlas, and returns
// the Atlas. If dir is missing or empty an inert Atlas is returned — callers
// never need to nil-check.
// Must be called after rl.InitWindow.
func Load(dir string) *Atlas {
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("emoji: cannot read %q (%v) — emoji will not render", dir, err)
		return &Atlas{rects: make(map[string]rl.Rectangle)}
	}

	var keys []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := strings.ToLower(e.Name())
		if filepath.Ext(name) == ".png" {
			keys = append(keys, strings.TrimSuffix(name, ".png"))
		}
	}
	if len(keys) == 0 {
		return &Atlas{rects: make(map[string]rl.Rectangle)}
	}
	sort.Strings(keys)

	cols := int(math.Ceil(math.Sqrt(float64(len(keys)))))
	rows := (len(keys) + cols - 1) / cols

	atlasImg := rl.GenImageColor(cols*cellSize, rows*cellSize, rl.Blank)
	rects := make(map[string]rl.Rectangle, len(keys))

	src := rl.Rectangle{X: 0, Y: 0, Width: cellSize, Height: cellSize}
	failed := 0
	for idx, key := range keys {
		img := rl.LoadImage(filepath.Join(dir, key+".png"))
		if img.Width == 0 {
			rl.UnloadImage(img)
			failed++
			log.Printf("emoji: failed to load %s.png", key)
			continue
		}
		rl.ImageResize(img, cellSize, cellSize)

		dst := rl.Rectangle{
			X:      float32((idx % cols) * cellSize),
			Y:      float32((idx / cols) * cellSize),
			Width:  cellSize,
			Height: cellSize,
		}
		rl.ImageDraw(atlasImg, img, src, dst, rl.White)
		rl.UnloadImage(img)
		rects[key] = dst
	}

	if failed > 0 {
		log.Printf("emoji: %d of %d images failed to load", failed, len(keys))
	}

	tex := rl.LoadTextureFromImage(atlasImg)
	rl.UnloadImage(atlasImg)
	rl.SetTextureFilter(tex, rl.FilterBilinear)

	return &Atlas{texture: tex, rects: rects, loaded: true}
}

// Draw renders the emoji for key at (x, y) scaled to size×size pixels.
// Returns false when the emoji is not in the atlas.
func (a *Atlas) Draw(key string, x, y, size float32, tint rl.Color) bool {
	rect, ok := a.find(key)
	if !ok {
		return false
	}
	rl.DrawTexturePro(a.texture, rect,
		rl.Rectangle{X: x, Y: y, Width: size, Height: size},
		rl.Vector2{}, 0, tint)
	return true
}

// Unload frees the GPU texture. Call before rl.CloseWindow.
func (a *Atlas) Unload() {
	if a != nil && a.loaded {
		rl.UnloadTexture(a.texture)
		a.loaded = false
	}
}

// find resolves a key, falling back to a FE0F-stripped variant.
func (a *Atlas) find(key string) (rl.Rectangle, bool) {
	if a == nil {
		return rl.Rectangle{}, false
	}
	if r, ok := a.rects[key]; ok {
		return r, true
	}
	if s := dropFE0F(key); s != key {
		if r, ok := a.rects[s]; ok {
			return r, true
		}
	}
	return rl.Rectangle{}, false
}

func dropFE0F(key string) string {
	parts := strings.Split(key, "-")
	out := parts[:0]
	for _, p := range parts {
		if p != "fe0f" {
			out = append(out, p)
		}
	}
	return strings.Join(out, "-")
}

// ────────────────────────────────────────────────────────────────────────────

// Segment is either a plain-text run or a single emoji.
type Segment struct {
	IsEmoji  bool
	Text     string // plain text when IsEmoji == false
	EmojiKey string // Twemoji lookup key, e.g. "1f600", when IsEmoji == true
}

// Split breaks text (already Reshape-processed) into alternating plain-text
// and emoji segments. Variation selectors and ZWJ are folded into the emoji
// key and never appear as separate segments.
func Split(text string) []Segment {
	runes := []rune(text)
	n := len(runes)
	var out []Segment
	i := 0

	for i < n {
		r := runes[i]

		switch {
		case isRegionalIndicator(r):
			if i+1 < n && isRegionalIndicator(runes[i+1]) {
				key := fmt.Sprintf("%x-%x", r, runes[i+1])
				out = append(out, Segment{IsEmoji: true, EmojiKey: key})
				i += 2
			} else {
				out = pushText(out, string(r))
				i++
			}

		case isBaseEmoji(r):
			seq := []rune{r}
			j := i + 1
			// variation selector
			if j < n && runes[j] == 0xFE0F {
				seq = append(seq, runes[j])
				j++
			}
			// skin-tone modifier
			if j < n && runes[j] >= 0x1F3FB && runes[j] <= 0x1F3FF {
				seq = append(seq, runes[j])
				j++
			}
			// ZWJ chain
			for j < n && runes[j] == 0x200D {
				seq = append(seq, runes[j])
				j++
				if j < n && isBaseEmoji(runes[j]) {
					seq = append(seq, runes[j])
					j++
					if j < n && runes[j] == 0xFE0F {
						seq = append(seq, runes[j])
						j++
					}
				}
			}
			out = append(out, Segment{IsEmoji: true, EmojiKey: runesKey(seq)})
			i = j

		case r == 0xFE0F || r == 0x200D:
			i++ // stray combiner — skip

		default:
			j := i + 1
			for j < n && !isBaseEmoji(runes[j]) && !isRegionalIndicator(runes[j]) &&
				runes[j] != 0xFE0F && runes[j] != 0x200D {
				j++
			}
			out = pushText(out, string(runes[i:j]))
			i = j
		}
	}
	return out
}

// DrawRun renders a pre-split segment slice at (x, y), returning the total width.
// emojiSize should typically equal fontSize for a clean baseline; pass a smaller
// value to vertically-centre emoji within the line height.
func DrawRun(segs []Segment, atlas *Atlas, font *rl.Font, x, y, fontSize float32, tint rl.Color) float32 {
	startX := x
	emojiSize := fontSize * 0.65
	for _, seg := range segs {
		if seg.IsEmoji {
			x += emojiSize * 0.25
			if atlas != nil && atlas.Draw(seg.EmojiKey, x, y-emojiSize/2, emojiSize, rl.White) {
				x += emojiSize + 2
			}
		} else {
			sz := rl.MeasureTextEx(*font, seg.Text, fontSize, 1)
			rl.DrawTextEx(*font, seg.Text, rl.Vector2{X: x, Y: y - sz.Y/2}, fontSize, 1, tint)
			x += sz.X
		}
	}
	return x - startX
}

// MeasureRun returns the total pixel width of segs without drawing anything.
func MeasureRun(segs []Segment, font *rl.Font, fontSize, emojiSize float32) float32 {
	var w float32
	for _, seg := range segs {
		w += emojiSize * 0.25
		if seg.IsEmoji {
			w += emojiSize + 2
		} else {
			w += rl.MeasureTextEx(*font, seg.Text, fontSize, 1).X
		}
	}
	return w
}

// pushText appends s to the last text segment, or creates a new one.
func pushText(segs []Segment, s string) []Segment {
	if len(segs) > 0 && !segs[len(segs)-1].IsEmoji {
		segs[len(segs)-1].Text += s
		return segs
	}
	return append(segs, Segment{Text: s})
}

func runesKey(seq []rune) string {
	parts := make([]string, len(seq))
	for i, r := range seq {
		parts[i] = fmt.Sprintf("%x", r)
	}
	return strings.Join(parts, "-")
}

func isBaseEmoji(r rune) bool {
	return (r >= 0x1F300 && r <= 0x1FAFF) ||
		(r >= 0x2600 && r <= 0x27BF)
}

func isRegionalIndicator(r rune) bool {
	return r >= 0x1F1E0 && r <= 0x1F1FF
}
