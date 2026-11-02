package deployer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"slices"
)

const settingsPath = ".claude/settings.json"

func (i *Installer) installSettings(ctx context.Context) (Step, error) {
	fragment, err := i.harness.Settings()
	if err != nil {
		return Step{}, err
	}
	have, err := i.files.ReadFile(ctx, settingsPath)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		if err := i.files.WriteFile(ctx, settingsPath, fragment); err != nil {
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
	if err := i.files.WriteFile(ctx, settingsPath, merged); err != nil {
		return Step{}, err
	}
	return Step{true, "added delta's hooks and permissions to " + settingsPath}, nil
}

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

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
