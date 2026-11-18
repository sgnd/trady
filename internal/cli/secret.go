package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"golang.org/x/term"

	"github.com/sgnd/trady/internal/out"
	"github.com/sgnd/trady/internal/sealdb"
	"github.com/sgnd/trady/internal/vault"
)

func cmdSecret(args []string) int {
	if len(args) == 0 || hasHelp(args) {
		out.P.Print(usage("secret"))
		return 0
	}
	ctx := context.Background()
	switch args[0] {
	case "set":
		return secretSet(ctx, args[1:])
	case "list":
		if _, _, ok := parse("secret list", args[1:], nil); !ok {
			return out.ExitUsage
		}
		return secretList(ctx)
	case "show":
		return secretShow(ctx, args[1:])
	case "rm", "delete":
		return secretRm(ctx, args[1:])
	default:
		return out.P.Fail(out.Usage("unknown secret subcommand %q", args[0]))
	}
}

func secretSet(ctx context.Context, args []string) int {
	pos, vals, perr := parseArgs("secret set", args, map[string]bool{"--from-env": true})
	if perr != nil {
		return out.P.Fail(perr)
	}
	if len(pos) > 1 {
		return out.P.Fail(out.Errf(out.ExitUsage, "SECRET_IN_ARGS", "use --from-env or stdin; a value on the command line stays in shell history",
			"refusing a secret value on the command line"))
	}
	if len(pos) == 0 {
		return out.P.Fail(out.Usage("secret set needs NAME"))
	}
	name, fromEnv := pos[0], vals["--from-env"]
	if !vault.ValidName(name) {
		return out.P.Fail(out.Usage("secret name %q: use 1 to 64 letters, digits, '_', '-' or '.'", name))
	}
	var value string
	switch {
	case fromEnv != "":
		value = os.Getenv(fromEnv)
		if value == "" {
			return out.P.Fail(out.Usage("env %s is empty", fromEnv))
		}
	default:
		if out.IsAgent() {
			return out.P.Fail(out.NeedsPerson("secret set without --from-env"))
		}
		fmt.Fprintf(os.Stderr, "Paste or type the value for %s, then press Enter (it stays hidden):\n", name)
		b, err := readHidden(os.Stdin)
		if err != nil {
			return out.P.Fail(err)
		}
		value = strings.TrimRight(string(b), "\r\n")
		if value == "" {
			return out.P.Fail(out.Usage("empty secret value"))
		}
	}
	if err := vault.Set(ctx, name, value); err != nil {
		return out.P.Fail(mapErr(err))
	}
	out.P.Print(map[string]any{"name": name, "updated": true})
	return 0
}

// readHidden reads the value without echo at a terminal, or from a pipe.
func readHidden(f *os.File) ([]byte, error) {
	if term.IsTerminal(int(f.Fd())) {
		b, err := term.ReadPassword(int(f.Fd()))
		fmt.Fprintln(os.Stderr)
		return b, err
	}
	return io.ReadAll(io.LimitReader(f, 1<<20))
}

type secretListData struct {
	Count   int `json:"count"`
	Secrets []struct {
		Name      string    `json:"name"`
		UpdatedAt time.Time `json:"updated_at"`
	} `json:"secrets"`
}

func (l secretListData) Text() string {
	if l.Count == 0 {
		return "no secrets yet: trady secret set NAME --from-env VAR"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%-24s %s\n", "NAME", "UPDATED")
	for _, s := range l.Secrets {
		fmt.Fprintf(&b, "%-24s %s\n", s.Name, s.UpdatedAt.Format(time.RFC3339))
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func secretList(ctx context.Context) int {
	items, err := vault.List(ctx)
	if err != nil {
		return out.P.Fail(mapErr(err))
	}
	var d secretListData
	d.Count = len(items)
	d.Secrets = []struct {
		Name      string    `json:"name"`
		UpdatedAt time.Time `json:"updated_at"`
	}{}
	for _, it := range items {
		d.Secrets = append(d.Secrets, struct {
			Name      string    `json:"name"`
			UpdatedAt time.Time `json:"updated_at"`
		}{it.Name, it.UpdatedAt})
	}
	out.P.Print(d)
	return 0
}

func secretShow(ctx context.Context, args []string) int {
	pos, vals, ok := parse("secret show", args, map[string]bool{"--reveal": false}, "NAME")
	if !ok {
		return out.ExitUsage
	}
	name, reveal := pos[0], vals["--reveal"] == "true"
	ok, err := vault.Exists(ctx, name)
	if err != nil {
		return out.P.Fail(mapErr(err))
	}
	if !ok {
		return out.P.Fail(out.Errf(out.ExitNotFound, "SECRET_NOT_FOUND", "trady secret list", "secret %q not found", name))
	}
	if !reveal {
		out.P.Print(map[string]any{"name": name, "exists": true, "value": "[hidden]"})
		return 0
	}
	// The terminal alone cannot tell a person from an AI agent in a
	// pseudo-terminal, so the person says yes in a window of their own.
	allowed, shown := askInWindow("Show the value of the trady secret \""+name+"\" in the terminal?", "Show value")
	if !shown {
		return out.P.Fail(out.NeedsPerson("secret show --reveal").WithHint("run it on a computer with a screen (macOS, or Linux with zenity); trady asks in a window there"))
	}
	if !allowed {
		return out.P.Fail(out.Errf(out.ExitPolicy, "DECLINED", "the value stays hidden", "the person did not allow showing secret %q", name))
	}
	val, err := vault.Get(ctx, name)
	if err != nil {
		return out.P.Fail(mapErr(err))
	}
	out.P.Print(map[string]any{"name": name, "value": val})
	return 0
}

func secretRm(ctx context.Context, args []string) int {
	pos, _, ok := parse("secret rm", args, nil, "NAME")
	if !ok {
		return out.ExitUsage
	}
	name := pos[0]
	if err := vault.Delete(ctx, name); err != nil {
		if errors.Is(err, sealdb.ErrNotFound) {
			return out.P.Fail(out.Errf(out.ExitNotFound, "SECRET_NOT_FOUND", "trady secret list", "secret %q not found", name))
		}
		return out.P.Fail(mapErr(err))
	}
	out.P.Print(map[string]any{"name": name, "removed": true})
	return 0
}
