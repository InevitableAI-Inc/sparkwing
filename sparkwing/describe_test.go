package sparkwing_test

import (
	"context"
	"testing"
	"time"

	"github.com/sparkwing-dev/sparkwing/sparkwing"
)

type populateInputs struct {
	Env       string            `flag:"env" required:"true" desc:"Target environment"`
	Version   string            `flag:"version" desc:"Image tag"`
	NoApply   bool              `flag:"no-apply"`
	Count     int               `flag:"count" default:"3"`
	Timeout   time.Duration     `flag:"timeout"`
	BagFields map[string]string `flag:",extra"`
}

type populatePipe struct {
	sparkwing.Base
	captured populateInputs
}

func (pp *populatePipe) Plan(_ context.Context, plan *sparkwing.Plan, in populateInputs, rc sparkwing.RunContext) error {
	pp.captured = in
	sparkwing.Job(plan, rc.Pipeline, func(ctx context.Context) error { return nil })
	return nil
}

func TestRegistration_InvokeParsesTypes(t *testing.T) {
	captured := &populatePipe{}
	sparkwing.Register[populateInputs]("populate-fixture", func() sparkwing.Pipeline[populateInputs] {
		return captured
	})
	reg, ok := sparkwing.Lookup("populate-fixture")
	if !ok {
		t.Fatal("not registered")
	}
	_, err := reg.Invoke(context.Background(), map[string]string{
		"env":      "prod",
		"no-apply": "true",
		"count":    "7",
		"timeout":  "1m30s",
		"version":  "v1.2.3",
		"unknown":  "stashed-in-bag",
	}, sparkwing.RunContext{Pipeline: "populate-fixture"})
	if err != nil {
		t.Fatalf("Invoke: %v", err)
	}
	if captured.captured.Env != "prod" {
		t.Errorf("Env = %q", captured.captured.Env)
	}
	if !captured.captured.NoApply {
		t.Error("NoApply false")
	}
	if captured.captured.Count != 7 {
		t.Errorf("Count = %d", captured.captured.Count)
	}
	if captured.captured.Timeout != 90*time.Second {
		t.Errorf("Timeout = %v", captured.captured.Timeout)
	}
	if captured.captured.Version != "v1.2.3" {
		t.Errorf("Version = %q", captured.captured.Version)
	}
	if captured.captured.BagFields["unknown"] != "stashed-in-bag" {
		t.Errorf("bag = %v", captured.captured.BagFields)
	}
}

func TestRegistration_InvokeBadInt(t *testing.T) {
	sparkwing.Register[populateInputs]("populate-bad-int", func() sparkwing.Pipeline[populateInputs] {
		return &populatePipe{}
	})
	reg, _ := sparkwing.Lookup("populate-bad-int")
	_, err := reg.Invoke(context.Background(), map[string]string{
		"env":   "prod",
		"count": "not-a-number",
	}, sparkwing.RunContext{Pipeline: "populate-bad-int"})
	if err == nil {
		t.Fatal("expected error on bad int")
	}
}
