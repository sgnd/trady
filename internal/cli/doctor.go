package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/sgnd/trady/internal/home"
	"github.com/sgnd/trady/internal/media"
	"github.com/sgnd/trady/internal/out"
	"github.com/sgnd/trady/internal/sealdb"
	"github.com/sgnd/trady/internal/store"
)

type doctorData struct {
	Home    string `json:"home"`
	Store   string `json:"store"`
	KeyMode string `json:"key_mode,omitempty"`
	FFmpeg  string `json:"ffmpeg,omitempty"`
	FFprobe string `json:"ffprobe,omitempty"`

	Network      string     `json:"network"`
	Whisper      string     `json:"whisper,omitempty"`
	WhisperModel string     `json:"whisper_model,omitempty"`
	Warnings     []string   `json:"warnings,omitempty"`
	Next         []nextStep `json:"next,omitempty"`
	OK           bool       `json:"ok"`
}

// nextStep is the one command that fixes one missing tool. trady prints it
// and never runs it: the person decides what gets installed.
type nextStep struct {
	Tool    string `json:"tool"`
	Command string `json:"command"`
}

// tools says what doctor found. NoLibass means ffmpeg is there but has no
// ass filter, so captions cannot be burned in.
type tools struct{ FFmpeg, FFprobe, Whisper, Model, NoLibass bool }

const modelURL = "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/" + media.DefaultModel

// nextSteps lists the command for each missing tool on goos, in the order to
// run them. modelPath is where trady looks for the whisper model.
func nextSteps(goos, modelPath string, have tools) []nextStep {
	var steps []nextStep
	if !have.FFmpeg || !have.FFprobe {
		tool := "ffmpeg"
		if have.FFmpeg {
			tool = "ffprobe" // it comes in the same package
		}
		cmd := "brew install ffmpeg"
		switch goos {
		case "linux":
			cmd = "sudo apt install ffmpeg"
		case "windows":
			cmd = "winget install Gyan.FFmpeg"
		}
		steps = append(steps, nextStep{tool, cmd})
	}
	if have.FFmpeg && have.NoLibass {
		cmd := "install an ffmpeg built with libass, for example: brew tap homebrew-ffmpeg/ffmpeg && brew install homebrew-ffmpeg/ffmpeg/ffmpeg"
		switch goos {
		case "linux":
			cmd = "sudo apt install ffmpeg (the Debian and Ubuntu package has libass)"
		case "windows":
			cmd = "winget install Gyan.FFmpeg (that build has libass)"
		}
		steps = append(steps, nextStep{"libass", cmd})
	}
	if !have.Whisper {
		cmd := "brew install whisper-cpp"
		switch goos {
		case "linux":
			cmd = "git clone https://github.com/ggml-org/whisper.cpp && cd whisper.cpp && cmake -B build && cmake --build build -j --config Release, then put build/bin/whisper-cli on your PATH"
		case "windows":
			cmd = "download whisper-bin-x64.zip from https://github.com/ggml-org/whisper.cpp/releases, unzip it, and put whisper-cli.exe in a folder on your PATH"
		}
		steps = append(steps, nextStep{"whisper-cli", cmd})
	}
	if !have.Model {
		q := shellQuote(modelPath)
		cmd := "mkdir -p " + shellQuote(modelDir(modelPath)) + " && curl -fL -o " + q + " " + modelURL
		if goos == "windows" {
			cmd = "curl.exe -fL -o " + q + " " + modelURL
		}
		steps = append(steps, nextStep{"whisper model", cmd})
	}
	return steps
}

// shellQuote quotes s for a shell only when it needs it.
func shellQuote(s string) string {
	if strings.IndexFunc(s, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || strings.ContainsRune("/._-:~", r))
	}) < 0 {
		return s
	}
	return `"` + strings.NewReplacer(`"`, `\"`, "$", `\$`, "`", "\\`").Replace(s) + `"`
}

