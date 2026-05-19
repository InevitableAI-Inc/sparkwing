package sparkwing

// runtimePlumbingKeys bundles the context keys that internal/sparkwingruntime
// needs in order to install and read the orchestrator-facing values
// (dry-run flag, runner info, target, step range, spawn handler, ref
// resolvers). Holding the keys in one struct keeps the public surface
// of this package small: a pipeline author sees a single
// `RuntimePlumbing` entry in autocomplete rather than seven.
type runtimePlumbingKeys struct {
	DryRun           any
	Runner           any
	SpawnHandler     any
	StepRange        any
	Target           any
	RefResolver      any
	JSONRefResolver  any
	PipelineResolver any
	PipelineAwaiter  any
}

// RuntimePlumbing exposes context keys to internal/sparkwingruntime so
// the runtime package can install and read these values without a
// circular import.
//
// Pipeline authors should NOT reach for it. The supported surface is
// the typed accessors: IsDryRun, Runner, Target, Ref[T].Get, and the
// SpawnHandler / WorkStep methods.
var RuntimePlumbing = runtimePlumbingKeys{
	DryRun:           dryRunKey{},
	Runner:           runnerCtxKey{},
	SpawnHandler:     keySpawnHandler,
	StepRange:        stepRangeKey{},
	Target:           targetKey{},
	RefResolver:      keyRefResolver,
	JSONRefResolver:  keyJSONRefResolver,
	PipelineResolver: keyPipelineResolver,
	PipelineAwaiter:  keyPipelineAwaiter,
}
