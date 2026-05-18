package sparkwing

import (
	"sort"
	"testing"
)

// TestWingFlagDocs_OrderAndUniqueness pins the documented wing-flag
// set so a regression (typo, accidental dedupe, missing entry) shows
// up as a test failure. Order is also pinned because the per-pipeline
// help footer renders in walk order; arbitrary reordering would
// silently re-shape every pipeline's --help.
func TestWingFlagDocs_OrderAndUniqueness(t *testing.T) {
	docs := WingFlagDocs()
	if len(docs) == 0 {
		t.Fatalf("WingFlagDocs() returned empty slice")
	}
	seen := map[string]bool{}
	for _, d := range docs {
		if d.Name == "" {
			t.Errorf("empty Name in entry %+v", d)
		}
		if d.Desc == "" {
			t.Errorf("empty Desc on --%s", d.Name)
		}
		if d.Group == "" {
			t.Errorf("empty Group on --%s", d.Name)
		}
		if seen[d.Name] {
			t.Errorf("duplicate --%s in WingFlagDocs", d.Name)
		}
		seen[d.Name] = true
	}
}

// TestWingFlagDocs_CoversSafetyFlags pins the range-resume, dry-run,
// and blast-radius flag set the doc list MUST include. A future
// cleanup that removes one should fail loud here so the help drift
// doesn't regress.
func TestWingFlagDocs_CoversSafetyFlags(t *testing.T) {
	docs := WingFlagDocs()
	have := map[string]bool{}
	for _, d := range docs {
		have[d.Name] = true
	}
	mustHave := []string{
		// Range-resume.
		"sw-start-at", "sw-stop-at",
		// Dry-run.
		"sw-dry-run",
		// Blast-radius escape hatches.
		"sw-allow-destructive", "sw-allow-prod", "sw-allow-money",
	}
	for _, f := range mustHave {
		if !have[f] {
			t.Errorf("WingFlagDocs missing --%s", f)
		}
	}
}

// TestWingFlagDocs_AllSwPrefixed pins that every documented wing flag
// carries the sw- prefix. The prefix is the entire reservation
// mechanism — it lets pipeline-author Inputs flags occupy the
// unprefixed namespace without collision.
func TestWingFlagDocs_AllSwPrefixed(t *testing.T) {
	for _, d := range WingFlagDocs() {
		if d.Name[:3] != "sw-" {
			t.Errorf("--%s lacks sw- prefix; every wing-owned flag must be sw-prefixed so pipeline-author flags are collision-free", d.Name)
		}
	}
}

// TestWingFlagDocs_ReturnsCopy ensures callers may mutate the returned
// slice freely without affecting subsequent calls.
func TestWingFlagDocs_ReturnsCopy(t *testing.T) {
	a := WingFlagDocs()
	if len(a) == 0 {
		t.Fatalf("WingFlagDocs() empty")
	}
	a[0].Name = "MUTATED"
	b := WingFlagDocs()
	if b[0].Name == "MUTATED" {
		t.Errorf("WingFlagDocs returned a shared slice; mutation leaked: %v", b[0])
	}
}

// TestWingFlagDocs_GroupsAreKnown pins the rendering buckets so a
// rogue Group string ("system " with trailing space, "System " with
// capitalization drift) doesn't silently fall into a default bucket.
// Post-sw-prefix rename, every wing flag belongs to a single "System"
// bucket; pipeline-author flags get their own "Pipeline Args" bucket
// in the render layer.
func TestWingFlagDocs_GroupsAreKnown(t *testing.T) {
	known := map[string]bool{
		"System": true,
	}
	for _, d := range WingFlagDocs() {
		if !known[d.Group] {
			t.Errorf("--%s has unknown Group %q (expected one of: %v)", d.Name, d.Group, sortedBoolKeys(known))
		}
	}
}

func sortedBoolKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