func (d doctorData) Text() string {
	var b strings.Builder
	fmt.Fprintf(&b, "home: %s\n", d.Home)
	fmt.Fprintf(&b, "store: %s\n", d.Store)
	if d.KeyMode != "" {
		fmt.Fprintf(&b, "key_mode: %s\n", d.KeyMode)
	}
	fmt.Fprintf(&b, "ffmpeg: %s\n", emptyDash(d.FFmpeg))
	fmt.Fprintf(&b, "ffprobe: %s\n", emptyDash(d.FFprobe))
	fmt.Fprintf(&b, "network: %s\n", d.Network)
	fmt.Fprintf(&b, "whisper: %s\n", emptyDash(d.Whisper))
	fmt.Fprintf(&b, "whisper model: %s\n", emptyDash(d.WhisperModel))
	for _, w := range d.Warnings {
		fmt.Fprintf(&b, "warning: %s\n", w)
	}
	for _, n := range d.Next {
		fmt.Fprintf(&b, "next: %s   (%s)\n", n.Command, n.Tool)
	}
	if d.OK {
		b.WriteString("ok\n")
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func emptyDash(s string) string {
	if s == "" {
		return "(not found)"
	}
	return s
}

func cmdDoctor(args []string) int {
	if hasHelp(args) {
		out.P.Print(usage("doctor") + "\nChecks TRADY_HOME, the sealed store, its key mode, ffmpeg, whisper-cli and the network rule.")
		return 0
	}
	if _, _, ok := parse("doctor", args, nil); !ok {
		return out.ExitUsage
	}
	ctx := context.Background()
	dir, err := home.Ensure()
	if err != nil {
		return out.P.Fail(err)
	}
	d := doctorData{Home: dir, OK: true, Network: networkNote}
	ffmpeg, ffprobe := media.LookPath()
	d.FFmpeg, d.FFprobe = ffmpeg, ffprobe
	if ffmpeg == "" || ffprobe == "" {
		d.Warnings = append(d.Warnings, "install ffmpeg (includes ffprobe) for probe and clip export")
	}
	d.Whisper, d.WhisperModel = media.WhisperPath(), media.ModelPath()
	if d.Whisper == "" || d.WhisperModel == "" {
		d.Warnings = append(d.Warnings, "install whisper-cli and a model for captions: see docs/features/captions.md")
	}
	have := tools{
		FFmpeg: ffmpeg != "", FFprobe: ffprobe != "", Whisper: d.Whisper != "", Model: d.WhisperModel != "",
		NoLibass: ffmpeg != "" && !media.CanBurnCaptions(ctx),
	}
	if w := libassWarning(have); w != "" {
		d.Warnings = append(d.Warnings, w)
	}
	d.Next = nextSteps(runtime.GOOS, media.ModelTarget(), have)

	path := store.Path(store.DBFile)
	meta, err := sealdb.Peek(ctx, path)
	if errors.Is(err, os.ErrNotExist) {
		d.Store = "missing (will be created on first secret/source write)"
		out.P.Print(d)
		return 0
	}
	if err != nil {
		return out.P.Fail(err)
	}
	d.KeyMode = meta["key_mode"]
	db, mode, err := store.Open(ctx, false)
	if err != nil {
		d.Store = "present but cannot open: " + err.Error()
		d.KeyMode = mode
		d.OK = false
		d.Warnings = append(d.Warnings, err.Error())
		out.P.Print(d)
		return 1
	}
	db.Close()
	d.Store = path
	d.KeyMode = mode
	out.P.Print(d)
	return 0
}

// networkNote is what doctor says about the network. A test in cmd/trady
// keeps network packages out of the binary, so this stays true.
const networkNote = "none: trady makes no network calls; ffmpeg, ffprobe and whisper-cli run on local files"

// modelDir is the folder of the model path, cut at its last separator of
// either kind, so the command keeps the path's own separators.
func modelDir(p string) string {
	i := strings.LastIndexAny(p, `/\\`)
	if i < 0 {
		return "."
	}
	return p[:i]
}

// libassWarning says what an ffmpeg without the ass filter means for clips.
func libassWarning(have tools) string {
	if !have.FFmpeg || !have.NoLibass {
		return ""
	}
	return "this ffmpeg has no ass filter (libass): captions are not burned into clips; they are saved as an .srt beside each clip"
}
