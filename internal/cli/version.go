package cli

import (
	"github.com/sgnd/trady/internal/out"
)

type versionData struct {
	Version string `json:"version"`
}

func (v versionData) Text() string { return "trady " + v.Version }

func cmdVersion(args []string) int {
	if hasHelp(args) {
		out.P.Print("usage: trady version")
		return 0
	}
	if _, _, ok := parse("version", args, nil); !ok {
		return out.ExitUsage
	}
	out.P.Print(versionData{Version: Version})
	return 0
}

func hasHelp(args []string) bool {
	for _, a := range args {
		if a == "-h" || a == "--help" {
			return true
		}
	}
	return false
}
