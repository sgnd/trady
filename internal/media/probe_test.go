package media

import "testing"

func TestSuggestMoments(t *testing.T) {
	cases := []struct {
		dur  float64
		want int
	}{
		{0, 1}, // unknown duration falls back to 60s
		{60, 1},
		{900, 8}, // capped at 8
		{10000, 8},
	}
	for _, c := range cases {
		ms := SuggestMoments("p", c.dur)
		if len(ms) != c.want {
			t.Fatalf("dur %v: got %d moments, want %d", c.dur, len(ms), c.want)
		}
		for _, m := range ms {
			if m.ProjectID != "p" || m.EndSec <= m.StartSec || m.Score < 50 || m.Score > 90 {
				t.Fatalf("dur %v: bad moment %+v", c.dur, m)
			}
			if c.dur > 0 && m.EndSec > c.dur {
				t.Fatalf("dur %v: moment ends past the source: %+v", c.dur, m)
			}
		}
	}
}

func TestSuggestMomentsTooShort(t *testing.T) {
	if ms := SuggestMoments("p", 3); len(ms) != 0 {
		t.Fatalf("3s source: want no moments, got %d", len(ms))
	}
}
