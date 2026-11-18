package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/sgnd/trady/internal/home"
	"github.com/sgnd/trady/internal/media"
	"github.com/sgnd/trady/internal/out"
	"github.com/sgnd/trady/internal/sealdb"
	"github.com/sgnd/trady/internal/store"
)

type doctorData struct {
	Home     string   `json:"home"`
	Store    string   `json:"store"`
	KeyMode  string   `json:"key_mode,omitempty"`
	FFmpeg   string   `json:"ffmpeg,omitempty"`
	FFprobe  string   `json:"ffprobe,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
	OK       bool     `json:"ok"`
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
	for _, w := range d.Warnings {
		fmt.Fprintf(&b, "warning: %s\n", w)
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
		out.P.Print("usage: trady doctor\nChecks TRADY_HOME, sealed store, key mode, and ffmpeg/ffprobe.")
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
	d := doctorData{Home: dir, OK: true}
	ffmpeg, ffprobe := media.LookPath()
	d.FFmpeg, d.FFprobe = ffmpeg, ffprobe
	if ffmpeg == "" || ffprobe == "" {
		d.Warnings = append(d.Warnings, "install ffmpeg (includes ffprobe) for probe and clip export")
	}

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
