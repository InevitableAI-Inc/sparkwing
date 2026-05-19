package sparkwing

import "context"

// targetKey scopes the active-target context value installed by
// internal/sparkwingruntime.WithTarget.
type targetKey struct{}

// Target returns the active target for the current run, or "" when no
// target was selected. Single-target pipelines auto-select their lone
// target; multi-target pipelines invoked without --for fall through
// to the empty string.
//
// Most pipeline code should NOT branch on Target directly -- declare
// the topology difference with OnTarget on the relevant Job or
// WorkStep so the scheduler computes the right DAG per target. The
// accessor exists for diagnostics, logging, and the rare case where
// neither OnTarget nor a typed Config field is a clean fit.
//
//	if t := sparkwing.Target(ctx); t == "prod" {
//	    sparkwing.Info(ctx, "running against prod")
//	}
func Target(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(targetKey{}).(string); ok {
		return v
	}
	return ""
}
