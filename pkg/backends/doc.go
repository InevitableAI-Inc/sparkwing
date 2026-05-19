// Package backends loads backends.yaml -- the file that declares
// where the three persistence surfaces live: cache (content-addressed
// artifacts and compiled pipeline binaries), logs (per-job log
// streams), and state (run records, plan snapshots, status).
//
// # Source precedence (per-field, repo wins)
//
//  1. .sparkwing/backends.yaml          -- team-shared, checked in
//  2. ~/.config/sparkwing/backends.yaml -- per-user additions / overrides
//
// A name in both files merges per non-zero field with repo values
// winning.
//
// # Selection at process start
//
//  1. Per-target overlay (pipelines.yaml targets.<name>.backend)
//  2. Auto-detected environment (first matching environments.<name>.detect)
//  3. [File.Defaults] block
//
// # Loading
//
// Use [Load] for a single file, [Resolve] to apply the repo / user
// precedence, [ResolveWithOverlay] to splice in a third overlay, or
// [Merge] to combine two [File]s directly. [Spec] describes one
// backend (its [Surface], type discriminator, and per-type fields);
// [Surfaces] groups Cache / Logs / State; [Environment] adds
// auto-detection via [Detect].
//
// # Shape (yaml)
//
//	defaults:
//	  cache:
//	    type: filesystem
//	    path: ~/.cache/sparkwing
//	  logs:
//	    type: filesystem
//	    path: ~/.cache/sparkwing/logs
//	  state:
//	    type: sqlite
//	    path: ~/.cache/sparkwing/state.db
//
//	environments:
//	  gha:
//	    detect: { env_var: GITHUB_ACTIONS, equals: "true" }
//	    cache: { type: s3, bucket: sparkwing-cache, prefix: ${GITHUB_REPOSITORY}/ }
//	    logs:  { type: s3, bucket: sparkwing-logs,  prefix: ${GITHUB_REPOSITORY}/ }
//	  kubernetes:
//	    detect: { env_var: KUBERNETES_SERVICE_HOST, present: true }
//	    cache: { type: controller }
//	    logs:  { type: controller }
package backends
