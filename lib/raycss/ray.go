package raycss

import (
	"math"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
	"mt-y0/ECE-screen/lib/emoji"
	lib "mt-y0/ECE-screen/lib/fonts"
)

// ── palette ───────────────────────────────────────────────────────────────────

var (
	BorderColor = rl.NewColor(220, 220, 220, 255)
	TextDark    = rl.NewColor(30, 30, 30, 255)
	TextMuted   = rl.NewColor(120, 120, 120, 255)
)

// ── Spacing (margin / padding box model) ──────────────────────────────────────

// Spacing represents per-side spacing in logical pixels at zoom = 1.
// Construct with S, SXY, or STRBL; call Scale(zoom) to get display-ready pixel values.
type Spacing struct {
	Top, Right, Bottom, Left float32
}

// S returns uniform spacing on all four sides.
func S(v float32) Spacing { return Spacing{v, v, v, v} }

// SXY returns spacing with equal horizontal (x = left/right) and vertical (y = top/bottom) values.
func SXY(x, y float32) Spacing { return Spacing{y, x, y, x} }

// STRBL returns spacing with explicit top, right, bottom, left values.
func STRBL(top, right, bottom, left float32) Spacing {
	return Spacing{top, right, bottom, left}
}

// Scale returns Spacing with all values multiplied by zoom and rounded to the nearest pixel.
func (s Spacing) Scale(zoom float64) Spacing {
	z := float32(zoom)
	return Spacing{
		Top:    float32(math.Round(float64(s.Top * z))),
		Right:  float32(math.Round(float64(s.Right * z))),
		Bottom: float32(math.Round(float64(s.Bottom * z))),
		Left:   float32(math.Round(float64(s.Left * z))),
	}
}

// H returns total horizontal spacing (Left + Right).
func (s Spacing) H() float32 { return s.Left + s.Right }

// V returns total vertical spacing (Top + Bottom).
func (s Spacing) V() float32 { return s.Top + s.Bottom }

// Inner returns r shrunk by s on all sides — use for padding.
func (s Spacing) Inner(r rl.Rectangle) rl.Rectangle {
	return rl.Rectangle{
		X:      r.X + s.Left,
		Y:      r.Y + s.Top,
		Width:  r.Width - s.H(),
		Height: r.Height - s.V(),
	}
}

// Outer returns r expanded by s on all sides — use for margin.
func (s Spacing) Outer(r rl.Rectangle) rl.Rectangle {
	return rl.Rectangle{
		X:      r.X - s.Left,
		Y:      r.Y - s.Top,
		Width:  r.Width + s.H(),
		Height: r.Height + s.V(),
	}
}

// ── line helpers ──────────────────────────────────────────────────────────────

func DrawHLine(x, y, w float32, c rl.Color) {
	rl.DrawRectangle(int32(x), int32(y), int32(w), 1, c)
}

func DrawVLine(x, y, h float32, c rl.Color) {
	rl.DrawRectangle(int32(x), int32(y), 1, int32(h), c)
}

// ── FitText ───────────────────────────────────────────────────────────────────

// Align controls horizontal or vertical alignment within the padded area.
type Align uint8

const (
	// AlignStart — horizontal: flush to the reading-start edge (left for LTR, right for RTL).
	//              vertical:   flush to the top.
	AlignStart Align = iota
	// AlignCenter — centred in the available space.
	AlignCenter
	// AlignEnd — horizontal: flush to the reading-end edge (right for LTR, left for RTL).
	//            vertical:   flush to the bottom.
	AlignEnd
	// AlignJustify — spread words to fill the full line width.
	// The last line of each paragraph is left-aligned (AlignStart).
	AlignJustify
)

// TextStyle collects all text layout settings for FitText.
// Spacing must match the value passed to rl.DrawTextEx — mismatching causes
// incorrect line widths and misaligned right edges.
type TextStyle struct {
	RTL           bool    // true = Arabic / right-to-left; AlignStart means right edge
	HAlign        Align   // horizontal alignment within the padded area
	VAlign        Align   // vertical alignment within the padded area
	Spacing       float32 // character spacing (0 for Arabic, 1 for Latin)
	UseEmoji      bool    // split lines into Segs and use emoji widths for measurement
	IgnoreSpace   bool    // true = strip all whitespace (strings.Fields); false = preserve leading spaces as indent
	IgnoreNewLine bool    // true = treat \n as space; false = treat \n as a line break
}

