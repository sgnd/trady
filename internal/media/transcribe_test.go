package media

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sgnd/trady/internal/inventory"
)

func TestParseWhisperJSON(t *testing.T) {
	b := []byte(`{"transcription":[
		{"offsets":{"from":0,"to":2500},"text":" Hello there."},
		{"offsets":{"from":2500,"to":2500},"text":" zero length"},
		{"offsets":{"from":3000,"to":4200},"text":"   "},
		{"offsets":{"from":4200,"to":6000},"text":" Second line"}]}`)
	segs, err := ParseWhisperJSON(b)
	if err != nil {
		t.Fatal(err)
	}
	if len(segs) != 2 || segs[0].Text != "Hello there." || segs[1].StartSec != 4.2 || segs[1].EndSec != 6 {
		t.Fatalf("got %+v", segs)
	}
	if _, err := ParseWhisperJSON([]byte("nope")); err == nil {
		t.Fatal("bad JSON: want error")
	}
}

func TestSRT(t *testing.T) {
	segs := []inventory.Segment{
		{StartSec: 5, EndSec: 9, Text: "before and inside"},
		{StartSec: 12, EndSec: 13.5, Text: "inside"},
		{StartSec: 100, EndSec: 101, Text: "after"},
	}
	got := SRT(segs, 8, 3608)
	want := "1\n00:00:00,000 --> 00:00:01,000\nbefore and inside\n\n" +
		"2\n00:00:04,000 --> 00:00:05,500\ninside\n\n" +
		"3\n00:01:32,000 --> 00:01:33,000\nafter\n\n"
	if got != want {
		t.Fatalf("got\n%q\nwant\n%q", got, want)
	}
	if SRT(segs, 200, 245) != "" {
		t.Fatal("window with no speech: want empty")
	}
	if srtTime(3725.5) != "01:02:05,500" {
		t.Fatalf("srtTime: %s", srtTime(3725.5))
	}
}

// TestCutWithCaptions renders a real clip. CI sets TRADY_REQUIRE_FFMPEG=1
// so the test runs there; elsewhere it is skipped when ffmpeg is missing.
func TestCutWithCaptions(t *testing.T) {
	ffmpeg, _ := LookPath()
	if ffmpeg == "" {
		if os.Getenv("TRADY_REQUIRE_FFMPEG") == "1" {
			t.Fatal("ffmpeg is required here")
		}
		t.Skip("ffmpeg not installed")
	}
	dir := t.TempDir()
	src := filepath.Join(dir, "sample with space.mp4")
	gen := exec.Command(ffmpeg, "-y", "-f", "lavfi", "-i", "testsrc=size=320x240:rate=25:duration=6",
		"-f", "lavfi", "-i", "sine=frequency=440:duration=6", "-shortest",
		"-c:v", "libx264", "-c:a", "aac", src)
	if b, err := gen.CombinedOutput(); err != nil {
		t.Fatalf("make sample: %v\n%s", err, b)
	}
	probe, err := Probe(context.Background(), src)
	if err != nil || probe.DurationSec < 5 {
		t.Fatalf("probe: %+v, %v", probe, err)
	}
	clip := filepath.Join(dir, "out", "clip.mp4")
	srt := SRT([]inventory.Segment{{StartSec: 1, EndSec: 3, Text: "Hello: it's 'quoted'"}}, 1, 4)
	if err := Cut(context.Background(), src, 1, 4, clip, Captions{SRT: srt}); err != nil {
		if strings.Contains(err.Error(), "subtitles") && os.Getenv("TRADY_REQUIRE_FFMPEG") != "1" {
			t.Skip("this ffmpeg has no subtitles filter (libass)")
		}
		t.Fatal(err)
	}
	got, err := Probe(context.Background(), clip)
	if err != nil || got.DurationSec < 2.5 || got.DurationSec > 3.5 {
		t.Fatalf("clip: %+v, %v", got, err)
	}
}

