package sparkwing_test

import (
	"context"
	"testing"

	"github.com/sparkwing-dev/sparkwing/sparkwing"
)

func TestRunner_NilWithoutInstall(t *testing.T) {
	if got := sparkwing.Runner(context.Background()); got != nil {
		t.Errorf("Runner(ctx) without WithRunner = %+v, want nil", got)
	}
}

func TestRunner_NilReceiverHasLabelSafe(t *testing.T) {
	var r *sparkwing.RunnerInfo
	if r.HasLabel("anything") {
		t.Error("HasLabel on nil receiver returned true")
	}
}

func TestWithRunner_RoundTrip(t *testing.T) {
	want := &sparkwing.RunnerInfo{
		Name:   "cloud-linux",
		Type:   "kubernetes",
		Labels: []string{"cloud-linux", "os=linux", "arch=amd64"},
	}
	ctx := sparkwing.WithRunner(context.Background(), want)
	got := sparkwing.Runner(ctx)
	if got != want {
		t.Errorf("Runner(ctx) returned different pointer: %p vs %p", got, want)
	}
}

func TestRunnerInfo_HasLabel_BareAndCommaOR(t *testing.T) {
	r := &sparkwing.RunnerInfo{Labels: []string{"local", "os=darwin", "arch=arm64"}}
	cases := []struct {
		term string
		want bool
	}{
		{"local", true},
		{"os=darwin", true},
		{"os=linux", false},
		{"os=linux,os=darwin", true},
		{"os=linux,os=macos", false},
		{"arch=arm64", true},
		{"", false},
	}
	for _, tc := range cases {
		t.Run(tc.term, func(t *testing.T) {
			if got := r.HasLabel(tc.term); got != tc.want {
				t.Errorf("HasLabel(%q) = %v, want %v", tc.term, got, tc.want)
			}
		})
	}
}

func TestWithRunner_NilInfoStillReadsAsNil(t *testing.T) {
	ctx := sparkwing.WithRunner(context.Background(), nil)
	if got := sparkwing.Runner(ctx); got != nil {
		t.Errorf("Runner(ctx) after WithRunner(nil) = %+v, want nil", got)
	}
}
