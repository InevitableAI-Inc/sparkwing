package sparkwing_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/sparkwing-dev/sparkwing/internal/sparkwingruntime"
	"github.com/sparkwing-dev/sparkwing/sparkwing"
)

func TestTarget_AccessorRoundTrip(t *testing.T) {
	if got := sparkwing.Target(context.Background()); got != "" {
		t.Fatalf("default Target should be empty, got %q", got)
	}
	ctx := sparkwingruntime.WithTarget(context.Background(), "prod")
	if got := sparkwing.Target(ctx); got != "prod" {
		t.Fatalf("Target after WithTarget: got %q want %q", got, "prod")
	}
}

func TestOnTarget_PopulatesList(t *testing.T) {
	plan := sparkwing.NewPlan()
	n := sparkwing.Job(plan, "deploy", &buildJob{}).OnTarget("prod", "staging")
	got := n.OnTargets()
	want := []string{"prod", "staging"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("OnTargets = %v, want %v", got, want)
	}
}

func TestOnTarget_EmptyClears(t *testing.T) {
	plan := sparkwing.NewPlan()
	n := sparkwing.Job(plan, "deploy", &buildJob{}).OnTarget("prod").OnTarget()
	if got := n.OnTargets(); got != nil {
		t.Fatalf("OnTargets after clear: %v", got)
	}
}

func TestOnTarget_GroupDelegates(t *testing.T) {
	plan := sparkwing.NewPlan()
	a := sparkwing.Job(plan, "a", &buildJob{})
	b := sparkwing.Job(plan, "b", &buildJob{})
	g := sparkwing.GroupJobs(plan, "deploys", a, b).OnTarget("dev")
	for _, m := range g.Members() {
		if got := m.OnTargets(); !reflect.DeepEqual(got, []string{"dev"}) {
			t.Fatalf("member %q OnTargets = %v", m.ID(), got)
		}
	}
}

func TestEffectiveJobTargets_Explicit(t *testing.T) {
	plan := sparkwing.NewPlan()
	a := sparkwing.Job(plan, "a", &buildJob{}).OnTarget("prod")
	_ = a
	got := sparkwingruntime.EffectiveJobTargets(plan)
	want := []string{"prod"}
	if !reflect.DeepEqual(got["a"], want) {
		t.Fatalf("effective[a] = %v, want %v", got["a"], want)
	}
}

func TestEffectiveJobTargets_InheritsFromConsumer(t *testing.T) {
	plan := sparkwing.NewPlan()
	build := sparkwing.Job(plan, "build", &buildJob{})
	sparkwing.Job(plan, "deploy", &buildJob{}).OnTarget("prod").Needs(build)
	got := sparkwingruntime.EffectiveJobTargets(plan)
	if !reflect.DeepEqual(got["build"], []string{"prod"}) {
		t.Fatalf("build should inherit from deploy: got %v", got["build"])
	}
	if !reflect.DeepEqual(got["deploy"], []string{"prod"}) {
		t.Fatalf("deploy explicit: got %v", got["deploy"])
	}
}

func TestEffectiveJobTargets_UniversalConsumerWins(t *testing.T) {
	plan := sparkwing.NewPlan()
	build := sparkwing.Job(plan, "build", &buildJob{})
	sparkwing.Job(plan, "deploy", &buildJob{}).OnTarget("prod").Needs(build)
	sparkwing.Job(plan, "publish", &buildJob{}).Needs(build)
	got := sparkwingruntime.EffectiveJobTargets(plan)
	if got["build"] != nil {
		t.Fatalf("build should be universal (nil), got %v", got["build"])
	}
}

func TestJobAllowsTarget(t *testing.T) {
	cases := []struct {
		eff    []string
		target string
		allow  bool
	}{
		{nil, "", true},
		{nil, "prod", true},
		{[]string{"prod"}, "prod", true},
		{[]string{"prod"}, "dev", false},
		{[]string{"prod"}, "", false},
		{[]string{"prod", "staging"}, "staging", true},
	}
	for _, c := range cases {
		got := sparkwingruntime.JobAllowsTarget(c.eff, c.target)
		if got != c.allow {
			t.Errorf("JobAllowsTarget(%v, %q) = %v, want %v", c.eff, c.target, got, c.allow)
		}
	}
}

func TestEffectiveStepTargets_InheritsFromStepConsumer(t *testing.T) {
	w := sparkwing.NewWork()
	fetch := sparkwing.Step(w, "fetch", func(context.Context) error { return nil })
	sparkwing.Step(w, "deploy", func(context.Context) error { return nil }).OnTarget("prod").Needs(fetch)
	got := sparkwingruntime.EffectiveStepTargets(w)
	if !reflect.DeepEqual(got["fetch"], []string{"prod"}) {
		t.Fatalf("fetch effective = %v, want [prod]", got["fetch"])
	}
}

func TestWorkStep_OnTarget_EmptyClears(t *testing.T) {
	w := sparkwing.NewWork()
	s := sparkwing.Step(w, "x", func(context.Context) error { return nil }).OnTarget("prod").OnTarget()
	if got := s.OnTargetList(); got != nil {
		t.Fatalf("after clear: %v", got)
	}
}
