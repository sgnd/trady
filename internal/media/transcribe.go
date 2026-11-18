package media

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sgnd/trady/internal/home"
	"github.com/sgnd/trady/internal/inventory"
)

// DefaultModel is the whisper.cpp model file trady looks for under TRADY_HOME/models.
const DefaultModel = "ggml-base.bin"

// WhisperPath finds the whisper.cpp command line tool on PATH.
func WhisperPath() string {
	for _, name := range []string{"whisper-cli", "whisper-cpp"} {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}
	return ""
}

// ModelPath is TRADY_WHISPER_MODEL, or TRADY_HOME/models/ggml-base.bin.
// It returns "" when the file is not there.
func ModelPath() string {
	p := os.Getenv("TRADY_WHISPER_MODEL")
	if p == "" {
		d, err := home.Dir()
		if err != nil {
			return ""
		}
		p = filepath.Join(d, "models", DefaultModel)
	}
	if st, err := os.Stat(p); err != nil || st.IsDir() {
		return ""
	}
	return p
}

// Transcribe turns the audio of src into timed text with whisper.cpp.
// Nothing leaves the machine: ffmpeg and whisper.cpp both run locally.
func Transcribe(ctx context.Context, src string) ([]inventory.Segment, string, error) {
	ffmpeg, _ := LookPath()
	if ffmpeg == "" {
		return nil, "", fmt.Errorf("ffmpeg not found in PATH")
	}
	whisper := WhisperPath()
	if whisper == "" {
		return nil, "", fmt.Errorf("whisper-cli not found in PATH")
	}
	model := ModelPath()
	if model == "" {
		return nil, "", fmt.Errorf("whisper model not found (set TRADY_WHISPER_MODEL or put %s in TRADY_HOME/models)", DefaultModel)
	}
	dir, err := os.MkdirTemp("", "trady-transcribe-")
	if err != nil {
		return nil, "", err
	}
	defer os.RemoveAll(dir)

	wav := filepath.Join(dir, "audio.wav")
	if err := run(ctx, "", ffmpeg, "-y", "-i", src, "-vn", "-ac", "1", "-ar", "16000", "-c:a", "pcm_s16le", wav); err != nil {
		return nil, "", err
	}
	base := filepath.Join(dir, "out")
	if err := run(ctx, "", whisper, "-m", model, "-f", wav, "-l", "auto", "-oj", "-of", base); err != nil {
		return nil, "", err
	}
	b, err := os.ReadFile(base + ".json")
	if err != nil {
		return nil, "", fmt.Errorf("whisper output: %w", err)
	}
	segs, err := ParseWhisperJSON(b)
	return segs, filepath.Base(model), err
}

// ParseWhisperJSON reads the JSON file that whisper-cli writes with -oj.
func ParseWhisperJSON(b []byte) ([]inventory.Segment, error) {
	var raw struct {
		Transcription []struct {
			Offsets struct {
				From int64 `json:"from"`
				To   int64 `json:"to"`
			} `json:"offsets"`
			Text string `json:"text"`
		} `json:"transcription"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, fmt.Errorf("whisper output: %w", err)
	}
	segs := []inventory.Segment{}
	for _, t := range raw.Transcription {
		text := strings.TrimSpace(t.Text)
		if text == "" || t.Offsets.To <= t.Offsets.From {
			continue
		}
		segs = append(segs, inventory.Segment{
			StartSec: float64(t.Offsets.From) / 1000,
			EndSec:   float64(t.Offsets.To) / 1000,
			Text:     text,
		})
	}
	return segs, nil
}

func run(ctx context.Context, dir, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	var errb bytes.Buffer
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(errb.String())
		if len(msg) > 500 {
			msg = msg[len(msg)-500:]
		}
		return fmt.Errorf("%s: %w: %s", filepath.Base(name), err, msg)
	}
	return nil
}
