package hook

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ronz204/deltas-ai/package/usecase/sentinel"
	"github.com/ronz204/deltas-ai/package/usecase/tracking"
)

type specs map[string]string

func (s specs) Path(slice string) string { return "deltas/specs/" + slice + ".spec.md" }
func (s specs) List(context.Context) ([]string, error) {
	if len(s) == 0 {
		return nil, nil
	}
	return []string{"billing"}, nil
}
func (s specs) Load(_ context.Context, slice string) ([]byte, error) {
	if src, ok := s[slice]; ok {
		return []byte(src), nil
	}
	return nil, errors.New("missing")
}

func run(t *testing.T, s specs, input string) (code int, out, errOut string) {
	t.Helper()
	root := t.TempDir()
	input = strings.ReplaceAll(input, "{root}", strings.ReplaceAll(filepath.ToSlash(root), `\`, `\\`))
	var o, e bytes.Buffer
	code = Run(context.Background(), sentinel.New(root, tracking.New(s), nil, nil, nil), strings.NewReader(input), &o, &e)
	return code, o.String(), e.String()
}

func TestBlockExitsTwoWithTheReasonOnStderr(t *testing.T) {
	code, out, errOut := run(t, nil, `{"hook_event_name":"PreToolUse","tool_name":"Write","tool_input":{"file_path":"{root}/deltas/specs/a.spec.md","content":"x"}}`)
	if code != 2 || out != "" || !strings.Contains(errOut, "deltas/specs/a.spec.md") {
		t.Fatalf("code = %d, stdout = %q, stderr = %q", code, out, errOut)
	}
}

func TestAllowedEventsExitZeroAndSayNothing(t *testing.T) {
	for _, input := range []string{
		`{"hook_event_name":"PreToolUse","tool_name":"Write","tool_input":{"file_path":"{root}/main.go"}}`,
		`{"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"command":"go test ./..."}}`,
		`{"hook_event_name":"PostToolUse","tool_name":"Write","tool_input":{"file_path":"{root}/deltas/specs/a.spec.md"},"tool_response":{"type":"create"}}`,
		`{"hook_event_name":"Stop","stop_hook_active":false,"last_assistant_message":"done"}`,
		`{"hook_event_name":"SessionStart","source":"startup"}`,
		`{}`,
	} {
		if code, out, errOut := run(t, nil, input); code != 0 || out != "" || errOut != "" {
			t.Errorf("%s = %d %q %q, want 0 and silence", input, code, out, errOut)
		}
	}
}

func TestSessionStartPrintsTheSlicesToStdout(t *testing.T) {
	s := specs{"billing": "---\nformat: 1\nslice: billing\nstatus: approved\n---\n- INV-01 [none]\n  When x, y.\n"}
	code, out, errOut := run(t, s, `{"hook_event_name":"SessionStart","source":"startup","cwd":"/x"}`)
	if code != 0 || errOut != "" || !strings.Contains(out, "- billing: approved, 1 claims") {
		t.Fatalf("code = %d, stdout = %q, stderr = %q", code, out, errOut)
	}
}

// A failure of delta itself must never look like a block: Claude Code blocks only on exit 2.
func TestBrokenInputIsExitOneNotTwo(t *testing.T) {
	for _, input := range []string{"", "not json", `{"hook_event_name":`, `[1,2]`} {
		code, _, errOut := run(t, nil, input)
		if code != 1 || !strings.HasPrefix(errOut, "delta hook: ") {
			t.Errorf("%q = %d %q, want exit 1 with a delta hook: message", input, code, errOut)
		}
	}
}

func TestAFailingSliceReportIsExitOneNotTwo(t *testing.T) {
	broken := failingSpecs{}
	var o, e bytes.Buffer
	code := Run(context.Background(), sentinel.New(t.TempDir(), tracking.New(broken), nil, nil, nil), strings.NewReader(`{"hook_event_name":"SessionStart"}`), &o, &e)
	if code != 1 || !strings.Contains(e.String(), "disk on fire") {
		t.Fatalf("code = %d, stderr = %q", code, e.String())
	}
}

type failingSpecs struct{}

func (failingSpecs) Path(string) string { return "" }
func (failingSpecs) List(context.Context) ([]string, error) {
	return nil, errors.New("disk on fire")
}
func (failingSpecs) Load(context.Context, string) ([]byte, error) { return nil, nil }
