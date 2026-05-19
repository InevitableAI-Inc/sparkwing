package runner

import "github.com/sparkwing-dev/sparkwing/internal/orchestrator"

// Main is the entry point for a user repo's compiled pipeline binary.
// .sparkwing/main.go calls runner.Main() after blank-importing the
// user's jobs package; runner.Main then dispatches into the
// orchestrator.
func Main() { orchestrator.Main() }
