package sparkwing

import "testing"

func TestParseDebug(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"", false},
		{"0", false},
		{"false", false},
		{"False", false},
		{"FALSE", false},
		{"1", true},
		{"true", true},
		{"yes", true},
	}
	for _, tc := range cases {
		t.Run("v="+tc.in, func(t *testing.T) {
			if got := parseDebug(tc.in); got != tc.want {
				t.Errorf("parseDebug(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestDebugEnabled_setDebugRoundTrip(t *testing.T) {
	prev := DebugEnabled()
	t.Cleanup(func() { setDebug(prev) })

	setDebug(true)
	if !DebugEnabled() {
		t.Error("setDebug(true) didn't enable")
	}
	setDebug(false)
	if DebugEnabled() {
		t.Error("setDebug(false) didn't disable")
	}
}
