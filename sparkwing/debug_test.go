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

func TestDebugEnabled_SetDebugRoundTrip(t *testing.T) {
	prev := DebugEnabled()
	t.Cleanup(func() { SetDebug(prev) })

	SetDebug(true)
	if !DebugEnabled() {
		t.Error("SetDebug(true) didn't enable")
	}
	SetDebug(false)
	if DebugEnabled() {
		t.Error("SetDebug(false) didn't disable")
	}
}
