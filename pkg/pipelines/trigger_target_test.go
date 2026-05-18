package pipelines_test

import (
	"strings"
	"testing"

	"github.com/sparkwing-dev/sparkwing/pkg/pipelines"
)

func TestParse_PushTriggerTargetParsed(t *testing.T) {
	yaml := `
pipelines:
  - name: release
    entrypoint: Release
    targets:
      prod: {}
      staging: {}
    on:
      push:
        branches: [main]
        target: prod
`
	cfg, err := pipelines.Parse(strings.NewReader(yaml))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	p := cfg.Find("release")
	if p == nil {
		t.Fatal("pipeline not found")
	}
	if got := p.On.Push.Target; got != "prod" {
		t.Fatalf("push.target = %q, want prod", got)
	}
	if got := p.TriggerTarget("push"); got != "prod" {
		t.Fatalf("TriggerTarget(push) = %q", got)
	}
}

func TestParse_RejectsTriggerTargetUndeclared(t *testing.T) {
	yaml := `
pipelines:
  - name: release
    entrypoint: Release
    targets:
      prod: {}
    on:
      push:
        branches: [main]
        target: staging
`
	_, err := pipelines.Parse(strings.NewReader(yaml))
	if err == nil || !strings.Contains(err.Error(), "not a declared target") {
		t.Fatalf("expected undeclared-target error, got %v", err)
	}
}

func TestParse_RejectsTriggerTargetOnNoTargetsPipeline(t *testing.T) {
	yaml := `
pipelines:
  - name: lint
    entrypoint: Lint
    on:
      push:
        branches: [main]
        target: prod
`
	_, err := pipelines.Parse(strings.NewReader(yaml))
	if err == nil || !strings.Contains(err.Error(), "pipeline declares no targets") {
		t.Fatalf("expected no-targets error, got %v", err)
	}
}
