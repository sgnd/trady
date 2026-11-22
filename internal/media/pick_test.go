package media

import "testing"

func TestValidatePick(t *testing.T) {
	cases := []struct {
		name               string
		start, end, source float64
		ok                 bool
	}{
		{"normal", 10, 40, 600, true},
		{"from the start", 0, 30, 600, true},
		{"unknown source length", 10, 40, 0, true},
		{"until the very end", 570, 600, 600, true},
		{"too short", 10, 12, 600, false},
		{"too long", 0, 200, 600, false},
		{"backwards", 40, 10, 600, false},
		{"negative start", -5, 20, 600, false},
		{"past the end", 590, 620, 600, false},
		{"not a number", 0, nan(), 600, false},
	}
	for _, c := range cases {
		err := ValidatePick(c.start, c.end, c.source)
		if (err == nil) != c.ok {
			t.Errorf("%s: ValidatePick(%v, %v, %v) = %v, want ok=%v", c.name, c.start, c.end, c.source, err, c.ok)
		}
	}
}

func nan() float64 { var z float64; return z / z }
