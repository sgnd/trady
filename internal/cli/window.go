package cli

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// askInWindow asks the person at this computer in a window of their own,
// which an AI agent in a terminal cannot answer. shown is false when no
// window can be shown (no screen, SSH, TRADY_NO_DIALOG=1, no tool for it).
// Tests replace it.
var askInWindow = func(question, yes string) (allowed, shown bool) {
	if os.Getenv("TRADY_NO_DIALOG") != "" || os.Getenv("SSH_CONNECTION") != "" || os.Getenv("SSH_TTY") != "" {
		return false, false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	switch runtime.GOOS {
	case "darwin":
		const script = `on run argv
set r to display dialog (item 1 of argv) with title "trady" buttons {"Cancel", (item 2 of argv)} default button "Cancel" cancel button "Cancel" giving up after 120 with icon caution
if gave up of r then return "gave up"
return button returned of r
end run`
		b, err := exec.CommandContext(ctx, "/usr/bin/osascript", "-e", script, question, yes).Output()
		var ee *exec.ExitError
		if err != nil && !errors.As(err, &ee) {
			return false, false
		}
		return strings.TrimSpace(string(b)) == yes, true
	case "linux":
		zenity, err := exec.LookPath("zenity")
		if err != nil || (os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "") {
			return false, false
		}
		err = exec.CommandContext(ctx, zenity, "--question", "--title=trady", "--text="+question, "--ok-label="+yes, "--cancel-label=Cancel", "--timeout=120").Run()
		return err == nil, true
	}
	return false, false
}
