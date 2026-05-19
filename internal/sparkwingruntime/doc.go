// Package sparkwingruntime holds orchestrator-only plumbing that the
// sparkwing package once exposed at the top level (WithDryRun,
// WithRunner, WithSpawnHandler, …). Relocating them here tightens the
// author-facing surface visible in sparkwing's IDE autocomplete: only
// the orchestrator and other runtime hosts import this package, while
// pipeline-author code (in .sparkwing/jobs/*.go of consumer repos) is
// blocked by the `internal/` boundary.
//
// The package adds no new behavior; every symbol here is the same
// function previously declared in github.com/sparkwing-dev/sparkwing/sparkwing,
// moved verbatim.
package sparkwingruntime
