package media

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// fakeFFmpeg puts an ffmpeg on PATH that writes down its arguments and the
// captions file it was given, so Cut can be checked without a real ffmpeg.
func fakeFFmpeg(t *testing.T) (record string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("shell script stand-in")
	}
	bin := t.TempDir()
	record = filepath.Join(t.TempDir(), "record")
	script := "#!/bin/sh\n" +
		"if [ \"$1\" = -hide_banner ] && [ \"$2\" = -filters ]; then [ -n \"$FAKE_NO_ASS\" ] || printf ' ... ass               V->V       Render ASS subtitles\\n ... subtitles         V->V       Render text subtitles\\n'; exit 0; fi\n" +
		"for a in \"$@\"; do printf '%s\\n' \"$a\" >> '" + record + ".args'; done\n" +
		"for f in captions.ass captions.srt; do [ -f \"$f\" ] && cp \"$f\" '" + record + ".'\"$f\"; done\n" +
		"eval last=\\${$#}; : > \"$last\"\n"
	if err := os.WriteFile(filepath.Join(bin, "ffmpeg"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bin, "ffprobe"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	return record
}

func TestCutPrefersWordCaptions(t *testing.T) {
	rec := fakeFFmpeg(t)
	out := filepath.Join(t.TempDir(), "clip.mp4")
	_, err := Cut(context.Background(), "in.mp4", 1, 4, out, Captions{SRT: "1\n00:00:00,000 --> 00:00:01,000\nline\n\n", ASS: "[Script Info]\nPlayResX: 1080\n"})
	if err != nil {
		t.Fatal(err)
	}
	args, _ := os.ReadFile(rec + ".args")
	if !strings.Contains(string(args), "ass=captions.ass") || strings.Contains(string(args), "subtitles=") {
		t.Errorf("ffmpeg arguments:\n%s", args)
	}
	if b, err := os.ReadFile(rec + ".captions.ass"); err != nil || !strings.Contains(string(b), "PlayResX: 1080") {
		t.Errorf("the ASS file was not given to ffmpeg: %v %q", err, b)
	}
}

func TestCutFallsBackToLineCaptions(t *testing.T) {
	rec := fakeFFmpeg(t)
	out := filepath.Join(t.TempDir(), "clip.mp4")
	if _, err := Cut(context.Background(), "in.mp4", 1, 4, out, Captions{SRT: "1\n00:00:00,000 --> 00:00:01,000\nline\n\n"}); err != nil {
		t.Fatal(err)
	}
	args, _ := os.ReadFile(rec + ".args")
	if !strings.Contains(string(args), "subtitles=captions.srt") || strings.Contains(string(args), "ass=") {
		t.Errorf("ffmpeg arguments:\n%s", args)
	}
}

func TestCutWithoutCaptionsHasNoFilter(t *testing.T) {
	rec := fakeFFmpeg(t)
	out := filepath.Join(t.TempDir(), "clip.mp4")
	if _, err := Cut(context.Background(), "in.mp4", 1, 4, out, Captions{}); err != nil {
		t.Fatal(err)
	}
	args, _ := os.ReadFile(rec + ".args")
	if strings.Contains(string(args), "-vf") {
		t.Errorf("a clip without captions got a video filter:\n%s", args)
	}
}

// TR-b: an ffmpeg without the ass filter (no libass) still cuts the clip,
// without burned captions, and writes the captions as an .srt beside it.
func TestCutWithoutAssFilterWritesSRT(t *testing.T) {
	rec := fakeFFmpeg(t)
	t.Setenv("FAKE_NO_ASS", "1")
	resetFilters()
	t.Cleanup(resetFilters)
	out := filepath.Join(t.TempDir(), "clip.mp4")
	srt := "1\n00:00:00,000 --> 00:00:01,000\nline\n\n"
	burned, err := Cut(context.Background(), "in.mp4", 1, 4, out, Captions{SRT: srt, ASS: "[Script Info]\n"})
	if err != nil {
		t.Fatal(err)
	}
	if burned {
		t.Error("reported burned captions without the ass filter")
	}
	args, _ := os.ReadFile(rec + ".args")
	if strings.Contains(string(args), "-vf") {
		t.Errorf("a caption filter was passed:\n%s", args)
	}
	if b, err := os.ReadFile(strings.TrimSuffix(out, ".mp4") + ".srt"); err != nil || string(b) != srt {
		t.Errorf("no .srt beside the clip: %v %q", err, b)
	}
}

func TestCanBurnCaptions(t *testing.T) {
	fakeFFmpeg(t)
	resetFilters()
	t.Cleanup(resetFilters)
	if !CanBurnCaptions(context.Background()) {
		t.Error("ffmpeg with the ass filter: want true")
	}
	t.Setenv("FAKE_NO_ASS", "1")
	resetFilters()
	if CanBurnCaptions(context.Background()) {
		t.Error("ffmpeg without the ass filter: want false")
	}
}
