package emoji

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestSplitPlainText(t *testing.T) {
	segs := Split("hello")
	if len(segs) != 1 || segs[0].IsEmoji || segs[0].Text != "hello" {
		t.Fatalf("got %+v", segs)
	}
}

func TestSplitEmojiInText(t *testing.T) {
	segs := Split("a😀b")
	if len(segs) != 3 || !segs[1].IsEmoji || segs[1].EmojiKey != "1f600" {
		t.Fatalf("got %+v", segs)
	}
}

func TestSplitFoldsVariationSelectorAndFlags(t *testing.T) {
	if segs := Split("☀️"); len(segs) != 1 || !segs[0].IsEmoji {
		t.Errorf("FE0F not folded: %+v", segs)
	}
	if segs := Split("🇸🇦"); len(segs) != 1 || segs[0].EmojiKey != "1f1f8-1f1e6" {
		t.Errorf("flag pair not folded: %+v", segs)
	}
}

func TestInertAtlasIsSafe(t *testing.T) {
	var a *Atlas
	if a.Draw("1f600", 0, 0, 10, rl.White) {
		t.Error("nil atlas drew something")
	}
}
