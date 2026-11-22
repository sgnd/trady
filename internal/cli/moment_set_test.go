package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/sgnd/trady/internal/inventory"
	"github.com/sgnd/trady/internal/out"
)

func trady(t *testing.T, run func([]string) int, name string, args ...string) (int, map[string]any) {
	t.Helper()
	old := out.P
	var buf bytes.Buffer
	out.P = &out.Printer{Out: &buf, Err: &bytes.Buffer{}, Format: "json"}
	out.P.Reset(name)
	t.Cleanup(func() { out.P = old })
	code := run(args)
	env := map[string]any{}
	if buf.Len() > 0 {
		if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
			t.Fatalf("not JSON: %q", buf.String())
		}
	}
	return code, env
}

func projectWithSource(t *testing.T, durationSec float64) (context.Context, inventory.Project, inventory.Source) {
	t.Helper()
	t.Setenv("TRADY_HOME", t.TempDir())
	t.Setenv("TRADY_VAULT_PASS", "")
	t.Setenv("TRADY_KEYCHAIN", "off")
	ctx := context.Background()
	src, err := inventory.AddSource(ctx, "ep1", "/media/ep1.mp4", inventory.Source{DurationSec: durationSec})
	if err != nil {
		t.Fatal(err)
	}
	p, err := inventory.CreateProject(ctx, "ep1-cuts", src.ID)
	if err != nil {
		t.Fatal(err)
	}
	return ctx, p, src
}

func TestMomentSetPicksAndClears(t *testing.T) {
	ctx, p, _ := projectWithSource(t, 600)

	code, env := trady(t, cmdMoment, "moment", "set", "ep1-cuts", "--start", "30", "--end", "62.5", "--title", "Why the plan failed")
	if code != 0 {
		t.Fatalf("exit %d: %v", code, env)
	}
	ms, _ := inventory.ListMoments(ctx, p.ID)
	if len(ms) != 1 || !ms[0].Picked || ms[0].StartSec != 30 || ms[0].EndSec != 62.5 || ms[0].Title != "Why the plan failed" {
		t.Fatalf("moments %+v", ms)
	}

	code, env = trady(t, cmdMoment, "moment", "list", "ep1-cuts")
	data, _ := env["data"].(map[string]any)
	if code != 0 || data["count"].(float64) != 1 {
		t.Errorf("list: exit %d %v", code, env)
	}

	if code, _ := trady(t, cmdMoment, "moment", "set", "ep1-cuts", "--clear"); code != 0 {
		t.Fatalf("clear exit %d", code)
	}
	if ms, _ := inventory.ListMoments(ctx, p.ID); len(ms) != 0 {
		t.Errorf("after --clear: %+v", ms)
	}
}

func TestMomentSetRefusesBadInput(t *testing.T) {
	projectWithSource(t, 600)
	cases := []struct {
		name string
		args []string
		exit int
		code string
	}{
		{"no range", []string{"set", "ep1-cuts"}, out.ExitUsage, "USAGE"},
		{"only start", []string{"set", "ep1-cuts", "--start", "5"}, out.ExitUsage, "USAGE"},
		{"not a number", []string{"set", "ep1-cuts", "--start", "abc", "--end", "40"}, out.ExitUsage, "USAGE"},
		{"too short", []string{"set", "ep1-cuts", "--start", "10", "--end", "11"}, out.ExitUsage, "USAGE"},
		{"past the end", []string{"set", "ep1-cuts", "--start", "590", "--end", "640"}, out.ExitUsage, "USAGE"},
		{"clear with a range", []string{"set", "ep1-cuts", "--clear", "--start", "10", "--end", "40"}, out.ExitUsage, "USAGE"},
		{"no such project", []string{"set", "nope", "--start", "10", "--end", "40"}, out.ExitNotFound, "NOT_FOUND"},
	}
	for _, c := range cases {
		code, env := trady(t, cmdMoment, "moment", c.args...)
		e, _ := env["error"].(map[string]any)
		if code != c.exit || e["code"] != c.code {
			t.Errorf("%s: exit %d, error %v", c.name, code, e)
		}
	}
}

func TestPickedMomentsReplaceScoring(t *testing.T) {
	ctx, p, src := projectWithSource(t, 600)
	segs := []inventory.Segment{{StartSec: 0, EndSec: 30, Text: "Honestly, why is this the biggest mistake?"}, {StartSec: 30, EndSec: 60, Text: "Because nobody checks the first step."}}

	got, how, err := chooseMoments(ctx, p, src, segs)
	if err != nil || len(got) == 0 || how == "" || got[0].Picked {
		t.Fatalf("without picks: %+v %q %v", got, how, err)
	}
	if _, err := inventory.AddMoment(ctx, inventory.Moment{ProjectID: p.ID, Title: "mine", StartSec: 100, EndSec: 140, Picked: true}); err != nil {
		t.Fatal(err)
	}
	got, how, err = chooseMoments(ctx, p, src, segs)
	if err != nil || len(got) != 1 || got[0].Title != "mine" || !got[0].Picked {
		t.Fatalf("with a pick: %+v %v", got, err)
	}
	if want := "picked with `trady moment set`"; how != want {
		t.Errorf("selection note %q, want %q", how, want)
	}
}

func TestCaptionNoteSaysWhichStyle(t *testing.T) {
	words := []inventory.Segment{{Text: "hi", Words: []inventory.Word{{Text: "hi", EndSec: 1}}}}
	lines := []inventory.Segment{{Text: "hi"}}
	if got := captionNote("from the saved transcript", words); got != "burned in word by word, from the saved transcript (1 segments)" {
		t.Errorf("words: %q", got)
	}
	got := captionNote("from the saved transcript", lines)
	if want := "burned in one line per segment, from the saved transcript (1 segments); run `trady source transcribe` again for word by word"; got != want {
		t.Errorf("lines: %q", got)
	}
}