// JustWord is a single word with its pre-computed X draw position.
// Only populated on lines produced with AlignJustify (non-last paragraph lines).
type JustWord struct {
	Text string
	X    float32
}

// Line is one display-ready row produced by FitText.
//
// Plain text (no emoji): draw with
//
//	rl.DrawTextEx(*font, line.Text, rl.Vector2{X: line.Rect.X, Y: line.Rect.Y}, line.FontSize, style.Spacing, color)
//
// With emoji (style.UseEmoji == true): draw with
//
//	emoji.DrawRun(line.Segs, atlas, font, line.Rect.X, line.Rect.Y, line.FontSize, color)
//
// For AlignJustify non-last lines: JustWords is non-nil — draw each word at its X.
type Line struct {
	Rect      rl.Rectangle // absolute draw position; Width = measured text width
	FontSize  float32
	Text      string          // shaped — always set; use with rl.DrawTextEx
	Segs      []emoji.Segment // only set when style.UseEmoji; use with emoji.DrawRun
	JustWords []JustWord      // non-nil only for AlignJustify non-last paragraph lines
}

// FitText reshapes raw Arabic (or mixed) text and lays it out inside the area
// (x, y, width, height) applying padding on all sides.
//
// Explicit "\n" in text forces a line break. Fit strategy:
//  1. Word-wrap at fontSize; return if all lines fit vertically.
//  2. Shrink font 1 px at a time down to 70 % of fontSize and retry.
//  3. If still overflowing at 70 %, keep as many lines as fit and append "..."
//     to the last visible line.
func FitText(text string, font *rl.Font, x, y, width, height, fontSize, padding float32, style TextStyle) []Line {
	innerX := x + padding
	innerY := y + padding
	innerW := width - 2*padding
	innerH := height - 2*padding
	if innerW <= 0 || innerH <= 0 {
		return nil
	}
	minFS := fontSize * 0.70

	for fs := fontSize; fs >= minFS; fs-- {
		if lines, ok := fitLayout(text, font, innerX, innerY, innerW, innerH, fs, style, false); ok {
			return lines
		}
	}
	lines, _ := fitLayout(text, font, innerX, innerY, innerW, innerH, minFS, style, true)
	return lines
}

