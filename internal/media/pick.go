package media

import (
	"fmt"
	"math"
)

// Length limits for a moment picked by hand or by an agent, in seconds.
// They are wider than the scorer's own (minClip, maxClip): the picker knows
// what the clip is for.
const (
	MinPick = 5.0
	MaxPick = 180.0
)

// ValidatePick checks a picked window. sourceSec is the recording's length,
// or 0 when it is not known.
func ValidatePick(start, end, sourceSec float64) error {
	switch {
	case math.IsNaN(start) || math.IsNaN(end) || math.IsInf(start, 0) || math.IsInf(end, 0):
		return fmt.Errorf("start and end must be numbers of seconds")
	case start < 0:
		return fmt.Errorf("start must not be negative")
	case end <= start:
		return fmt.Errorf("end must be after start")
	case end-start < MinPick:
		return fmt.Errorf("a clip must be at least %.0f seconds long", MinPick)
	case end-start > MaxPick:
		return fmt.Errorf("a clip must be at most %.0f seconds long", MaxPick)
	case sourceSec > 0 && end > sourceSec+0.5:
		return fmt.Errorf("end (%.1f) is past the end of the recording (%.1f seconds)", end, sourceSec)
	}
	return nil
}
