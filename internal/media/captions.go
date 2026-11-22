package media

import (
	"fmt"
	"math"
	"regexp"
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
// <font>), so words in a recording cannot restyle or hide the captions, and
// turns line breaks and other control characters into spaces, so a word
// cannot end a caption line or start a new line of the captions file.
func plainCaption(s string) string {
	s = strings.Map(func(r rune) rune {
		switch {
		case r == '{' || r == '}' || r == '<' || r == '>' || r == '\\':
			return -1
		case r < 0x20 || r == 0x7f || r == '\u2028' || r == '\u2029':
			return ' '
		}
		return r
	}, s)
	return strings.Join(strings.Fields(s), " ")
}

var styleBlocks = regexp.MustCompile(`\{[^}]*\}|<[^>]*>`)

// cleanWord makes a spoken word safe for an ASS file: style blocks such as
// {\an8} and <b> are removed first, then any character libass reads as markup.
func cleanWord(s string) string {
	return plainCaption(styleBlocks.ReplaceAllString(s, ""))
}

const wordsPerLine = 4

// ASS builds word-by-word captions for the window [start,end) of the source
// as an Advanced SubStation file: each line shows up to four words, and every
// word fills with colour while it is spoken. Times start at 0, the start of
// the clip. width and height are the video's size (0 when unknown).
// It returns "" when no word with a time falls in the window, so the caller
// can use SRT instead.
func ASS(segs []inventory.Segment, start, end float64, width, height int) string {
	var words []inventory.Word
	for _, s := range segs {
		for _, w := range s.Words {
			if w.EndSec <= start || w.StartSec >= end {
				continue
			}
			text := cleanWord(w.Text)
			if text == "" {
				continue
			}
			words = append(words, inventory.Word{Text: text, StartSec: math.Max(w.StartSec, start) - start, EndSec: math.Min(w.EndSec, end) - start})
		}
	}
	if len(words) == 0 {
		return ""
	}
	if width <= 0 || height <= 0 {
		width, height = 1920, 1080
	}
	var b strings.Builder
	fmt.Fprintf(&b, "[Script Info]\nScriptType: v4.00+\nPlayResX: %d\nPlayResY: %d\nWrapStyle: 2\nScaledBorderAndShadow: yes\n\n", width, height)
	b.WriteString("[V4+ Styles]\nFormat: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding\n")
	// white text that fills with yellow as each word is spoken, black outline
	size := int(math.Round(float64(min(width, height)) * 0.06))
	fmt.Fprintf(&b, "Style: Default,Arial,%d,&H0000D4FF,&H00FFFFFF,&H00000000,&H64000000,-1,0,0,0,100,100,0,0,1,%d,0,2,%d,%d,%d,1\n\n",
		size, max(2, size/18), width/16, width/16, int(math.Round(float64(height)*0.08)))
	b.WriteString("[Events]\nFormat: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text\n")
	for i := 0; i < len(words); {
		j := i + 1
		for j < len(words) && j-i < wordsPerLine && !sentenceEnd(words[j-1].Text) {
			j++
		}
		line := words[i:j]
		var text strings.Builder
		for k, w := range line {
			dur := w.EndSec - w.StartSec
			if k+1 < len(line) {
				dur = line[k+1].StartSec - w.StartSec // a pause belongs to the word before it
			}
			sep := " "
			if k == len(line)-1 {
				sep = ""
			}
			fmt.Fprintf(&text, "{\\kf%d}%s%s", int(math.Round(dur*100)), w.Text, sep)
		}
		fmt.Fprintf(&b, "Dialogue: 0,%s,%s,Default,,0,0,0,,%s\n", assTime(line[0].StartSec), assTime(line[len(line)-1].EndSec), text.String())
		i = j
	}
	return b.String()
}

func assTime(sec float64) string {
	cs := int64(math.Round(sec * 100))
	return fmt.Sprintf("%d:%02d:%02d.%02d", cs/360000, cs/6000%60, cs/100%60, cs%100)
}
