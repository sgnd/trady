// Package media probes and cuts local video with ffprobe/ffmpeg when present.
package media

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/sgnd/trady/internal/inventory"
)

func LookPath() (ffmpeg, ffprobe string) {
	ffmpeg, _ = exec.LookPath("ffmpeg")
	ffprobe, _ = exec.LookPath("ffprobe")
	return ffmpeg, ffprobe
}

func Probe(ctx context.Context, path string) (inventory.Source, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return inventory.Source{}, err
	}
	st, err := os.Stat(abs)
	if err != nil {
		return inventory.Source{}, err
	}
	if st.IsDir() {
		return inventory.Source{}, fmt.Errorf("path is a directory: %s", abs)
	}
	out := inventory.Source{Path: abs}
	_, ffprobe := LookPath()
	if ffprobe == "" {
		out.ProbeNote = "ffprobe not found; duration unknown"
		return out, nil
	}
	cmd := exec.CommandContext(ctx, ffprobe, "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", abs)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		out.ProbeNote = "ffprobe failed: " + strings.TrimSpace(buf.String())
		return out, nil
	}
	var raw struct {
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
		Streams []struct {
			CodecType string `json:"codec_type"`
			CodecName string `json:"codec_name"`
			Width     int    `json:"width"`
			Height    int    `json:"height"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(buf.Bytes(), &raw); err != nil {
		out.ProbeNote = "ffprobe json parse failed"
		return out, nil
	}
	if d, err := strconv.ParseFloat(raw.Format.Duration, 64); err == nil {
		out.DurationSec = d
	}
	for _, s := range raw.Streams {
		if s.CodecType == "video" {
			out.Width, out.Height, out.Codec = s.Width, s.Height, s.CodecName
			break
		}
	}
	return out, nil
}

// Captions are what gets burned into a clip. ASS (word by word) is used when
// it is not empty, else SRT (one line per spoken segment), else nothing.
type Captions struct {
	ASS string
	SRT string
}

// Cut extracts [start,end) seconds to outPath as MP4 (re-encode for reliability)
// and burns in the captions.
func Cut(ctx context.Context, src string, start, end float64, outPath string, caps Captions) error {
	ffmpeg, _ := LookPath()
	if ffmpeg == "" {
		return fmt.Errorf("ffmpeg not found in PATH")
	}
	if end <= start {
		return fmt.Errorf("end must be after start")
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	absOut, err := filepath.Abs(outPath)
	if err != nil {
		return err
	}
	args := []string{
		"-y", "-ss", fmt.Sprintf("%.3f", start), "-i", src,
		"-t", fmt.Sprintf("%.3f", end-start),
	}
	dir := ""
	if caps.ASS != "" || caps.SRT != "" {
		// ffmpeg filter arguments need escaping for most paths, so the
		// captions file gets a plain name in its own folder instead.
		dir, err = os.MkdirTemp("", "trady-captions-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(dir)
		if caps.ASS != "" {
			if err := os.WriteFile(filepath.Join(dir, "captions.ass"), []byte(caps.ASS), 0o600); err != nil {
				return err
			}
			args = append(args, "-vf", "ass=captions.ass")
		} else {
			if err := os.WriteFile(filepath.Join(dir, "captions.srt"), []byte(caps.SRT), 0o600); err != nil {
				return err
			}
			args = append(args, "-vf", "subtitles=captions.srt:force_style='FontSize=18,Outline=2,MarginV=40'")
		}
	}
	args = append(args,
		"-c:v", "libx264", "-preset", "veryfast", "-crf", "23",
		"-c:a", "aac", "-movflags", "+faststart",
		absOut,
	)
	return run(ctx, dir, ffmpeg, args...)
}

// SuggestMoments builds evenly spaced placeholder moments when AI scoring is not available.
func SuggestMoments(projectID string, durationSec float64) []inventory.Moment {
	if durationSec <= 0 {
		durationSec = 60
	}
	const clipLen = 45.0
	var out []inventory.Moment
	n := int(durationSec / 90)
	if n < 1 {
		n = 1
	}
	if n > 8 {
		n = 8
	}
	cats := []string{"story", "quotable", "informational", "hot take"}
	for i := 0; i < n; i++ {
		start := float64(i) * (durationSec / float64(n))
		end := start + clipLen
		if end > durationSec {
			end = durationSec
		}
		if end-start < 5 {
			continue
		}
		score := 90 - i*5
		if score < 50 {
			score = 50
		}
		out = append(out, inventory.Moment{
			ProjectID: projectID,
			Title:     fmt.Sprintf("Moment %d", i+1),
			Category:  cats[i%len(cats)],
			Score:     score,
			StartSec:  start,
			EndSec:    end,
		})
	}
	return out
}

// Default timeout for probe/cut.
func WithTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, 10*time.Minute)
}
