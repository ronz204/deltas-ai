package setup

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func decode(t *testing.T, src []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(src, &m); err != nil {
		t.Fatalf("not JSON: %v\n%s", err, src)
	}
	return m
}

func TestMergeAddsWhatDeltaNeedsToAnEmptyFile(t *testing.T) {
	merged, changed, err := mergeSettings([]byte(`{}`), []byte(fragment))
	if err != nil || !changed {
		t.Fatalf("changed = %v, err = %v", changed, err)
	}
	if !reflect.DeepEqual(decode(t, merged), decode(t, []byte(fragment))) {
		t.Fatalf("merged = %s", merged)
	}
}

func TestMergeKeepsWhatTheProjectHasAndAddsOnlyWhatIsMissing(t *testing.T) {
	mine := `{
	  "model": "opus",
	  "permissions": {"allow": ["Bash(make *)"], "deny": ["Bash(rm -rf *)"]},
	  "hooks": {"Stop": [{"hooks": [{"type": "command", "command": "my-stop-hook"}]}], "Notification": [{"hooks": []}]}
	}`
	merged, changed, err := mergeSettings([]byte(mine), []byte(fragment))
	if err != nil || !changed {
		t.Fatalf("changed = %v, err = %v", changed, err)
	}
	got := decode(t, merged)
	if got["model"] != "opus" {
		t.Errorf("a key the project has was lost: %v", got)
	}
	perms := got["permissions"].(map[string]any)
	if !reflect.DeepEqual(perms["allow"], []any{"Bash(make *)"}) || !reflect.DeepEqual(perms["deny"], []any{"Bash(rm -rf *)", "Read(**/.env)"}) {
		t.Errorf("permissions = %v, want the project's rules first and delta's added", perms)
	}
	hooks := got["hooks"].(map[string]any)
	if len(hooks["Stop"].([]any)) != 2 || len(hooks["Notification"].([]any)) != 1 {
		t.Errorf("hooks = %v, want the project's Stop hook kept beside delta's", hooks)
	}
}

func TestMergeIsIdempotent(t *testing.T) {
	once, _, err := mergeSettings([]byte(`{"model":"opus"}`), []byte(fragment))
	if err != nil {
		t.Fatal(err)
	}
	twice, changed, err := mergeSettings(once, []byte(fragment))
	if err != nil || changed || string(twice) != string(once) {
		t.Fatalf("a second merge changed the file: changed = %v, err = %v\n%s", changed, err, twice)
	}
}

func TestADeltaHookThatTheProjectTunedIsNotAddedAgain(t *testing.T) {
	tuned := `{"permissions":{"deny":["Read(**/.env)"]},"hooks":{"Stop":[{"hooks":[{"type":"command","command":"delta hook","timeout":60}]}]}}`
	merged, changed, err := mergeSettings([]byte(tuned), []byte(fragment))
	if err != nil || changed || string(merged) != tuned {
		t.Fatalf("changed = %v, err = %v; a different timeout is the project's choice", changed, err)
	}
}

func TestMergeRefusesWhatItCannotRead(t *testing.T) {
	for name, have := range map[string]string{
		"not json":              `{"hooks": `,
		"not an object":         `[1, 2]`,
		"null":                  `null`,
		"hooks is a list":       `{"hooks": []}`,
		"an event is an object": `{"hooks": {"Stop": {}}}`,
		"permissions is text":   `{"permissions": "all"}`,
	} {
		if _, changed, err := mergeSettings([]byte(have), []byte(fragment)); err == nil || changed {
			t.Errorf("%s: changed = %v, err = %v; want an error and nothing changed", name, changed, err)
		}
	}
}

func TestInitLeavesASettingsFileItCannotMergeAsItIs(t *testing.T) {
	w := newWorld()
	w.files[".claude/settings.json"] = []byte(`{"hooks": `)
	steps, err := installer(w).Run(ctx)
	if err != nil {
		t.Fatal(err)
	}
	last := steps[len(steps)-1]
	if last.Done || !strings.Contains(last.What, "left as it is") || string(w.files[".claude/settings.json"]) != `{"hooks": ` {
		t.Fatalf("step = %+v, file = %q", last, w.files[".claude/settings.json"])
	}
}