// fitLayout performs the word-wrap and returns the Line slice.
// When clip is true it limits output to the lines that fit and truncates the
// last one with "..." instead of returning (nil, false).
func fitLayout(text string, font *rl.Font, x, y, w, h float32, fs float32, style TextStyle, clip bool) ([]Line, bool) {
	sp := style.Spacing
	lineH := rl.MeasureTextEx(*font, "|", fs, sp).Y + 2
	spaceW := rl.MeasureTextEx(*font, " ", fs, sp).X
	rightEdge := x + w
	emojiSz := fs * 0.85 // mirrors emoji.DrawRun's internal formula
	maxLines := int(h / lineH)
	if maxLines < 1 {
		maxLines = 1
	}

	// measureShaped returns the pixel width of an already-shaped string,
	// using emoji-aware measurement when UseEmoji is set.
	measureShaped := func(shaped string) float32 {
		if style.UseEmoji {
			return emoji.MeasureRun(emoji.Split(shaped), font, fs, emojiSz)
		}
		return rl.MeasureTextEx(*font, shaped, fs, sp).X
	}

	// ── step 1: group original (unshaped) words into line buckets ────────────
	// Words are kept unshaped so we Reshape the whole joined line at once,
	// preserving correct cross-word contextual Arabic forms.
	//
	// leadIndent is the extra X offset applied to the FIRST line of each paragraph
	// when IgnoreSpace is false — it represents the width of the leading whitespace
	// characters in the input (capped at 4 spaces to avoid extreme indentation).
	type bucket struct {
		words      []string
		leadIndent float32 // non-zero only for first line of a paragraph with leading spaces
		isParaEnd  bool    // true for the last bucket of each paragraph (not justified)
	}
	var buckets []bucket

	// IgnoreNewLine: replace \n with space so the split below produces one paragraph.
	inputText := text
	if style.IgnoreNewLine {
		inputText = strings.ReplaceAll(inputText, "\n", " ")
	}

	for _, para := range strings.Split(inputText, "\n") {
		paraFirstBucket := len(buckets)
		var leadIndent float32

		// IgnoreSpace=false: measure leading whitespace, strip it, use it as indent.
		// IgnoreSpace=true: strings.Fields strips everything — no extra work needed.
		effectivePara := para
		if !style.IgnoreSpace {
			trimmed := strings.TrimLeft(para, " \t")
			nSpaces := len([]rune(para)) - len([]rune(trimmed))
			if nSpaces > 4 {
				nSpaces = 4 // cap so extreme padding doesn't push text off-screen
			}
			leadIndent = float32(nSpaces) * spaceW
			effectivePara = trimmed
		}

		// Build word list and per-word gap widths.
		// IgnoreSpace=true  → gaps are all exactly spaceW (standard single space).
		// IgnoreSpace=false → gaps reflect the actual number of spaces in the input
		//                     so "word1   word2" gets 3× spaceW between them.
		type wordEntry struct {
			text string
			gap  float32 // pixel gap BEFORE this word (0 for the first word on a line)
		}
		var entries []wordEntry
		if style.IgnoreSpace {
			for _, wd := range strings.Fields(effectivePara) {
				entries = append(entries, wordEntry{wd, spaceW})
			}
		} else {
			pending := 0
			for _, part := range strings.Split(effectivePara, " ") {
				if part == "" {
					pending++
				} else {
					g := float32(pending+1) * spaceW // +1 for the split separator itself
					if len(entries) == 0 {
						g = 0 // no pre-gap before the very first word
					}
					entries = append(entries, wordEntry{part, g})
					pending = 0
				}
			}
		}
		if len(entries) == 0 {
			buckets = append(buckets, bucket{nil, 0, true})
			continue
		}

		widths := make([]float32, len(entries))
		for i, e := range entries {
			widths[i] = measureShaped(lib.Reshape(e.text))
		}

		words := make([]string, len(entries))
		for i, e := range entries {
			words[i] = e.text
		}

		start, cur := 0, widths[0]
		for i := 1; i < len(entries); i++ {
			gap := entries[i].gap
			if cur+gap+widths[i] <= w {
				cur += gap + widths[i]
			} else {
				buckets = append(buckets, bucket{words[start:i], 0, false})
				start, cur = i, widths[i]
			}
		}
		buckets = append(buckets, bucket{words[start:], 0, true}) // last of this paragraph

		// Attach the leading indent to the first bucket of this paragraph.
		if leadIndent > 0 && paraFirstBucket < len(buckets) {
			buckets[paraFirstBucket].leadIndent = leadIndent
		}
	}

	fits := len(buckets) <= maxLines
	if !fits && !clip {
		return nil, false
	}

	// ── step 2: build Line objects (top-aligned first pass) ──────────────────
	var result []Line
	for i, b := range buckets {
		if i >= maxLines {
			break
		}
		lineY := y + float32(i)*lineH

		if len(b.words) == 0 {
			result = append(result, Line{
				Rect:     rl.Rectangle{X: x, Y: lineY, Width: 0, Height: lineH},
				FontSize: fs,
				Text:     "",
			})
			continue
		}

		isLastVisible := clip && i == maxLines-1 && i < len(buckets)-1
		var shaped string
		if isLastVisible {
			shaped = fitTruncateWords(b.words, w, spaceW, measureShaped)
		} else {
			shaped = lib.Reshape(strings.Join(b.words, " "))
		}

		lineW := measureShaped(shaped)

		// Justify: distribute words to fill the full line width.
		// Last paragraph line and single-word lines fall back to AlignStart.
		var justWords []JustWord
		hAlign := style.HAlign
		if style.HAlign == AlignJustify {
			if !b.isParaEnd && !isLastVisible && len(b.words) > 1 {
				justWords = make([]JustWord, len(b.words))
				totalWordW := float32(0)
				shapedWds := make([]string, len(b.words))
				wwidths := make([]float32, len(b.words))
				for wi, word := range b.words {
					sw := lib.Reshape(word)
					shapedWds[wi] = sw
					wwidths[wi] = measureShaped(sw)
					totalWordW += wwidths[wi]
				}
				justGap := (w - b.leadIndent - totalWordW) / float32(len(b.words)-1)
				if justGap < spaceW {
					justGap = spaceW
				}
				cx := x + b.leadIndent
				for wi, sw := range shapedWds {
					justWords[wi] = JustWord{Text: sw, X: cx}
					cx += wwidths[wi] + justGap
				}
			}
			hAlign = AlignStart // last line of para or single word: flush start
		}
		lineX := fitLineX(x, w, lineW, rightEdge, b.leadIndent, style.RTL, hAlign)

		ln := Line{
			Rect:      rl.Rectangle{X: lineX, Y: lineY, Width: lineW, Height: lineH},
			FontSize:  fs,
			Text:      shaped,
			JustWords: justWords,
		}
		if style.UseEmoji {
			ln.Segs = emoji.Split(shaped)
		}
		result = append(result, ln)
	}

	// ── step 3: apply vertical alignment offset ───────────────────────────────
	if len(result) > 0 && style.VAlign != AlignStart {
		totalH := float32(len(result)) * lineH
		var vOff float32
		switch style.VAlign {
		case AlignCenter:
			vOff = (h - totalH) / 2
		case AlignEnd:
			vOff = h - totalH
		}
		if vOff < 0 {
			vOff = 0
		}
		for i := range result {
			result[i].Rect.Y += vOff
		}
	}

	return result, fits
}

