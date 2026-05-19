package sparkwing

import "github.com/sparkwing-dev/sparkwing/pkg/storage"

// Cache is the artifact store interface the orchestrator and pipeline
// authors reach for. It is content-addressed and also holds compiled
// pipeline binaries under bin/<hash>. Backend selection (filesystem,
// S3, sparkwing-cache, ...) is declared in .sparkwing/backends.yaml
// under the cache: surface; every implementation in pkg/storage/*
// satisfies this interface.
type Cache = storage.ArtifactStore

// Logs is the per-job log stream store the orchestrator and pipeline
// authors reach for. Implementations buffer log bytes keyed by
// (runID, nodeID). Backend selection is declared in
// .sparkwing/backends.yaml under the logs: surface; every
// implementation in pkg/storage/* satisfies this interface.
type Logs = storage.LogStore

// State is the run-record store: persists runs, nodes, steps,
// annotations, approvals, and the schema migrations the orchestrator
// depends on. Backend selection is declared in
// .sparkwing/backends.yaml under the state: surface; every
// implementation in pkg/storage/* satisfies this interface.
//
// Implementations today: sqlite. Recognized but not implemented in
// this build: postgres, mysql, controller.
type State = storage.StateStore
