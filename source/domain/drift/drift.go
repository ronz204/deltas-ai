// Package drift decides when a contract is out of date: the files a slice depends on are not the ones it was verified against.
package drift

import "slices"

// Changed lists the files that differ between two fingerprints (path to content hash): added, removed or
// edited. It is sorted, and empty when the two are the same.
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
