// Package conformance ships portable test suites for the public
// plug-in interfaces in [github.com/sparkwing-dev/sparkwing/pkg/storage].
// Implementations -- in-tree or downstream -- call the suite from
// their own *_test.go to prove they honor the contract.
//
// # Usage
//
//	package mybackend
//
//	import (
//	    "testing"
//	    "github.com/sparkwing-dev/sparkwing/pkg/storage"
//	    "github.com/sparkwing-dev/sparkwing/pkg/storage/conformance"
//	)
//
//	func TestConformance(t *testing.T) {
//	    conformance.TestArtifactStore(t, func() storage.ArtifactStore {
//	        return New(t.TempDir())
//	    })
//	    conformance.TestLogStore(t, func() storage.LogStore {
//	        return New(t.TempDir())
//	    })
//	}
//
// The factory must return a fresh, empty store for each subtest --
// the suite assumes isolation.
//
// # Partial implementations
//
// A backend that doesn't support every operation (a write-only
// LogStore, an [storage.ArtifactStore] without List, etc.) signals
// opt-out by returning an error that wraps
// [storage.ErrNotSupported] or the more specific
// [storage.ErrListNotSupported]. The suite checks via errors.Is
// and reports the subtest as skipped (with the wrapped error
// message) rather than failed. Partial implementations therefore
// pass conformance for what they DO support without artificial
// parity.
package conformance
