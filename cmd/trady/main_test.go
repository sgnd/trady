package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// envelope is the part of the JSON output the tests check.
type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error *struct {
		Code string `json:"code"`
	} `json:"error"`
}

// buildBinary compiles trady once into a temp folder.
func buildBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "trady")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, b)
	}
	return bin
}

type runner struct {
	t    *testing.T
	bin  string
	home string
	pass string
}

const testSecret = "tok-0123456789-abcdef"

func (r runner) run(args ...string) (envelope, int) {
	r.t.Helper()
	cmd := exec.Command(r.bin, append(args, "-o", "json")...)
	cmd.Env = append(os.Environ(), "TRADY_HOME="+r.home, "TRADY_VAULT_PASS="+r.pass, "TRADY_AGENT=1", "TRADY_NO_DIALOG=1", "TRADY_TEST_TOKEN="+testSecret, "TRADY_TEST_EMPTY=")
	b, err := cmd.Output()
	code := 0
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		code = ee.ExitCode()
	} else if err != nil {
		r.t.Fatalf("%v: %v", args, err)
	}
	if bytes.Contains(b, []byte(testSecret)) {
		r.t.Fatalf("%v: output holds the secret value", args)
	}
	var e envelope
	if err := json.Unmarshal(b, &e); err != nil {
		r.t.Fatalf("%v: bad JSON %q: %v", args, b, err)
	}
	return e, code
}

func (r runner) ok(args ...string) envelope {
	r.t.Helper()
	e, code := r.run(args...)
	if !e.OK || code != 0 {
		r.t.Fatalf("%v: want ok, got exit %d, error %+v", args, code, e.Error)
	}
	return e
}

func (r runner) fail(wantCode string, wantExit int, args ...string) {
	r.t.Helper()
	e, code := r.run(args...)
	if e.OK || e.Error == nil || e.Error.Code != wantCode || code != wantExit {
		r.t.Fatalf("%v: want %s exit %d, got exit %d, error %+v", args, wantCode, wantExit, code, e.Error)
	}
}

func count(t *testing.T, e envelope) int {
	t.Helper()
	var d struct {
		Count int `json:"count"`
	}
	if err := json.Unmarshal(e.Data, &d); err != nil {
		t.Fatal(err)
	}
	return d.Count
}

func TestEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("builds the binary")
	}
	home := t.TempDir()
	r := runner{t: t, bin: buildBinary(t), home: home, pass: "a-test-passphrase-long-enough"}
	media := filepath.Join(home, "ep.mp4")
	if err := os.WriteFile(media, []byte("not really a video"), 0o600); err != nil {
		t.Fatal(err)
	}

	r.ok("version")
	r.ok("doctor")

	// empty lists are [] in JSON, never null
	e := r.ok("source", "list")
	var sl struct {
		Sources []any `json:"sources"`
	}
	if err := json.Unmarshal(e.Data, &sl); err != nil || sl.Sources == nil {
		t.Fatalf("source list: want [], got %s", e.Data)
	}

	r.ok("source", "add", media, "--name", "ep1")
	r.fail("CONFLICT", 8, "source", "add", media, "--name", "ep1")
	r.fail("FILE_NOT_FOUND", 3, "source", "add", filepath.Join(home, "missing.mp4"))

	r.ok("project", "create", "ep1", "--name", "cuts")
	r.fail("CONFLICT", 8, "project", "create", "ep1", "--name", "cuts")
	r.fail("NOT_FOUND", 3, "project", "create", "nope", "--name", "x")

	r.ok("project", "run", "cuts", "--out", filepath.Join(home, "out"))
	if n := count(t, r.ok("moment", "list", "cuts")); n == 0 {
		t.Fatal("moment list: want at least one moment")
	}
	if n := count(t, r.ok("clip", "list", "cuts")); n == 0 {
		t.Fatal("clip list: want at least one clip")
	}
	r.ok("list")
	r.ok("job", "list")

	// running again replaces the moments instead of adding more
	before := count(t, r.ok("moment", "list", "cuts"))
	r.ok("project", "run", "cuts", "--out", filepath.Join(home, "out"))
	if after := count(t, r.ok("moment", "list", "cuts")); after != before {
		t.Fatalf("second run: %d moments, want %d", after, before)
	}

	// removing a project removes its moments and clips
	r.ok("project", "rm", "cuts")
	r.fail("NOT_FOUND", 3, "moment", "list", "cuts")
	r.ok("project", "create", "ep1", "--name", "cuts")
	if n := count(t, r.ok("moment", "list", "cuts")); n != 0 {
		t.Fatalf("moment list after rm and re-create: want 0, got %d", n)
	}

	r.ok("source", "rm", "ep1")
	r.fail("NOT_FOUND", 3, "source", "rm", "ep1")
	r.fail("USAGE", 2, "nosuchcommand")
	r.fail("USAGE", 2, "clip", "export", "cuts")

	// argument checks
	r.fail("USAGE", 2, "source", "list", "extra")
	r.fail("USAGE", 2, "source", "add")
	r.fail("USAGE", 2, "source", "add", media, "another")
	r.fail("USAGE", 2, "source", "add", media, "--name")
	r.fail("USAGE", 2, "source", "add", media, "--bogus")
	r.fail("USAGE", 2, "project", "get", "a", "b")
	r.fail("USAGE", 2, "version", "extra")
	r.fail("USAGE", 2, "doctor", "--bogus")
	r.fail("USAGE", 2, "list", "extra")
	r.fail("SECRET_IN_ARGS", 2, "secret", "set", "TOKEN", "the-value")
	r.ok("source", "add", media, "--name=ep2")
	r.fail("NO_TRANSCRIPT", 3, "source", "transcript", "ep2")
}

func TestShortPassphrase(t *testing.T) {
	if testing.Short() {
		t.Skip("builds the binary")
	}
	home := t.TempDir()
	r := runner{t: t, bin: buildBinary(t), home: home, pass: "short"}
	media := filepath.Join(home, "ep.mp4")
	if err := os.WriteFile(media, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	r.fail("KEY", 6, "source", "add", media)
}

func TestSecrets(t *testing.T) {
	if testing.Short() {
		t.Skip("builds the binary")
	}
	r := runner{t: t, bin: buildBinary(t), home: t.TempDir(), pass: "a-test-passphrase-long-enough"}

	e := r.ok("secret", "list")
	var l struct {
		Secrets []any `json:"secrets"`
	}
	if err := json.Unmarshal(e.Data, &l); err != nil || l.Secrets == nil {
		t.Fatalf("secret list: want [], got %s", e.Data)
	}

	r.ok("secret", "set", "api", "--from-env", "TRADY_TEST_TOKEN")
	r.ok("secret", "set", "api", "--from-env", "TRADY_TEST_TOKEN") // overwrite
	if n := count(t, r.ok("secret", "list")); n != 1 {
		t.Fatalf("secret list: want 1, got %d", n)
	}
	r.ok("secret", "show", "api")

	// an agent can never print or type a value
	r.fail("NEEDS_PERSON", 6, "secret", "show", "api", "--reveal")
	r.fail("NEEDS_PERSON", 6, "secret", "set", "other")
	r.fail("SECRET_IN_ARGS", 2, "secret", "set", "other", testSecret)

	r.fail("USAGE", 2, "secret", "set", "other", "--from-env", "TRADY_TEST_EMPTY")
	r.fail("USAGE", 2, "secret", "set", "bad name", "--from-env", "TRADY_TEST_TOKEN")
	r.fail("SECRET_NOT_FOUND", 3, "secret", "show", "nope")

	r.ok("secret", "rm", "api")
	r.fail("SECRET_NOT_FOUND", 3, "secret", "rm", "api")
}
