package sparkwingruntime

import "sort"

// SortedUniqueRisks returns the deduplicated, lexicographically sorted
// union of the provided label slices. Used by validators to render
// stable error messages and by describe to emit a stable wire shape.
func SortedUniqueRisks(slices ...[]string) []string {
	seen := map[string]bool{}
	for _, sl := range slices {
		for _, l := range sl {
			if l == "" || seen[l] {
				continue
			}
			seen[l] = true
		}
	}
	out := make([]string, 0, len(seen))
	for l := range seen {
		out = append(out, l)
	}
	sort.Strings(out)
	return out
}
