package lib

import "testing"

func TestReshapeNonArabicUnchanged(t *testing.T) {
	for _, s := range []string{"", "Hello 123", "A/B-C"} {
		if got := Reshape(s); got != s {
			t.Errorf("Reshape(%q) = %q, want unchanged", s, got)
		}
	}
}

func TestReshapeArabicUsesPresentationForms(t *testing.T) {
	got := Reshape("سلام")
	if got == "سلام" {
		t.Fatalf("Reshape left Arabic text unshaped")
	}
	for _, r := range got {
		if r < 0xFB50 || r > 0xFEFF {
			t.Errorf("rune %U is outside the Arabic presentation-form blocks", r)
		}
	}
}

func TestReshapeDeterministic(t *testing.T) {
	in := "المرحلة الرابعة"
	if Reshape(in) != Reshape(in) {
		t.Error("Reshape is not deterministic")
	}
}
