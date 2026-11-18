package cli

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/sgnd/trady/internal/out"
	"github.com/sgnd/trady/internal/vault"
)

func TestRevealAsksInWindow(t *testing.T) {
	t.Setenv("TRADY_HOME", t.TempDir())
	t.Setenv("TRADY_VAULT_PASS", "a-test-passphrase-long-enough")
	if err := vault.Set(t.Context(), "api", "tok-value-123"); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		allowed, shown bool
		wantExit       int
		wantCode       string
	}{
		{false, false, out.ExitPolicy, "NEEDS_PERSON"},
		{false, true, out.ExitPolicy, "DECLINED"},
		{true, true, 0, ""},
	}
	for _, c := range cases {
		old := askInWindow
		var asked string
		askInWindow = func(q, yes string) (bool, bool) { asked = q; return c.allowed, c.shown }
		var b bytes.Buffer
		out.P.Out, out.P.Format = &b, "json"
		code := secretShow(t.Context(), []string{"api", "--reveal"})
		askInWindow = old
		var e struct {
			Data  map[string]any `json:"data"`
			Error *out.Error     `json:"error"`
		}
		_ = json.Unmarshal(b.Bytes(), &e)
		if code != c.wantExit {
			t.Fatalf("%+v: exit %d, want %d", c, code, c.wantExit)
		}
		if c.wantCode != "" && (e.Error == nil || e.Error.Code != c.wantCode) {
			t.Fatalf("%+v: error %+v, want %s", c, e.Error, c.wantCode)
		}
		if c.allowed != bytes.Contains(b.Bytes(), []byte("tok-value-123")) {
			t.Fatalf("%+v: value shown = %v", c, !c.allowed)
		}
		if asked == "" || !bytes.Contains([]byte(asked), []byte(`"api"`)) {
			t.Fatalf("window question does not name the secret: %q", asked)
		}
	}
}
