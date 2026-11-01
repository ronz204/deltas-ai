// Package hook speaks Claude Code's hook protocol: one JSON event on stdin, an exit status back.
package hook

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/ronz204/deltas-ai/source/usecase/guard"
)

type event struct {
	Name       string `json:"hook_event_name"`
	StopActive bool   `json:"stop_hook_active"`
	Tool       string `json:"tool_name"`
	ToolInput  struct {
		Path    string `json:"file_path"`
		Command string `json:"command"`
	} `json:"tool_input"`
}

// Run answers one event. It returns 0 to let the action happen, 2 to block it with the reason on
// errOut, and 1 when delta itself failed: Claude Code does not block on 1, so a broken delta
// never stops the agent.
func Run(ctx context.Context, g *guard.Guard, in io.Reader, out, errOut io.Writer) int {
	var e event
	if err := json.NewDecoder(in).Decode(&e); err != nil {
		fmt.Fprintln(errOut, "delta hook: cannot read the event:", err)
		return 1
	}
	d, err := g.Decide(ctx, guard.Event{Name: e.Name, Tool: e.Tool, Path: e.ToolInput.Path, Command: e.ToolInput.Command, StopActive: e.StopActive})
	switch {
	case err != nil:
		fmt.Fprintln(errOut, "delta hook:", err)
		return 1
	case d.Block:
		fmt.Fprintln(errOut, d.Reason)
		return 2
	}
	if d.Context != "" {
		fmt.Fprintln(out, d.Context)
	}
	return 0
}
