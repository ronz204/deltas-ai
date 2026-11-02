package deployer

import (
	"slices"
)

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
