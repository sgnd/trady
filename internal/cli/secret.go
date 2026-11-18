package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/sgnd/trady/internal/out"
	"github.com/sgnd/trady/internal/sealdb"
	"github.com/sgnd/trady/internal/vault"
)

func cmdSecret(args []string) int {
	if len(args) == 0 || hasHelp(args) {
		out.P.Print(strings.TrimSpace(`
usage: trady secret set NAME [--from-env VAR]
       trady secret list
       trady secret show NAME [--reveal]
       trady secret rm NAME
`) + "\n")
		return 0
	}
	ctx := context.Background()
	switch args[0] {
	case "set":
		return secretSet(ctx, args[1:])
	case "list":
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
	var name, fromEnv string
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--from-env" && i+1 < len(args):
			fromEnv = args[i+1]
			i++
		case strings.HasPrefix(args[i], "-"):
			return out.P.Fail(out.Usage("unknown flag %s", args[i]))
		case name == "":
			name = args[i]
		default:
			return out.P.Fail(out.Errf(out.ExitUsage, "SECRET_IN_ARGS", "use --from-env or stdin; a value on the command line stays in shell history",
				"refusing a secret value on the command line"))
		}
	}
	if name == "" {
		return out.P.Fail(out.Usage("secret set needs a NAME"))
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

func readHidden(r io.Reader) ([]byte, error) {
	// Phase 0/1: plain read; terminal echo hide can come later.
	return io.ReadAll(io.LimitReader(r, 1<<20))
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
	reveal := false
	var name string
	for _, a := range args {
		switch a {
		case "--reveal":
			reveal = true
		default:
			if strings.HasPrefix(a, "-") {
				return out.P.Fail(out.Usage("unknown flag %s", a))
			}
			name = a
		}
	}
	if name == "" {
		return out.P.Fail(out.Usage("secret show needs a NAME"))
	}
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
	if out.IsAgent() {
		return out.P.Fail(out.NeedsPerson("secret show --reveal"))
	}
	val, err := vault.Get(ctx, name)
	if err != nil {
		return out.P.Fail(mapErr(err))
	}
	out.P.Print(map[string]any{"name": name, "value": val})
	return 0
}

func secretRm(ctx context.Context, args []string) int {
	if len(args) == 0 {
		return out.P.Fail(out.Usage("secret rm needs a NAME"))
	}
	name := args[0]
	if err := vault.Delete(ctx, name); err != nil {
		if errors.Is(err, sealdb.ErrNotFound) {
			return out.P.Fail(out.Errf(out.ExitNotFound, "SECRET_NOT_FOUND", "trady secret list", "secret %q not found", name))
		}
		return out.P.Fail(mapErr(err))
	}
	out.P.Print(map[string]any{"name": name, "removed": true})
	return 0
}
