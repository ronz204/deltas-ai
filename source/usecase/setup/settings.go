package setup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"slices"
)

const settingsPath = ".claude/settings.json"

// installSettings adds delta's hooks and permissions to .claude/settings.json and leaves the rest of the
// file as the project has it.
func (i *Installer) installSettings(ctx context.Context) (Step, error) {
	fragment, err := i.harness.Settings()
	if err != nil {
		return Step{}, err
	}
	have, err := i.files.ReadFile(ctx, settingsPath)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		if err := i.files.WriteFile(ctx, settingsPath, fragment, 0o644); err != nil {
			return Step{}, err
		}
		return Step{true, "installed " + settingsPath}, nil
	case err != nil:
		return Step{}, err
	}
	merged, changed, err := mergeSettings(have, fragment)
	switch {
	case err != nil:
		return Step{false, fmt.Sprintf("%s could not be merged (%v): left as it is", settingsPath, err)}, nil
	case !changed:
		return Step{false, settingsPath + " already has delta's hooks and permissions"}, nil
	}
	if err := i.files.WriteFile(ctx, settingsPath, merged, 0o644); err != nil {
		return Step{}, err
	}
	return Step{true, "added delta's hooks and permissions to " + settingsPath}, nil
}

// mergeSettings adds to have what fragment brings and have lacks: hook groups that do not already run the
// same commands for the same matcher, permission rules, and top-level keys. Nothing is removed or changed.
// The merged file is written with sorted keys. changed is false, and have comes back as it was, when there
// was nothing to add.
func mergeSettings(have, fragment []byte) (merged []byte, changed bool, err error) {
	var cur, add map[string]any
	if err := json.Unmarshal(have, &cur); err != nil || cur == nil {
		return have, false, errors.New("it is not a JSON object")
	}
	if err := json.Unmarshal(fragment, &add); err != nil {
		return have, false, err
	}
	for key, value := range add {
		var did bool
		switch key {
		case "hooks":
			did, err = mergeHooks(cur, value)
		case "permissions":
			did, err = mergePermissions(cur, value)
		default:
			_, exists := cur[key]
			if !exists {
				cur[key], did = value, true
			}
		}
		if err != nil {
			return have, false, err
		}
		changed = changed || did
	}
	if !changed {
		return have, false, nil
	}
	out, err := json.MarshalIndent(cur, "", "  ")
	return append(out, '\n'), true, err
}

var errShape = errors.New("a hooks or permissions entry is not shaped as Claude Code expects")

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

// covers says whether some group already has the matcher of group and runs all its commands.
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

func mergePermissions(cur map[string]any, value any) (bool, error) {
	add, ok := value.(map[string]any)
	if !ok {
		return false, errShape
	}
	perms, ok := cur["permissions"].(map[string]any)
	if cur["permissions"] != nil && !ok {
		return false, errShape
	}
	if perms == nil {
		perms = map[string]any{}
	}
	changed := false
	for _, kind := range sortedKeys(add) {
		rules, ok := add[kind].([]any)
		if !ok {
			return false, errShape
		}
		have, ok := perms[kind].([]any)
		if perms[kind] != nil && !ok {
			return false, errShape
		}
		for _, r := range rules {
			if _, ok := r.(string); !ok {
				return false, errShape
			}
			if !slices.Contains(have, r) {
				have, changed = append(have, r), true
			}
		}
		perms[kind] = have
	}
	if changed {
		cur["permissions"] = perms
	}
	return changed, nil
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
