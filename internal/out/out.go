// Package out renders every command result as a stable envelope.
package out

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/term"
)

const (
	ExitOK       = 0
	ExitError    = 1
	ExitUsage    = 2
	ExitNotFound = 3
	ExitPolicy   = 6
	ExitConflict = 8
)

const Contract = 1

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Hint    string `json:"hint,omitempty"`
	Exit    int    `json:"-"`
}

func (e *Error) Error() string { return e.Code + ": " + e.Message }
func (e *Error) WithHint(h string) *Error {
	e.Hint = h
	return e
}

func Errf(exit int, code, hint, format string, a ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, a...), Hint: hint, Exit: exit}
}

func Usage(format string, a ...any) *Error {
	return Errf(ExitUsage, "USAGE", "run with --help or `trady schema` for the command reference", format, a...)
}

type Envelope struct {
	OK      bool   `json:"ok"`
	Command string `json:"command"`
	Data    any    `json:"data,omitempty"`
	Error   *Error `json:"error,omitempty"`
	Meta    Meta   `json:"meta"`
}

type Meta struct {
	Contract   int    `json:"contract"`
	Version    string `json:"version"`
	DurationMS int64  `json:"duration_ms"`
	DryRun     bool   `json:"dry_run,omitempty"`
}

type Texter interface{ Text() string }

type Printer struct {
	mu      sync.Mutex
	Out     io.Writer
	Err     io.Writer
	Format  string // json | text | auto
	Version string
	start   time.Time
	cmd     string
}

var P = &Printer{Out: os.Stdout, Err: os.Stderr, Format: "auto", start: time.Now()}

func (p *Printer) Reset(command string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cmd = command
	p.start = time.Now()
}

func (p *Printer) useJSON() bool {
	f := p.Format
	if f == "" || f == "auto" {
		if os.Getenv("TRADY_OUTPUT") == "json" {
			return true
		}
		return !term.IsTerminal(int(os.Stdout.Fd()))
	}
	return f == "json"
}

func (p *Printer) meta() Meta {
	return Meta{
		Contract:   Contract,
		Version:    p.Version,
		DurationMS: time.Since(p.start).Milliseconds(),
	}
}

func (p *Printer) Print(data any) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.useJSON() {
		enc := json.NewEncoder(p.Out)
		enc.SetEscapeHTML(false)
		_ = enc.Encode(Envelope{OK: true, Command: p.cmd, Data: data, Meta: p.meta()})
		return
	}
	switch t := data.(type) {
	case Texter:
		fmt.Fprintln(p.Out, t.Text())
	case string:
		fmt.Fprintln(p.Out, t)
	case nil:
	default:
		b, _ := json.MarshalIndent(data, "", "  ")
		fmt.Fprintln(p.Out, string(b))
	}
}

func (p *Printer) Fail(err error) int {
	p.mu.Lock()
	defer p.mu.Unlock()
	e, ok := err.(*Error)
	if !ok {
		e = Errf(ExitError, "ERROR", "read the message; `trady doctor` checks this computer", "%v", err)
	}
	if e.Exit == 0 {
		e.Exit = ExitError
	}
	if p.useJSON() {
		enc := json.NewEncoder(p.Out)
		enc.SetEscapeHTML(false)
		_ = enc.Encode(Envelope{OK: false, Command: p.cmd, Error: e, Meta: p.meta()})
	} else {
		fmt.Fprintf(p.Err, "%s: %s\n", e.Code, e.Message)
		if e.Hint != "" {
			fmt.Fprintf(p.Err, "hint: %s\n", e.Hint)
		}
	}
	return e.Exit
}

// IsAgent is true when prompts must not run.
func IsAgent() bool {
	if os.Getenv("TRADY_AGENT") == "1" {
		return true
	}
	return !term.IsTerminal(int(os.Stdin.Fd()))
}

// NeedsPerson stops an agent that needs a human at a TTY.
func NeedsPerson(what string) *Error {
	return Errf(ExitPolicy, "NEEDS_PERSON", "ask a person to run this in their own terminal",
		"%s needs a person at a terminal", what)
}

func ParseOutputFlag(args []string) (rest []string, format string) {
	format = "auto"
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-o" || a == "--output":
			if i+1 < len(args) {
				format = args[i+1]
				i++
			}
		case strings.HasPrefix(a, "-o="):
			format = strings.TrimPrefix(a, "-o=")
		case strings.HasPrefix(a, "--output="):
			format = strings.TrimPrefix(a, "--output=")
		default:
			rest = append(rest, a)
		}
	}
	return rest, format
}
