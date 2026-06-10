package lib

import (
	"unicode"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type GlyphGroup struct {
	Isolated, Initial, Medial, Final rune
	ConnectsToNext                   bool
}

var Registry = map[rune]GlyphGroup{
	// ── Dual-joining ─────────────────────────────────────────────────────────
	0x0628: {0xFE8F, 0xFE91, 0xFE92, 0xFE90, true}, // Ba    ب
	0x062A: {0xFE95, 0xFE97, 0xFE98, 0xFE96, true}, // Ta    ت
	0x062B: {0xFE99, 0xFE9B, 0xFE9C, 0xFE9A, true}, // Tha   ث
	0x062C: {0xFE9D, 0xFE9F, 0xFEA0, 0xFE9E, true}, // Jeem  ج
	0x062D: {0xFEA1, 0xFEA3, 0xFEA4, 0xFEA2, true}, // Ha    ح
	0x062E: {0xFEA5, 0xFEA7, 0xFEA8, 0xFEA6, true}, // Kha   خ
	0x0633: {0xFEB1, 0xFEB3, 0xFEB4, 0xFEB2, true}, // Seen  س
	0x0634: {0xFEB5, 0xFEB7, 0xFEB8, 0xFEB6, true}, // Sheen ش
	0x0635: {0xFEB9, 0xFEBB, 0xFEBC, 0xFEBA, true}, // Sad   ص
	0x0636: {0xFEBD, 0xFEBF, 0xFEC0, 0xFEBE, true}, // Dad   ض
	0x0637: {0xFEC1, 0xFEC3, 0xFEC4, 0xFEC2, true}, // Ta'   ط
	0x0638: {0xFEC5, 0xFEC7, 0xFEC8, 0xFEC6, true}, // Za'   ظ
	0x0639: {0xFEC9, 0xFECB, 0xFECC, 0xFECA, true}, // Ain   ع
	0x063A: {0xFECD, 0xFECF, 0xFED0, 0xFECE, true}, // Ghain غ
	0x0641: {0xFED1, 0xFED3, 0xFED4, 0xFED2, true}, // Fa    ف
	0x0642: {0xFED5, 0xFED7, 0xFED8, 0xFED6, true}, // Qaf   ق
	0x0643: {0xFED9, 0xFEDB, 0xFEDC, 0xFEDA, true}, // Kaf   ك
	0x0644: {0xFEDD, 0xFEDF, 0xFEE0, 0xFEDE, true}, // Lam   ل
	0x0645: {0xFEE1, 0xFEE3, 0xFEE4, 0xFEE2, true}, // Meem  م
	0x0646: {0xFEE5, 0xFEE7, 0xFEE8, 0xFEE6, true}, // Noon  ن
	0x0647: {0xFEE9, 0xFEEB, 0xFEEC, 0xFEEA, true}, // Ha    ه
	0x064A: {0xFEF1, 0xFEF3, 0xFEF4, 0xFEF2, true}, // Ya    ي
	0x0649: {0xFEEF, 0xFEF0, 0xFEF0, 0xFEEE, true}, // Alef Maqsura ى
	0x0626: {0xFE89, 0xFE8B, 0xFE8C, 0xFE8A, true}, // Ya+Hamza    ئ

	// ── Right-joining only ───────────────────────────────────────────────────
	0x0621: {0xFE80, 0xFE80, 0xFE80, 0xFE80, false}, // Hamza       ء
	0x0622: {0xFE81, 0xFE81, 0xFE81, 0xFE82, false}, // Alef Madda  آ
	0x0623: {0xFE83, 0xFE83, 0xFE83, 0xFE84, false}, // Alef+Hamza  أ
	0x0624: {0xFE85, 0xFE85, 0xFE85, 0xFE86, false}, // Waw+Hamza   ؤ
	0x0625: {0xFE87, 0xFE87, 0xFE87, 0xFE88, false}, // Alef below  إ
	0x0627: {0xFE8D, 0xFE8D, 0xFE8D, 0xFE8E, false}, // Alef        ا
	0x0629: {0xFE93, 0xFE93, 0xFE93, 0xFE94, false}, // Ta Marbuta  ة
	0x062F: {0xFEA9, 0xFEA9, 0xFEA9, 0xFEAA, false}, // Dal         د
	0x0630: {0xFEAB, 0xFEAB, 0xFEAB, 0xFEAC, false}, // Dhal        ذ
	0x0631: {0xFEAD, 0xFEAD, 0xFEAD, 0xFEAE, false}, // Ra          ر
	0x0632: {0xFEAF, 0xFEAF, 0xFEAF, 0xFEB0, false}, // Zay         ز
	0x0648: {0xFEED, 0xFEED, 0xFEED, 0xFEEE, false}, // Waw         و
	0x0671: {0xFB50, 0xFB50, 0xFB50, 0xFB51, false}, // Alef Wasla  ٱ
}

func isTashkeel(r rune) bool {
	return (r >= 0x064B && r <= 0x065F) || r == 0x0670
}

func isArabicLetter(r rune) bool {
	_, ok := Registry[r]
	return ok
}

func isArabicRune(r rune) bool {
	return unicode.Is(unicode.Arabic, r) || isArabicLetter(r)
}

// Reshape converts Arabic (and mixed) text into presentation forms for correct
// display in raylib's LTR renderer. Arabic runs are shaped and reversed; LTR
// runs (Latin, digits, punctuation) are preserved as-is.
func Reshape(input string) string {
	runes := []rune(input)
	n := len(runes)
	if n == 0 {
		return ""
	}

	// ── Step 1: contextual shaping ────────────────────────────────────────────
	shaped := make([]rune, n)
	for i := range n {
		cur := runes[i]

		if isTashkeel(cur) {
			shaped[i] = cur
			continue
		}

		grp, known := Registry[cur]
		if !known {
			shaped[i] = cur
			continue
		}

		// Look BEHIND (skip tashkeel)
		hasLeft := false
		prev := i - 1
		for prev >= 0 && isTashkeel(runes[prev]) {
			prev--
		}
		if prev >= 0 && runes[prev] != ' ' {
			if pg, ok := Registry[runes[prev]]; ok && pg.ConnectsToNext {
				hasLeft = true
			}
		}

		// Look AHEAD (skip tashkeel) — right-joining-only letters cannot connect leftward.
		hasRight := false
		if grp.ConnectsToNext {
			next := i + 1
			for next < n && isTashkeel(runes[next]) {
				next++
			}
			if next < n && runes[next] != ' ' {
				if _, ok := Registry[runes[next]]; ok {
					hasRight = true
				}
			}
		}

		switch {
		case hasLeft && hasRight:
			shaped[i] = grp.Medial
		case hasLeft:
			shaped[i] = grp.Final
		case hasRight:
			shaped[i] = grp.Initial
		default:
			shaped[i] = grp.Isolated
		}
	}

	// ── Step 2: Lam–Alef ligatures ────────────────────────────────────────────
	alefSet := map[rune]bool{
		0x0627: true, 0x0622: true, 0x0623: true, 0x0625: true, 0x0671: true,
	}
	lamRune := rune(0x0644)

	ligatured := make([]rune, 0, n)
	for i := 0; i < n; i++ {
		j := i + 1
		for j < n && isTashkeel(runes[j]) {
			j++
		}

		if runes[i] == lamRune && j < n && alefSet[runes[j]] {
			lamHasLeft := false
			prev := i - 1
			for prev >= 0 && isTashkeel(runes[prev]) {
				prev--
			}
			if prev >= 0 && runes[prev] != ' ' {
				if pg, ok := Registry[runes[prev]]; ok && pg.ConnectsToNext {
					lamHasLeft = true
				}
			}

			var lig rune
			switch runes[j] {
			case 0x0622:
				if lamHasLeft {
					lig = 0xFEF6
				} else {
					lig = 0xFEF5
				}
			case 0x0623:
				if lamHasLeft {
					lig = 0xFEF8
				} else {
					lig = 0xFEF7
				}
			case 0x0625:
				if lamHasLeft {
					lig = 0xFEFA
				} else {
					lig = 0xFEF9
				}
			default:
				if lamHasLeft {
					lig = 0xFEFC
				} else {
					lig = 0xFEFB
				}
			}
			ligatured = append(ligatured, lig)
			for k := i + 1; k < j; k++ {
				ligatured = append(ligatured, shaped[k])
			}
			i = j
		} else {
			ligatured = append(ligatured, shaped[i])
		}
	}

	// ── Step 3: bidi-aware reversal ───────────────────────────────────────────
	type run struct {
		runes []rune
		isLTR bool
	}

	var runs []run
	if len(ligatured) == 0 {
		return ""
	}

	cur := ligatured[0]
	curIsLTR := !isArabicRune(cur) && !isTashkeel(cur)
	runStart := []rune{cur}

	flush := func(r []rune, ltr bool) {
		if len(r) > 0 {
			runs = append(runs, run{r, ltr})
		}
	}

	for _, r := range ligatured[1:] {
		rIsLTR := !isArabicRune(r) && !isTashkeel(r)
		if rIsLTR == curIsLTR {
			runStart = append(runStart, r)
		} else {
			flush(runStart, curIsLTR)
			runStart = []rune{r}
			curIsLTR = rIsLTR
		}
	}
	flush(runStart, curIsLTR)

	for i := range runs {
		if !runs[i].isLTR {
			s := runs[i].runes
			for lo, hi := 0, len(s)-1; lo < hi; lo, hi = lo+1, hi-1 {
				s[lo], s[hi] = s[hi], s[lo]
			}
		}
	}

	for lo, hi := 0, len(runs)-1; lo < hi; lo, hi = lo+1, hi-1 {
		runs[lo], runs[hi] = runs[hi], runs[lo]
	}

	out := make([]rune, 0, len(ligatured))
	for _, run := range runs {
		out = append(out, run.runes...)
	}

	return string(out)
}

// GetArabicCodepoints returns codepoints for the Arabic font atlas.
func GetArabicCodepoints() []int32 {
	var list []int32

	for r := int32(32); r <= 126; r++ {
		list = append(list, r)
	}
	for r := int32(0x00A0); r <= 0x00FF; r++ {
		list = append(list, r)
	}
	for r := int32(0x2000); r <= 0x206F; r++ {
		list = append(list, r)
	}
	for r := int32(0x0600); r <= 0x06FF; r++ {
		list = append(list, r)
	}
	for r := int32(0x08A0); r <= 0x08FF; r++ {
		list = append(list, r)
	}
	for r := int32(0xFB50); r <= 0xFDFF; r++ {
		list = append(list, r)
	}
	for r := int32(0xFE70); r <= 0xFEFF; r++ {
		list = append(list, r)
	}

	return list
}

func LoadFont(fontdata []byte) *rl.Font {
	cps := GetArabicCodepoints()
	font := rl.LoadFontFromMemory(".ttf", fontdata, 64, cps)
	rl.SetTextureFilter(font.Texture, rl.FilterBilinear)
	return &font
}
