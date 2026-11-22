// Package skills holds the agent skill that ships inside the binary.
package skills

import _ "embed"

// Trady is skills/trady/SKILL.md. `trady guide install` writes it.
//
//go:embed trady/SKILL.md
var Trady string
