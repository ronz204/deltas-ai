package bundle

import (
	"encoding/json"
	"testing"

	"github.com/ronz204/deltas-ai/source/harness"
)

type settings struct {
	Permissions map[string][]string `json:"permissions"`
	Hooks       map[string][]struct {
		Matcher string `json:"matcher"`
		Hooks   []struct {
			Type    string `json:"type"`
			Command string `json:"command"`
			Timeout int    `json:"timeout"`
		} `json:"hooks"`
	} `json:"hooks"`
}

// The shape follows Claude Code's documentation: event, then matcher groups, then handlers.
func TestSettingsHooksAllCallDeltaHook(t *testing.T) {
	raw, err := harness.FS.ReadFile("core/configs/settings.json")
	if err != nil {
		t.Fatal(err)
	}
	var s settings
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Fatalf("settings.json is not valid JSON: %v", err)
	}
	if len(s.Permissions["deny"]) == 0 {
		t.Error("the permission rules were lost")
	}
	matchers := map[string][]string{}
	for event, groups := range s.Hooks {
		for _, g := range groups {
			matchers[event] = append(matchers[event], g.Matcher)
			if len(g.Hooks) == 0 {
				t.Errorf("%s group %q has no handler", event, g.Matcher)
			}
			for _, h := range g.Hooks {
				if h.Type != "command" || h.Command != "delta hook" || h.Timeout <= 0 {
					t.Errorf("%s handler = %+v, want a command running delta hook with a timeout", event, h)
				}
			}
		}
	}
	if len(matchers["SessionStart"]) != 1 {
		t.Errorf("SessionStart groups = %v, want one", matchers["SessionStart"])
	}
	if len(matchers["Stop"]) != 1 {
		t.Errorf("Stop groups = %v, want one", matchers["Stop"])
	}
	if got := matchers["PreToolUse"]; len(got) != 2 || got[0] != "Write|Edit" || got[1] != "Bash" {
		t.Errorf("PreToolUse matchers = %q, want Write|Edit and Bash", got)
	}
}
