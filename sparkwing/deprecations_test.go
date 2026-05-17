package sparkwing

import (
	"os"
	"strings"
	"testing"
)

// TestDeprecationMarkers_PresentOnPublicSurfaces guards the godoc
// "Deprecated:" prefix on the public symbols entering the
// deprecation cycle in step 11. gopls and tools that honor SA1019
// rely on the marker; a rename / refactor that drops the prefix
// would silently break editor warnings without this test failing.
func TestDeprecationMarkers_PresentOnPublicSurfaces(t *testing.T) {
	cases := []struct {
		path   string
		symbol string
		hint   string
	}{
		{"context.go", "TriggerInfo", "Env map[string]string"},
		{"context.go", "TriggerEnv", "func (r RunContext) TriggerEnv"},
		{"runtime_config.go", "RunConfig", "type RunConfig = RuntimeConfig"},
		{"runtime_config.go", "CurrentRunConfig", "func CurrentRunConfig()"},
	}
	for _, tc := range cases {
		t.Run(tc.symbol, func(t *testing.T) {
			body, err := os.ReadFile(tc.path)
			if err != nil {
				t.Fatalf("read %s: %v", tc.path, err)
			}
			s := string(body)
			idx := strings.Index(s, tc.hint)
			if idx == -1 {
				t.Fatalf("symbol anchor %q not found in %s", tc.hint, tc.path)
			}
			window := s[:idx]
			lastDeprecated := strings.LastIndex(window, "// Deprecated:")
			lastBlankLine := strings.LastIndex(window, "\n\n")
			if lastDeprecated == -1 || lastDeprecated < lastBlankLine {
				t.Errorf("%s: missing `// Deprecated:` line immediately above %q", tc.path, tc.hint)
			}
		})
	}
}
