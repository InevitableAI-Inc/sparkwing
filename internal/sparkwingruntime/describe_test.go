package sparkwingruntime_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/sparkwing-dev/sparkwing/internal/sparkwingruntime"
	"github.com/sparkwing-dev/sparkwing/sparkwing"
)

type deployInputs struct {
	Env     string        `flag:"env" required:"true" desc:"Target environment"`
	Version string        `flag:"version" desc:"Image tag"`
	NoApply bool          `flag:"no-apply" desc:"Preview without applying"`
	Count   int           `flag:"count" default:"3" desc:"Number of replicas"`
	Timeout time.Duration `flag:"timeout" desc:"Deadline"`
	//lint:ignore U1000 fixture field verifies the parser skips fields without a flag tag
	unexported string
}

type deployForDescribe struct{ sparkwing.Base }

func (deployForDescribe) Plan(_ context.Context, plan *sparkwing.Plan, _ deployInputs, rc sparkwing.RunContext) error {
	sparkwing.Job(plan, rc.Pipeline, func(ctx context.Context) error { return nil })
	return nil
}

func TestDescribePipelineShape(t *testing.T) {
	sparkwing.Register[deployInputs]("describe-fixture", func() sparkwing.Pipeline[deployInputs] {
		return deployForDescribe{}
	})

	dp, ok, err := sparkwingruntime.DescribePipelineByName("describe-fixture")
	if err != nil {
		t.Fatalf("describe: %v", err)
	}
	if !ok {
		t.Fatal("pipeline should be registered")
	}
	if dp.Name != "describe-fixture" {
		t.Errorf("Name = %q, want %q", dp.Name, "describe-fixture")
	}
	if len(dp.Args) != 5 {
		t.Fatalf("Args count = %d, want 5 (excluding unexported), got: %+v", len(dp.Args), dp.Args)
	}

	byName := map[string]sparkwing.DescribeArg{}
	for _, a := range dp.Args {
		byName[a.Name] = a
	}

	env := byName["env"]
	if env.Type != "string" || !env.Required || env.Desc != "Target environment" {
		t.Errorf("env = %+v", env)
	}
	if env.GoName != "Env" {
		t.Errorf("env.GoName = %q, want Env", env.GoName)
	}
	dry := byName["no-apply"]
	if dry.Type != "bool" || dry.Required {
		t.Errorf("no-apply = %+v", dry)
	}
	count := byName["count"]
	if count.Type != "int" || count.Default != "3" {
		t.Errorf("count = %+v", count)
	}
	to := byName["timeout"]
	if to.Type != "duration" {
		t.Errorf("timeout = %+v", to)
	}

	// Round-trip through JSON so the wire shape is locked down. The
	// sparkwing CLI consumes this exact encoding out of the describe
	// subprocess.
	blob, err := json.Marshal(dp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back sparkwing.DescribePipeline
	if err := json.Unmarshal(blob, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(back.Args) != len(dp.Args) {
		t.Errorf("round-trip args count mismatch")
	}
}
