// Package store is the persisted run / node / event data model
// shared between the orchestrator engine, the controller HTTP
// surface, and dashboard readers. Stability promise: this is the
// public data model for sparkwing pipelines; types here are part of
// the SDK surface and version under module SemVer (see VERSIONING.md).
//
// # Opening a store
//
// [Open] returns a [*Store] backed by a SQLite database with WAL
// journaling. The store serializes writes via a single open
// connection; callers can hold one *Store for the process lifetime
// and share it across goroutines.
//
// # Primary records
//
// [Run] is the per-pipeline-invocation row (status, trigger, git
// snapshot, retry / replay lineage, invocation map). Nodes and
// events hang off it via the methods on *Store; concurrency
// admission lives in [ConcurrencyState], [ConcurrencyHolder], and
// [ConcurrencyWaiter]. [Secret] persists named secret material.
// [Session] / [User] back the dashboard's auth surface.
package store
