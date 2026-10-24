package components

import "testing"

func TestParse12HourTime(t *testing.T) {
	good := map[string][2]int{ // input -> hour, minute
		"08:00 AM":   {8, 0},
		"8:00 AM":    {8, 0},
		"8:00 am":    {8, 0},
		" 12:30 PM ": {12, 30},
		"01:05 PM":   {13, 5},
		"12:00 AM":   {0, 0},
	}
	for in, want := range good {
		got, err := Parse12HourTime(in)
		if err != nil {
			t.Errorf("%q: unexpected error %v", in, err)
			continue
		}
		if h, m, _ := got.Clock(); h != want[0] || m != want[1] {
			t.Errorf("%q: got %02d:%02d, want %02d:%02d", in, h, m, want[0], want[1])
		}
	}
	for _, in := range []string{"", "25:00 PM", "noon", "13:00"} {
		if _, err := Parse12HourTime(in); err == nil {
			t.Errorf("%q: expected an error", in)
		}
	}
}
