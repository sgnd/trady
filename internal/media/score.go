package media

import (
	"sort"
	"strings"
	"unicode"

	"github.com/sgnd/trady/internal/inventory"
)

// Clip length limits, in seconds.
const (
	minClip = 15.0
	maxClip = 60.0
)

// Words that often mark a moment worth sharing, by category.
var cues = map[string][]string{
	"hot take":      {"wrong", "never", "always", "honestly", "truth", "overrated", "underrated", "mistake", "stop", "hate", "unpopular"},
	"informational": {"how", "why", "step", "tip", "learn", "because", "means", "percent", "number", "first", "second"},
	"story":         {"when i", "i was", "remember", "then", "suddenly", "happened", "story", "years ago", "one day"},
	"quotable":      {"secret", "key", "best", "worst", "most", "everyone", "nobody", "simple", "rule"},
}

// ScoreMoments picks up to max moments from a transcript. It runs locally
// and needs no model: it builds windows on sentence edges, scores each by
// cue words, questions, numbers and how dense the speech is, and keeps the
// best windows that do not overlap. Without segments it returns nil.
func ScoreMoments(projectID string, segs []inventory.Segment, max int) []inventory.Moment {
	type cand struct {
		start, end float64
		text       string
		score      float64
		cat        string
	}
	var cands []cand
	for i := range segs {
		var text strings.Builder
		for j := i; j < len(segs); j++ {
			dur := segs[j].EndSec - segs[i].StartSec
			if dur > maxClip {
				break
			}
			text.WriteString(segs[j].Text)
			text.WriteString(" ")
			if dur < minClip || !sentenceEnd(segs[j].Text) {
				continue
			}
			t := strings.TrimSpace(text.String())
			score, cat := scoreText(t, dur)
			cands = append(cands, cand{segs[i].StartSec, segs[j].EndSec, t, score, cat})
		}
	}
	sort.SliceStable(cands, func(a, b int) bool { return cands[a].score > cands[b].score })

	var picked []cand
	for _, c := range cands {
		if len(picked) == max {
			break
		}
		overlaps := false
		for _, p := range picked {
			if c.start < p.end && p.start < c.end {
				overlaps = true
				break
			}
		}
		if !overlaps {
			picked = append(picked, c)
		}
	}
	sort.Slice(picked, func(a, b int) bool { return picked[a].start < picked[b].start })

	out := make([]inventory.Moment, 0, len(picked))
	for _, c := range picked {
		out = append(out, inventory.Moment{
			ProjectID: projectID,
			Title:     title(c.text),
			Category:  c.cat,
			Score:     clamp(int(c.score), 1, 99),
			StartSec:  c.start,
			EndSec:    c.end,
		})
	}
	return out
}

func scoreText(text string, dur float64) (float64, string) {
	lower := strings.ToLower(text)
	words := strings.Fields(lower)
	score := 40.0
	best, bestHits := "story", 0
	for _, cat := range []string{"hot take", "informational", "story", "quotable"} {
		hits := 0
		for _, cue := range cues[cat] {
			hits += countPhrase(words, cue)
		}
		score += float64(min(hits, 4)) * 6
		if hits > bestHits {
			best, bestHits = cat, hits
		}
	}
	score += float64(min(strings.Count(text, "?"), 2)) * 5
	score += float64(min(strings.Count(text, "!"), 2)) * 3
	if strings.IndexFunc(text, unicode.IsDigit) >= 0 {
		score += 5
	}
	// lively speech (about 2.5 words per second or more) scores higher
	rate := float64(len(words)) / dur
	score += min(rate, 3) * 5
	// a window that opens with a question or a cue word hooks the viewer
	if first := firstSentence(text); strings.HasSuffix(first, "?") {
		score += 6
	}
	return score, best
}

func countPhrase(words []string, phrase string) int {
	p := strings.Fields(phrase)
	n := 0
	for i := 0; i+len(p) <= len(words); i++ {
		ok := true
		for k, w := range p {
			if strings.TrimFunc(words[i+k], notLetter) != w {
				ok = false
				break
			}
		}
		if ok {
			n++
		}
	}
	return n
}

func notLetter(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) }

func sentenceEnd(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasSuffix(s, ".") || strings.HasSuffix(s, "?") || strings.HasSuffix(s, "!")
}

func firstSentence(s string) string {
	if i := strings.IndexAny(s, ".?!"); i >= 0 {
		return s[:i+1]
	}
	return s
}

// title is the first sentence, cut to 60 characters on a word edge.
func title(text string) string {
	t := strings.TrimSpace(firstSentence(text))
	if len([]rune(t)) <= 60 {
		return t
	}
	r := []rune(t)[:60]
	if i := strings.LastIndex(string(r), " "); i > 20 {
		return strings.TrimSpace(string(r)[:i]) + "…"
	}
	return string(r) + "…"
}

func clamp(v, lo, hi int) int {
	return max(lo, min(v, hi))
}

// MomentCount is how many moments to look for in a source of this length.
func MomentCount(durationSec float64) int {
	return clamp(int(durationSec/120), 1, 8)
}
