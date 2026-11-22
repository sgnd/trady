package media

import (
	"strings"
	"testing"

	"github.com/sgnd/trady/internal/inventory"
)

// whisper-cli with -ml 1 -sow writes one entry per word.
const perWordJSON = `{"transcription":[
 {"offsets":{"from":0,"to":400},"text":" [_BEG_]"},
 {"offsets":{"from":400,"to":900},"text":" Why"},
 {"offsets":{"from":900,"to":1300},"text":" is"},
 {"offsets":{"from":1300,"to":2000},"text":" this"},
 {"offsets":{"from":2000,"to":2600},"text":" hard"},
 {"offsets":{"from":2600,"to":2700},"text":"?"},
 {"offsets":{"from":2700,"to":3000},"text":" (music)"},
 {"offsets":{"from":3000,"to":3500},"text":" Because"},
 {"offsets":{"from":3500,"to":4000},"text":" nobody"},
 {"offsets":{"from":6000,"to":6500},"text":" checks"},
 {"offsets":{"from":6500,"to":7000},"text":" first."}]}`

func TestParseWhisperWordsGroupsIntoSentences(t *testing.T) {
	segs, err := ParseWhisperJSON([]byte(perWordJSON))
	if err != nil {
		t.Fatal(err)
	}
	if len(segs) != 3 {
		t.Fatalf("want 3 segments (a question, a clause cut by a 2 s pause, the rest), got %d: %+v", len(segs), segs)
	}
	if segs[0].Text != "Why is this hard?" || segs[0].StartSec != 0.4 || segs[0].EndSec != 2.7 {
		t.Errorf("first segment %+v", segs[0])
	}
	if got := len(segs[0].Words); got != 4 {
		t.Fatalf("first segment has %d words, want 4: %+v", got, segs[0].Words)
	}
	if w := segs[0].Words[3]; w.Text != "hard?" || w.StartSec != 2.0 || w.EndSec != 2.7 {
		t.Errorf("the question mark joins its word: %+v", w)
	}
	if segs[1].Text != "Because nobody" || segs[2].Text != "checks first." {
		t.Errorf("segments %q, %q", segs[1].Text, segs[2].Text)
	}
}

func TestParseWhisperSegmentsHaveNoWords(t *testing.T) {
	segs, err := ParseWhisperJSON([]byte(`{"transcription":[{"offsets":{"from":0,"to":2500},"text":" Hello there."}]}`))
	if err != nil || len(segs) != 1 || len(segs[0].Words) != 0 {
		t.Fatalf("%+v %v", segs, err)
	}
}

func TestASSCaptionsForAWindow(t *testing.T) {
	segs, _ := ParseWhisperJSON([]byte(perWordJSON))
	ass := ASS(segs, 0.4, 4.0, 1080, 1920)
	for _, want := range []string{
		"PlayResX: 1080", "PlayResY: 1920", "Style: Default,",
		"Dialogue: 0,0:00:00.00,0:00:02.30,Default,,0,0,0,,{\\kf50}Why {\\kf40}is {\\kf70}this {\\kf70}hard?",
	} {
		if !strings.Contains(ass, want) {
			t.Errorf("ASS lacks %q:\n%s", want, ass)
		}
	}
	if !strings.Contains(ass, "Because") || !strings.Contains(ass, "nobody") {
		t.Error("words after the first sentence are missing")
	}
	if strings.Contains(ass, "checks") {
		t.Error("a word after the window is in the captions")
	}
	if ASS(segs, 100, 130, 1080, 1920) != "" {
		t.Error("a window with no speech must give no captions")
	}
	if ASS([]inventory.Segment{{StartSec: 0, EndSec: 5, Text: "no word times"}}, 0, 5, 1080, 1920) != "" {
		t.Error("a transcript without word times has no word captions")
	}
}

func TestASSGroupsAtMostFourWords(t *testing.T) {
	var words []inventory.Word
	for i := 0; i < 9; i++ {
		words = append(words, inventory.Word{Text: "w", StartSec: float64(i) * 0.3, EndSec: float64(i)*0.3 + 0.3})
	}
	ass := ASS([]inventory.Segment{{StartSec: 0, EndSec: 2.7, Text: "x", Words: words}}, 0, 2.7, 1920, 1080)
	if n := strings.Count(ass, "Dialogue:"); n != 3 {
		t.Errorf("9 words in lines of 4: want 3 dialogue lines, got %d\n%s", n, ass)
	}
}

// Words in a recording must not restyle the captions or run libass commands.
func TestASSDropsStyleTagsFromWords(t *testing.T) {
	seg := inventory.Segment{StartSec: 0, EndSec: 2, Text: "x", Words: []inventory.Word{
		{Text: `{\an8\fs500}Hello`, StartSec: 0, EndSec: 1},
		{Text: `<b>world\N</b>`, StartSec: 1, EndSec: 2},
	}}
	ass := ASS([]inventory.Segment{seg}, 0, 2, 1920, 1080)
	body := ass[strings.Index(ass, "Dialogue:"):]
	if strings.Contains(body, `\an8`) || strings.Contains(body, `\fs500`) || strings.Contains(body, "<b>") || strings.Contains(body, `\N`) {
		t.Errorf("a word restyled the captions: %s", body)
	}
	if !strings.Contains(body, "Hello") || !strings.Contains(body, "world") {
		t.Errorf("the words themselves are gone: %s", body)
	}
}

func TestASSScalesTextWithTheVideo(t *testing.T) {
	seg := inventory.Segment{StartSec: 0, EndSec: 1, Text: "x", Words: []inventory.Word{{Text: "hi", StartSec: 0, EndSec: 1}}}
	if p, l := ASS([]inventory.Segment{seg}, 0, 1, 1080, 1920), ASS([]inventory.Segment{seg}, 0, 1, 1920, 1080); p == l {
		t.Error("portrait and landscape videos need different canvas sizes")
	}
	if !strings.Contains(ASS([]inventory.Segment{seg}, 0, 1, 0, 0), "PlayResY: 1080") {
		t.Error("an unknown size falls back to 1920x1080")
	}
}

// A word or a segment holding a line break must not end the caption line:
// in an ASS file the next line could be read as a new Style.
func TestCaptionsDropLineBreaksAndControlCharacters(t *testing.T) {
	evil := "a\nStyle: Default,Arial,400,&H00FFFFFF\r\n\u2028b\u2029c\x07d"
	seg := inventory.Segment{StartSec: 0, EndSec: 2, Text: evil, Words: []inventory.Word{{Text: evil, StartSec: 0, EndSec: 2}}}

	ass := ASS([]inventory.Segment{seg}, 0, 2, 1920, 1080)
	if strings.Count(ass, "\nStyle:") != 1 {
		t.Errorf("a word started a second Style line:\n%s", ass)
	}
	dialogue := ass[strings.Index(ass, "Dialogue:"):]
	if strings.Count(strings.TrimRight(dialogue, "\n"), "\n") != 0 {
		t.Errorf("the dialogue spans several lines: %q", dialogue)
	}
	if strings.ContainsAny(dialogue, "\x07\u2028\u2029\r") {
		t.Errorf("control characters left: %q", dialogue)
	}

	srt := SRT([]inventory.Segment{seg}, 0, 2)
	blocks := strings.Split(strings.TrimSpace(srt), "\n\n")
	if len(blocks) != 1 || strings.Count(blocks[0], "\n") != 2 {
		t.Errorf("the SRT block was broken: %q", srt)
	}
}
