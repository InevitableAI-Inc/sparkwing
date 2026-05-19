package backends_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sparkwing-dev/sparkwing/pkg/backends"
)

// ExampleLoad writes a small backends.yaml to a tempdir and inspects
// the default cache backend. Production code uses [backends.Resolve]
// to apply the standard repo / user precedence.
func ExampleLoad() {
	dir, _ := os.MkdirTemp("", "sparkwing-backends-")
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "backends.yaml")
	_ = os.WriteFile(path, []byte(`
defaults:
  cache:
    type: filesystem
    path: /tmp/sparkwing-cache
`), 0o644)

	file, err := backends.Load(path)
	if err != nil {
		fmt.Println("load:", err)
		return
	}
	fmt.Printf("cache: type=%s\n", file.Defaults.Cache.Type)
	// Output: cache: type=filesystem
}
