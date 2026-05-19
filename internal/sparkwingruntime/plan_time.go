package sparkwingruntime

import (
	"context"

	"github.com/sparkwing-dev/sparkwing/sparkwing/planguard"
)

// GuardPlanTime panics if invoked from inside a Pipeline.Plan() call.
// `what` names the helper that triggered the guard (e.g.
// "sparkwing.Bash") so the panic message tells the author exactly
// which call to lift into a Job. Custom helpers can guard their own
// ctx-taking entry points by calling this.
func GuardPlanTime(ctx context.Context, what string) {
	planguard.Guard(ctx, what)
}

// IsPlanTime reports whether ctx is currently inside a Plan() call.
// Mostly for tests; production code should prefer GuardPlanTime.
func IsPlanTime(ctx context.Context) bool {
	return planguard.Active(ctx)
}