// fitLineX computes the draw X for a line given alignment settings.
// leadIndent shrinks the available area from the reading-start edge:
//   - RTL: indent comes from the right (rightEdge moves left by leadIndent)
//   - LTR: indent comes from the left  (innerX moves right by leadIndent)
func fitLineX(innerX, innerW, lineW, rightEdge, leadIndent float32, rtl bool, ha Align) float32 {
	// Adjust boundaries based on direction
	leftBound := innerX
	rightBound := rightEdge
	if rtl {
		rightBound -= leadIndent // indent eats into the right (reading-start) side
	} else {
		leftBound += leadIndent // indent eats into the left (reading-start) side
	}
	availW := rightBound - leftBound

	var lx float32
	switch ha {
	case AlignCenter:
		lx = leftBound + (availW-lineW)/2
	case AlignEnd:
		if rtl {
			lx = leftBound // RTL reading-end is the left edge
		} else {
			lx = rightBound - lineW // LTR reading-end is the right edge
		}
	default: // AlignStart
		if rtl {
			lx = rightBound - lineW // RTL reading-start is the right edge
		} else {
			lx = leftBound // LTR reading-start is the left edge
		}
	}
	// Clamp: never draw outside the inner area
	if lx < innerX {
		lx = innerX
	}
	if lx+lineW > rightEdge {
		lx = rightEdge - lineW
		if lx < innerX {
			lx = innerX
		}
	}
	return lx
}

// fitTruncateWords packs as many words as fit in w (leaving room for "..."),
// reshapes the result, then appends "...".
// measure is a function that returns the pixel width of an already-shaped string.
func fitTruncateWords(words []string, w, spaceW float32, measure func(string) float32) string {
	const ellipsis = "..."
	ellW := measure(ellipsis)

	var kept []string
	used := float32(0)
	for _, word := range words {
		wW := measure(lib.Reshape(word))
		gap := spaceW
		if len(kept) == 0 {
			gap = 0
		}
		if used+gap+wW+ellW <= w {
			kept = append(kept, word)
			used += gap + wW
		} else {
			break
		}
	}
	if len(kept) == 0 {
		return ellipsis
	}
	return lib.Reshape(strings.Join(kept, " ")) + ellipsis
}

// DrawSoftShadow simulates box-shadow: 0 4px 6px rgba(0,0,0,0.05).
// radius is the pixel corner radius of the card (same value as its
// MaskContainer); each layer grows it by its spread so corners stay concentric.
func DrawSoftShadow(rect rl.Rectangle, radius float32, segments int32) {
	const offsetY = float32(4)
	spreads := [5]float32{5, 4, 3, 2, 1}
	alphas := [5]uint8{1, 2, 3, 4, 5}
	for i := range spreads {
		spread := spreads[i]
		s := rl.NewRectangle(
			rect.X-spread/2,
			rect.Y+offsetY-spread/2,
			rect.Width+spread,
			rect.Height+spread,
		)
		// raylib roundness is relative to min(w,h): radius = roundness*min/2
		r := radius + spread/2
		roundness := 2 * r / min(s.Width, s.Height)
		rl.DrawRectangleRounded(s, roundness, segments, rl.NewColor(0, 0, 0, alphas[i]))
	}
}

