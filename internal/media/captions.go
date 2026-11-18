package media

import (
	"fmt"
	"math"
	"strings"

	"github.com/sgnd/trady/internal/inventory"
)

// SRT builds captions for the window [start,end) of the source.
// Times in the file start at 0, the start of the clip.
// It returns "" when no segment falls in the window.
func SRT(segs []inventory.Segment, start, end float64) string {
	var b strings.Builder
	n := 0
	for _, s := range segs {
		from := math.Max(s.StartSec, start)
		to := math.Min(s.EndSec, end)
		if to-from < 0.05 {
			continue
		}
		n++
		fmt.Fprintf(&b, "%d\n%s --> %s\n%s\n\n", n, srtTime(from-start), srtTime(to-start), plainCaption(s.Text))
	}
	return b.String()
}

func srtTime(sec float64) string {
	ms := int64(math.Round(sec * 1000))
	h := ms / 3600000
	m := ms / 60000 % 60
	s := ms / 1000 % 60
	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, s, ms%1000)
}

// plainCaption drops the characters libass reads as style tags ({\...} and
// <font>), so words in a recording cannot restyle or hide the captions.
func plainCaption(s string) string {
	s = strings.Map(func(r rune) rune {
		switch r {
		case '{', '}', '<', '>', '\\':
			return -1
		}
		return r
	}, s)
	return strings.TrimSpace(s)
}
