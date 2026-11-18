package cli

import (
	"strings"

	"github.com/sgnd/trady/internal/out"
)

// parseArgs splits args into positionals and flag values.
// flags lists the allowed flags; true means the flag takes a value.
// A flag may be written as "--name VALUE" or "--name=VALUE".
func parseArgs(cmd string, args []string, flags map[string]bool) ([]string, map[string]string, *out.Error) {
	pos := []string{}
	vals := map[string]string{}
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			pos = append(pos, args[i+1:]...)
			break
		}
		if !strings.HasPrefix(a, "-") || a == "-" {
			pos = append(pos, a)
			continue
		}
		name, val, hasVal := strings.Cut(a, "=")
		takesValue, known := flags[name]
		if !known {
			return nil, nil, out.Usage("%s: unknown flag %s", cmd, name)
		}
		if !takesValue {
			if hasVal {
				return nil, nil, out.Usage("%s: %s takes no value", cmd, name)
			}
			vals[name] = "true"
			continue
		}
		if !hasVal {
			if i+1 >= len(args) {
				return nil, nil, out.Usage("%s: %s needs a value", cmd, name)
			}
			i++
			val = args[i]
		}
		if val == "" {
			return nil, nil, out.Usage("%s: %s needs a value", cmd, name)
		}
		vals[name] = val
	}
	return pos, vals, nil
}

// want checks the number of positionals: exactly n, with names for the message.
func want(cmd string, pos []string, names ...string) *out.Error {
	switch {
	case len(pos) < len(names):
		return out.Usage("%s needs %s", cmd, strings.Join(names[len(pos):], " "))
	case len(pos) > len(names):
		return out.Usage("%s: unexpected argument %q", cmd, pos[len(names)])
	}
	return nil
}

// parse parses and checks in one step. It returns false after printing the error.
func parse(cmd string, in []string, flags map[string]bool, names ...string) ([]string, map[string]string, bool) {
	pos, vals, err := parseArgs(cmd, in, flags)
	if err == nil {
		err = want(cmd, pos, names...)
	}
	if err != nil {
		out.P.Fail(err)
		return nil, nil, false
	}
	return pos, vals, true
}
