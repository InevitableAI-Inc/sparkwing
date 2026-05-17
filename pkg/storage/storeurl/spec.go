package storeurl

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sparkwing-dev/sparkwing/orchestrator/store"
	"github.com/sparkwing-dev/sparkwing/pkg/backends"
	"github.com/sparkwing-dev/sparkwing/pkg/storage"
	"github.com/sparkwing-dev/sparkwing/pkg/storage/fs"
	s3store "github.com/sparkwing-dev/sparkwing/pkg/storage/s3"
	"github.com/sparkwing-dev/sparkwing/pkg/storage/stdoutlogs"
)

// OpenArtifactStoreFromSpec constructs an ArtifactStore from a
// backends.Spec. The spec must already have passed pkg/backends
// validation (surface allow-list, required fields per type).
//
// Recognized but not yet implemented backend types return a clear
// error so callers surface a configuration problem instead of
// silently falling back.
func OpenArtifactStoreFromSpec(ctx context.Context, spec backends.Spec) (storage.ArtifactStore, error) {
	switch spec.Type {
	case backends.TypeFilesystem:
		path, err := expandPath(spec.Path)
		if err != nil {
			return nil, fmt.Errorf("cache filesystem: %w", err)
		}
		return fs.NewArtifactStore(path)
	case backends.TypeS3:
		client, err := newS3Client(ctx)
		if err != nil {
			return nil, err
		}
		return s3store.NewArtifactStore(spec.Bucket, spec.Prefix, client), nil
	case backends.TypeGCS, backends.TypeAzureBlob, backends.TypeController:
		return nil, unimplemented("cache", spec.Type)
	default:
		return nil, fmt.Errorf("cache backend type %q is not recognized", spec.Type)
	}
}

// OpenLogStoreFromSpec constructs a LogStore from a backends.Spec.
// See OpenArtifactStoreFromSpec for error semantics.
func OpenLogStoreFromSpec(ctx context.Context, spec backends.Spec) (storage.LogStore, error) {
	switch spec.Type {
	case backends.TypeFilesystem:
		path, err := expandPath(spec.Path)
		if err != nil {
			return nil, fmt.Errorf("logs filesystem: %w", err)
		}
		return fs.NewLogStore(path)
	case backends.TypeS3:
		client, err := newS3Client(ctx)
		if err != nil {
			return nil, err
		}
		return s3store.NewLogStore(spec.Bucket, spec.Prefix, client), nil
	case backends.TypeStdout:
		if err := stdoutlogs.CheckSpec(spec.Bucket, spec.Prefix, spec.Path, spec.URL, spec.URLSource, spec.Token); err != nil {
			return nil, err
		}
		return stdoutlogs.New(), nil
	case backends.TypeGCS, backends.TypeAzureBlob, backends.TypeController:
		return nil, unimplemented("logs", spec.Type)
	default:
		return nil, fmt.Errorf("logs backend type %q is not recognized", spec.Type)
	}
}

// OpenStateStoreFromSpec constructs a StateStore from a backends.Spec.
// See OpenArtifactStoreFromSpec for error semantics.
//
// For type=sqlite, spec.Path is required and names the SQLite database
// file. Callers that want the historical default (~/.sparkwing/state.db)
// should pass that path explicitly so the factory has a single,
// caller-provided source of truth.
func OpenStateStoreFromSpec(_ context.Context, spec backends.Spec) (storage.StateStore, error) {
	switch spec.Type {
	case backends.TypeSQLite:
		path, err := expandPath(spec.Path)
		if err != nil {
			return nil, fmt.Errorf("state sqlite: %w", err)
		}
		return store.Open(path)
	case backends.TypePostgres, backends.TypeMySQL, backends.TypeController:
		return nil, unimplemented("state", spec.Type)
	default:
		return nil, fmt.Errorf("state backend type %q is not recognized", spec.Type)
	}
}

func unimplemented(surface, t string) error {
	return fmt.Errorf("%s backend type %q is recognized but not implemented in this build", surface, t)
}

func expandPath(p string) (string, error) {
	if p == "" {
		return "", fmt.Errorf("path is required")
	}
	if strings.HasPrefix(p, "~/") || p == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if p == "~" {
			return home, nil
		}
		return home + p[1:], nil
	}
	return p, nil
}