// fragShaderSrc clips pixels outside a rounded rectangle using a signed distance field.
// uniforms:
//
//	uSize   vec2  — pixel dimensions of the render texture (width, height)
//	uRadius float — corner radius in pixels
const fragShaderSrc = `
#version 330
in vec2 fragTexCoord;
in vec4 fragColor;
out vec4 finalColor;

uniform sampler2D texture0;
uniform vec2  uSize;
uniform float uRadius;

// Returns < 0 inside the rounded rect, > 0 outside.
float roundedRectSDF(vec2 p, vec2 halfSize, float r) {
    vec2 q = abs(p) - halfSize + r;
    return length(max(q, 0.0)) + min(max(q.x, q.y), 0.0) - r;
}

void main() {
    vec4 tex = texture(texture0, fragTexCoord);

    // Map UV [0,1] → pixel space centred at (0,0)
    vec2 pixelPos = (fragTexCoord - vec2(0.5)) * uSize;

    float dist = roundedRectSDF(pixelPos, uSize * 0.5, uRadius);

float alpha = 1.0 - smoothstep(-1.0, 1.0, dist);
finalColor = vec4(tex.rgb, tex.a * fragColor.a * alpha);
}
`

// MaskContainer renders its children into an off-screen texture, then blits that
// texture to the screen with a ounded-rectangle clip applied via an SDF shader.
// Children draw with (0,0) as their local origin — they never need to know about
// the rounding.
type MaskContainer struct {
	target   rl.RenderTexture2D
	shader   rl.Shader
	uSizeLoc int32
	uRadLoc  int32
	width    int32
	height   int32
	radius   float32
}

// NewMaskContainer allocates the render texture and compiles the clip shader.
// Call Unload() when the container is no longer needed.
func NewMaskContainer(width, height int32, radius float32) MaskContainer {
	shader := rl.LoadShaderFromMemory("", fragShaderSrc)

	mc := MaskContainer{
		target:   rl.LoadRenderTexture(width, height),
		shader:   shader,
		uSizeLoc: rl.GetShaderLocation(shader, "uSize"),
		uRadLoc:  rl.GetShaderLocation(shader, "uRadius"),
		width:    width,
		height:   height,
		radius:   radius,
	}

	// Upload uniforms once — they never change for a fixed-size container.
	rl.SetShaderValue(shader, mc.uSizeLoc,
		[]float32{float32(width), float32(height)},
		rl.ShaderUniformVec2)
	rl.SetShaderValue(shader, mc.uRadLoc,
		[]float32{radius},
		rl.ShaderUniformFloat)

	return mc
}

// Unload frees the GPU resources. Must be called from the same goroutine/thread
// that called NewMaskContainer.
func (mc *MaskContainer) Unload() {
	rl.UnloadRenderTexture(mc.target)
	rl.UnloadShader(mc.shader)
}

// PreRender renders drawChildren into the off-screen buffer without blitting to screen.
// Call Blit afterwards (from within another render target or the main framebuffer) to
// composite the result. This allows nested use: pre-render inner masks before the outer
// mask's BeginTextureMode is active, then Blit inside the outer draw callback.
func (mc *MaskContainer) PreRender(backgroundColor rl.Color, drawChildren func()) {
	rl.BeginTextureMode(mc.target)
	rl.ClearBackground(backgroundColor)
	drawChildren()
	rl.EndTextureMode()
}

// Blit composites the pre-rendered texture onto the current render target at (x, y)
// with the rounded-corner clip shader applied.
func (mc *MaskContainer) Blit(x, y float32) {
	src := rl.NewRectangle(0, 0, float32(mc.width), -float32(mc.height))
	dst := rl.NewRectangle(x, y, float32(mc.width), float32(mc.height))
	rl.BeginShaderMode(mc.shader)
	rl.DrawTexturePro(mc.target.Texture, src, dst, rl.Vector2Zero(), 0, rl.White)
	rl.EndShaderMode()
}

// Draw renders drawChildren into the off-screen buffer, then composites the result
// onto the current render target at (x, y) with rounded corners.
// All drawing inside drawChildren should use local coordinates — (0, 0) is the
// top-left corner of the card.
func (mc *MaskContainer) Draw(x, y float32, backgroundColor rl.Color, drawChildren func()) {
	mc.PreRender(backgroundColor, drawChildren)
	mc.Blit(x, y)
}
