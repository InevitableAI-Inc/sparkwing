# Changelog

## Unreleased

### Changed

- `WorkStep.Destructive()` / `.AffectsProduction()` / `.CostsMoney()` replaced
  by `.Risk("destructive")` / `.Risk("prod")` / `.Risk("money")`; labels are
  now author-defined (any kebab-case string works, e.g. `.Risk("rotates-key")`).
  Consumer repos using the old methods must update.
- `--sw-allow-destructive` / `--sw-allow-prod` / `--sw-allow-money` collapsed
  into one `--sw-allow LABEL[,LABEL...]` flag (repeatable; comma-separated
  allowed). Profile `auto_allow` is now a list of labels
  (`auto_allow: [destructive]`) instead of per-marker booleans.
- Renamed `--sw-change-directory` to `--sw-cd`. The `-C` short form is unchanged.
- Renamed `--sw-for` to `--sw-target`. The `Job.OnTarget("...")` author API is unchanged.
- Renamed `--sw-on` to `--sw-profile` (and its argument `NAME` to `PROFILE`).
- Renamed `--sw-from` to `--sw-ref` (and the env-var bridge `SPARKWING_FROM` to `SPARKWING_REF`).
- Tightened `--sw-*` flag descriptions in `--help`. No behavior change.
- Tightened `--sw-dry-run` description (no behavior change).
- Moved 15 orchestrator-only plumbing functions out of the sparkwing package
  into `internal/sparkwingruntime`. Pipeline authors never call these;
  relocation tightens the author-facing surface visible in IDE autocomplete
  and godoc. No behavior change.
- Moved Plan-layer plumbing (`GuardPlanTime`, `IsPlanTime`, `ValidateStepRange`,
  `SuggestClosest`, `PreviewPlan`) from `sparkwing` to
  `internal/sparkwingruntime`. Pipeline authors do not call these. No behavior
  change.
- Renamed `JobNode.OnTargetList()` to `JobNode.OnTargets()`. The setter
  `OnTarget(...)` is unchanged.

### Removed

- `JobNode.OnFailureNodeID()`. Use `OnFailureNode().ID()` with a nil check.
- Retired `--sw-retry-of` and `--sw-full`; use `sparkwing runs retry RUN_ID [--failed | --all]`.
- Retired `--sw-job` and `--sw-prefer`; runner selection is now exclusively Plan-layer via `Job.Requires` / `Job.Prefers`. If you used these flags, declare the constraint in the pipeline instead.
- Retired `--sw-backends-env`. `backends.yaml` environment selection is now exclusively auto-detect — if it picks wrong, fix the `match:` rules in `backends.yaml` or the `DetectEnvironment` logic.
