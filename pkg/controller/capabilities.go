package controller

import (
	"errors"
	"io"
	"net/http"

	"github.com/sparkwing-dev/sparkwing/pkg/storage"
)

// handleArtifactGet streams the artifact at {key} to the response.
// Registered only when WithArtifactStore was set; absent otherwise.
// Returns 404 for unknown keys.
func (s *Server) handleArtifactGet(w http.ResponseWriter, r *http.Request) {
	if s.artifactStore == nil {
		// Defensive: the route is gated at register-time, so this
		// branch is only reachable if a caller invoked the handler
		// directly (a test, typically). Mirror the gated-route
		// behavior so callers see one shape regardless.
		http.NotFound(w, r)
		return
	}
	key := r.PathValue("key")
	if key == "" {
		http.Error(w, "missing key", http.StatusBadRequest)
		return
	}
	rc, err := s.artifactStore.Get(r.Context(), key)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer rc.Close()
	w.Header().Set("Content-Type", "application/octet-stream")
	_, _ = io.Copy(w, rc)
}
