package orchestrator

import (
	"bytes"
	"strings"
	"testing"
)

// TestPrintWingFlagsSection_ContainsArcFlags pins per-pipeline help
// (`wing <pipeline> --help`, `sparkwing run <pipeline> --help`) so it
// enumerates the wing flags. Pre-fix the footer was a hand-coded
// "(--on, --from, --config)" line that omitted --start-at, --stop-at,
// --dry-run, and the --allow-* set entirely. A future regression that
// drops one of these from sparkwing.WingFlagDocs() fails this test
// loud.
func TestPrintWingFlagsSection_ContainsArcFlags(t *testing.T) {
	var buf bytes.Buffer
	printWingFlagsSection(&buf)
	out := buf.String()

	// range-resume.
	mustContain(t, out, "--sw-start-at")
	mustContain(t, out, "--sw-stop-at")
	// dry-run.
	mustContain(t, out, "--sw-dry-run")
	// blast-radius escape hatches.
	mustContain(t, out, "--sw-allow-destructive")
	mustContain(t, out, "--sw-allow-prod")
	mustContain(t, out, "--sw-allow-money")

	// Older staples must still appear -- the regression we want to
	// avoid is REPLACING the old hand-coded list with an equally stale
	// newer one.
	mustContain(t, out, "--sw-on")
	mustContain(t, out, "--sw-from")
	mustContain(t, out, "--sw-config")
	mustContain(t, out, "--sw-retry-of")

	// The header label keeps the section discoverable.
	mustContain(t, out, "WING FLAGS")
}

// TestPrintWingFlagsSection_GroupsRender pins that every wing flag
// renders under a single [System] group header. The pre-sw-prefix
// rename used Source/Range/Safety/Selection sub-groups; after the
// rename, pipeline-author flag namespace is fully unprefixed and
// wing flags live under one unified [System] label so the operator
// only sees two top-level groupings: pipeline args (unprefixed) and
// sparkwing args (--sw-* under System).
func TestPrintWingFlagsSection_GroupsRender(t *testing.T) {
	var buf bytes.Buffer
	printWingFlagsSection(&buf)
	out := buf.String()
	if !strings.Contains(out, "[System]") {
		t.Errorf("expected group label [System] in output:\n%s", out)
	}
	for _, label := range []string{"[Source]", "[Range]", "[Safety]", "[Selection]"} {
		if strings.Contains(out, label) {
			t.Errorf("did not expect sub-group label %q in output (collapsed under [System]):\n%s", label, out)
		}
	}
}

func mustContain(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Errorf("expected output to contain %q; got:\n%s", needle, haystack)
	}
}
