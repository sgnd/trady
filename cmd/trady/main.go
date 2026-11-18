// Command trady turns long local recordings into short captioned clips.
// For people and AI agents. Local-first: state lives under ~/.trady.
package main

import (
	"os"

	"github.com/sgnd/trady/internal/cli"
)

func main() { os.Exit(cli.Execute()) }
