package out

import (
	"fmt"
	"strings"
)

// SafeText makes text that came from a file name, a recording or a remote
// tool safe to print on a terminal: control characters, including the escape
// that starts terminal commands, are shown as text (\e, \x07). Tabs and
// newlines stay.
func SafeText(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == '\x1b':
			b.WriteString(`\e`)
		case r == '\t' || r == '\n':
			b.WriteRune(r)
		case r < 0x20 || r == 0x7f || r >= 0x80 && r <= 0x9f:
			fmt.Fprintf(&b, `\x%02x`, r)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
