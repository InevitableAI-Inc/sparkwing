package sparkwing

import (
	"os"
	"path/filepath"
	"sync"
)

// RuntimeConfig is the snapshot of "what is true about this process
// at the moment it started." Populated once at package init by
// walking up from cwd to find the project root; stable for the
// lifetime of the run.
//
// Only the WorkDir + Git fields remain. Earlier shapes carried
// IsLocal, RunID, NodeID, and a Debug flag derived from environment
// variables; those have moved out:
//
//   - "Am I local?" is per-job and reachable via Runner(ctx). The
//     orchestrator installs RunnerInfo on the ctx the step body
//     receives; adapters branch on r.HasLabel("local") or r.Type.
//   - RunID and NodeID belong to RunContext and per-job context;
//     read them from rc.RunID / NodeFromContext(ctx) instead of a
//     package-global.
//   - Debug() is a free function (debug.go) reading SPARKWING_DEBUG
//     once at package init.
type RuntimeConfig struct {
	// WorkDir is the directory the pipeline should treat as the
	// repo root. Discovered at process init by walking up from cwd
	// looking for a `.sparkwing/` subdir. Empty when no project
	// was found above cwd; helpers (Path, ReadFile, ...) then
	// refuse to run with a clear error.
	WorkDir string

	// Git describes the source state being built. Same instance
	// as RunContext.Git. Always non-nil so live methods are safe
	// to call from init time; data fields stay empty until SetGit
	// fills them.
	Git *Git
}

var (
	runtimeMu sync.RWMutex
	runtime   = detectRuntime()
)

func detectRuntime() RuntimeConfig {
	rc := RuntimeConfig{}
	if cwd, err := os.Getwd(); err == nil {
		// Git-style auto-discovery: walk up from cwd looking for a
		// `.sparkwing/` subdir. Empty result means no project here;
		// helpers that need WorkDir fail loudly rather than fall
		// back to cwd.
		rc.WorkDir = walkUpToProject(cwd)
	}
	// Pre-populate Git so callers can do `runtime.Git.IsDirty(ctx)`
	// from init time without nil-checking.
	rc.Git = &Git{workDir: rc.WorkDir}
	return rc
}

// walkUpToProject ascends from start looking for a directory that
// contains a `.sparkwing/` child. Returns that directory (the repo
// root) on success, or "" on failure.
func walkUpToProject(start string) string {
	dir := start
	for {
		marker := filepath.Join(dir, ".sparkwing")
		if info, err := os.Stat(marker); err == nil && info.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// CurrentRuntime returns the RuntimeConfig snapshot for this process.
// The returned value copies scalar fields; the Git pointer is shared,
// so updates via SetGit are visible to subsequent callers.
func CurrentRuntime() RuntimeConfig {
	runtimeMu.RLock()
	defer runtimeMu.RUnlock()
	return runtime
}

// SetWorkDir overrides the WorkDir field. Intended for tests; also
// updates the Git workDir so live methods follow.
func SetWorkDir(dir string) {
	runtimeMu.Lock()
	defer runtimeMu.Unlock()
	runtime.WorkDir = dir
	if runtime.Git == nil {
		runtime.Git = &Git{workDir: dir}
	} else {
		runtime.Git.workDir = dir
	}
}

// SetGit attaches a fully-populated Git to the runtime. Called by the
// orchestrator at run start once the trigger has been parsed. Same
// instance also lives on RunContext.Git. Later calls overwrite.
func SetGit(g *Git) {
	runtimeMu.Lock()
	defer runtimeMu.Unlock()
	if g == nil {
		// Keep the field non-nil so callers don't panic.
		runtime.Git = &Git{workDir: runtime.WorkDir}
		return
	}
	runtime.Git = g
}
