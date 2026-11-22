package cli

import (
	"strings"
	"testing"
)

func TestNextStepsNameTheExactCommand(t *testing.T) {
	model := "/home/me/.trady/models/ggml-base.bin"
	cases := []struct {
		name  string
		goos  string
		have  tools
		wants []string // each must appear in some step's command, in this order of tools
		tools []string
	}{
		{"everything there", "darwin", tools{FFmpeg: true, FFprobe: true, Whisper: true, Model: true}, nil, nil},
		{"mac, nothing", "darwin", tools{}, []string{"brew install ffmpeg", "brew install whisper-cpp", "curl -fL -o " + model}, []string{"ffmpeg", "whisper-cli", "whisper model"}},
		{"linux, only the model", "linux", tools{FFmpeg: true, FFprobe: true, Whisper: true}, []string{"mkdir -p /home/me/.trady/models && curl -fL -o " + model}, []string{"whisper model"}},
		{"linux, no whisper", "linux", tools{FFmpeg: true, FFprobe: true, Model: true}, []string{"cmake -B build"}, []string{"whisper-cli"}},
		{"linux, no ffmpeg", "linux", tools{Whisper: true, Model: true}, []string{"sudo apt install ffmpeg"}, []string{"ffmpeg"}},
		{"windows", "windows", tools{}, []string{"winget install Gyan.FFmpeg", "whisper-bin-x64.zip", "curl.exe -fL -o"}, []string{"ffmpeg", "whisper-cli", "whisper model"}},
		{"ffprobe missing means ffmpeg install", "darwin", tools{FFmpeg: true, Whisper: true, Model: true}, []string{"brew install ffmpeg"}, []string{"ffprobe"}},
	}
	for _, c := range cases {
		steps := nextSteps(c.goos, model, c.have)
		if len(steps) != len(c.wants) {
			t.Errorf("%s: %d steps, want %d: %+v", c.name, len(steps), len(c.wants), steps)
			continue
		}
		for i, s := range steps {
			if !strings.Contains(s.Command, c.wants[i]) || s.Tool != c.tools[i] {
				t.Errorf("%s: step %d = %+v, want tool %q and command with %q", c.name, i, s, c.tools[i], c.wants[i])
			}
		}
	}
}

func TestDoctorTextShowsNextCommands(t *testing.T) {
	d := doctorData{Home: "h", Network: "none", OK: true, Next: []nextStep{{Tool: "ffmpeg", Command: "brew install ffmpeg"}}}
	got := d.Text()
	if !strings.Contains(got, "next: brew install ffmpeg") || !strings.Contains(got, "ffmpeg") {
		t.Errorf("text lacks the next command:\n%s", got)
	}
}

func TestShellQuote(t *testing.T) {
	cases := map[string]string{
		"/home/me/.trady/models/m.bin": "/home/me/.trady/models/m.bin",
		"/home/me/my models/m.bin":     `"/home/me/my models/m.bin"`,
		`/tmp/a"b`:                     `"/tmp/a\"b"`,
		"/tmp/$HOME":                   `"/tmp/\$HOME"`,
		"/tmp/a`id`":                   "\"/tmp/a\\`id\\`\"",
	}
	for in, want := range cases {
		if got := shellQuote(in); got != want {
			t.Errorf("shellQuote(%q) = %s, want %s", in, got, want)
		}
	}
}

// The model folder comes from the model path whatever the path's separators,
// so a Linux path stays a Linux path when the test runs on Windows.
func TestModelFolderKeepsTheSeparators(t *testing.T) {
	cases := map[string]string{
		"/home/me/.trady/models/ggml-base.bin":    "/home/me/.trady/models",
		`C:\Users\me\.trady\models\ggml-base.bin`: `C:\Users\me\.trady\models`,
		"ggml-base.bin": ".",
	}
	for in, want := range cases {
		if got := modelDir(in); got != want {
			t.Errorf("modelDir(%q) = %q, want %q", in, got, want)
		}
	}
}
