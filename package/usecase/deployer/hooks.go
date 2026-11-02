package deployer

import (
	"slices"
)

func mergeHooks(cur map[string]any, value any) (bool, error) {
	add, ok := value.(map[string]any)
	if !ok {
		return false, errShape
	}
	hooks, ok := cur["hooks"].(map[string]any)
	if cur["hooks"] != nil && !ok {
		return false, errShape
	}
	if hooks == nil {
		hooks = map[string]any{}
	}
	changed := false
	for _, event := range sortedKeys(add) {
		groups, ok := add[event].([]any)
		if !ok {
			return false, errShape
		}
		have, ok := hooks[event].([]any)
		if hooks[event] != nil && !ok {
			return false, errShape
		}
		for _, g := range groups {
			group, ok := g.(map[string]any)
			if !ok {
				return false, errShape
			}
			if !covers(have, group) {
				have, changed = append(have, g), true
			}
		}
		hooks[event] = have
	}
	if changed {
		cur["hooks"] = hooks
	}
	return changed, nil
}

func covers(groups []any, group map[string]any) bool {
	matcher, _ := group["matcher"].(string)
	want := commands(group)
	for _, g := range groups {
		other, ok := g.(map[string]any)
		if !ok {
			continue
		}
		if m, _ := other["matcher"].(string); m != matcher {
			continue
		}
		have := commands(other)
		missing := slices.ContainsFunc(want, func(c string) bool { return !slices.Contains(have, c) })
		if len(want) > 0 && !missing {
			return true
		}
	}
	return false
}

func commands(group map[string]any) []string {
	var out []string
	handlers, _ := group["hooks"].([]any)
	for _, h := range handlers {
		if handler, ok := h.(map[string]any); ok {
			if c, ok := handler["command"].(string); ok {
				out = append(out, c)
			}
		}
	}
	return out
}
