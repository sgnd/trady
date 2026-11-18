package media

import (
	"strings"
	"testing"

	"github.com/sgnd/trady/internal/inventory"
)

// talk builds a transcript of 5-second sentences.
func talk(lines ...string) []inventory.Segment {
	segs := make([]inventory.Segment, len(lines))
	for i, l := range lines {
		segs[i] = inventory.Segment{StartSec: float64(i * 5), EndSec: float64(i*5 + 5), Text: l}
	}
	return segs
}

func TestScoreMomentsPicksTheLivelyPart(t *testing.T) {
	var lines []string
	for i := 0; i < 12; i++ {
		lines = append(lines, "We talked about the weather a bit.")
	}
	lines = append(lines,
		"Why does everyone get this wrong?",
		"Honestly, the secret is simple.",
		"Stop doing the first step and learn the rule.",
		"It saved me 40 percent of my time!",
	)
	for i := 0; i < 12; i++ {
		lines = append(lines, "And then we had some lunch.")
	}
	ms := ScoreMoments("p", talk(lines...), 1)
	if len(ms) != 1 {
		t.Fatalf("got %d moments", len(ms))
	}
	m := ms[0]
	// the lively part runs from 60s to 80s
	if m.StartSec > 60 || m.EndSec < 80 {
		t.Fatalf("moment %v-%v misses the lively part: %+v", m.StartSec, m.EndSec, m)
	}
	if m.EndSec-m.StartSec < minClip || m.EndSec-m.StartSec > maxClip {
		t.Fatalf("bad length: %+v", m)
	}
	if m.ProjectID != "p" || m.Title == "" || m.Score < 1 || m.Score > 99 {
		t.Fatalf("bad moment: %+v", m)
	}
}

func TestScoreMomentsNoOverlap(t *testing.T) {
	var lines []string
	for i := 0; i < 60; i++ {
		lines = append(lines, "Why is this the secret rule everyone gets wrong?")
	}
	ms := ScoreMoments("p", talk(lines...), 5)
	if len(ms) != 5 {
		t.Fatalf("got %d moments", len(ms))
	}
	for i := 1; i < len(ms); i++ {
		if ms[i].StartSec < ms[i-1].EndSec {
			t.Fatalf("moments overlap: %+v and %+v", ms[i-1], ms[i])
		}
	}
}

func TestScoreMomentsEmpty(t *testing.T) {
	if ms := ScoreMoments("p", nil, 3); len(ms) != 0 {
		t.Fatalf("no transcript: got %d", len(ms))
	}
	// too short for a clip
	if ms := ScoreMoments("p", talk("Hello.", "Bye."), 3); len(ms) != 0 {
		t.Fatalf("10s transcript: got %d", len(ms))
	}
}

func TestTitle(t *testing.T) {
	if got := title("Short one. More text."); got != "Short one." {
		t.Fatalf("got %q", got)
	}
	long := strings.Repeat("word ", 30)
	if got := title(long); len([]rune(got)) > 61 || !strings.HasSuffix(got, "…") {
		t.Fatalf("got %q", got)
	}
}

func TestMomentCount(t *testing.T) {
	for dur, want := range map[float64]int{0: 1, 300: 2, 3600: 8} {
		if got := MomentCount(dur); got != want {
			t.Errorf("%v: got %d want %d", dur, got, want)
		}
	}
}
