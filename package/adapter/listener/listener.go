package listener

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/ronz204/deltas-ai/package/usecase/sentinel"
)

type event struct {
	Name       string `json:"hook_event_name"`
	Session    string `json:"session_id"`
	StopActive bool   `json:"stop_hook_active"`
	Tool       string `json:"tool_name"`
	ToolInput  struct {
		Path string `json:"file_path"`
	} `json:"tool_input"`
}

func Run(ctx context.Context, g *sentinel.Guard, in io.Reader, out, errOut io.Writer) int {
	var e event
	if err := json.NewDecoder(in).Decode(&e); err != nil {
		fmt.Fprintln(errOut, "delta hook: cannot read the event:", err)
		return 1
	}
	d, err := g.Decide(ctx, sentinel.Event{Name: e.Name, Tool: e.Tool, Path: e.ToolInput.Path, Session: e.Session, StopActive: e.StopActive})
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