// TestCutWithWordCaptions burns word-by-word captions with a real ffmpeg. It
// runs where TRADY_REQUIRE_FFMPEG=1 (CI) and is skipped when ffmpeg or its
// ass filter (libass) is missing.
func TestCutWithWordCaptions(t *testing.T) {
	ffmpeg, _ := LookPath()
	if ffmpeg == "" {
		if os.Getenv("TRADY_REQUIRE_FFMPEG") == "1" {
			t.Fatal("ffmpeg is required here")
		}
		t.Skip("ffmpeg not installed")
	}
	dir := t.TempDir()
	src := filepath.Join(dir, "sample.mp4")
	gen := exec.Command(ffmpeg, "-y", "-f", "lavfi", "-i", "testsrc=size=320x240:rate=25:duration=6",
		"-f", "lavfi", "-i", "sine=frequency=440:duration=6", "-shortest",
		"-c:v", "libx264", "-c:a", "aac", src)
	if b, err := gen.CombinedOutput(); err != nil {
		t.Fatalf("make sample: %v\n%s", err, b)
	}
	segs := []inventory.Segment{{StartSec: 1, EndSec: 3, Text: "Hello there friend", Words: []inventory.Word{
		{Text: "Hello", StartSec: 1, EndSec: 1.6}, {Text: "there", StartSec: 1.6, EndSec: 2.2}, {Text: "friend.", StartSec: 2.2, EndSec: 3}}}}
	clip := filepath.Join(dir, "out", "clip.mp4")
	caps := Captions{ASS: ASS(segs, 1, 4, 320, 240)}
	if err := Cut(context.Background(), src, 1, 4, clip, caps); err != nil {
		if strings.Contains(err.Error(), "ass") && os.Getenv("TRADY_REQUIRE_FFMPEG") != "1" {
			t.Skip("this ffmpeg has no ass filter (libass)")
		}
		t.Fatal(err)
	}
	got, err := Probe(context.Background(), clip)
	if err != nil || got.DurationSec < 2.5 || got.DurationSec > 3.5 {
		t.Fatalf("clip: %+v, %v", got, err)
	}
}

func TestSRTDropsStyleTags(t *testing.T) {
	segs := []inventory.Segment{{StartSec: 0, EndSec: 2, Text: `{\fs200\alpha&HFF&}hidden <font color="red">red</font> a\Nb`}}
	lines := strings.Split(SRT(segs, 0, 2), "\n")
	got := lines[2] // the caption text line, after the number and the times
	for _, bad := range []string{"{", "}", "<", ">", `\`} {
		if strings.Contains(got, bad) {
			t.Fatalf("caption keeps %q: %q", bad, got)
		}
	}
	if !strings.Contains(got, "hidden") || !strings.Contains(got, "red") {
		t.Fatalf("caption lost its words: %q", got)
	}
}

// TestTranscribeReal runs ffmpeg and whisper-cli on speech made by espeak-ng.
// CI sets TRADY_REQUIRE_WHISPER=1 and TRADY_TEST_SPEECH to such a file;
// elsewhere the test is skipped.
func TestTranscribeReal(t *testing.T) {
	speech := os.Getenv("TRADY_TEST_SPEECH")
	if speech == "" || WhisperPath() == "" || ModelPath() == "" {
		if os.Getenv("TRADY_REQUIRE_WHISPER") == "1" {
			t.Fatal("whisper-cli, a model and TRADY_TEST_SPEECH are required here")
		}
		t.Skip("whisper-cli, a model or TRADY_TEST_SPEECH missing")
	}
	segs, model, err := Transcribe(context.Background(), speech)
	if err != nil {
		t.Fatal(err)
	}
	var text strings.Builder
	for _, s := range segs {
		text.WriteString(strings.ToLower(s.Text) + " ")
	}
	if model == "" || !strings.Contains(text.String(), "clip") {
		t.Fatalf("model %q, transcript %q: want the word clip", model, text.String())
	}
}
