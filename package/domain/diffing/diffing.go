package diffing

import "slices"

func Changed(before, after map[string]string) []string {
	var out []string
	for name, hash := range after {
		if old, ok := before[name]; !ok || old != hash {
			out = append(out, name)
		}
	}
	for name := range before {
		if _, ok := after[name]; !ok {
			out = append(out, name)
		}
	}
	slices.Sort(out)
	return out
}
