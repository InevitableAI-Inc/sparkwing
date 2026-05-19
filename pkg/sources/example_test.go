package sources_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sparkwing-dev/sparkwing/pkg/sources"
)

// ExampleLoad reads a small sources.yaml and prints the default
// source name + the dotenv source's path. Production code uses
// [sources.Resolve] to apply the repo / user precedence.
func ExampleLoad() {
	dir, _ := os.MkdirTemp("", "sparkwing-sources-")
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "sources.yaml")
	_ = os.WriteFile(path, []byte(`
default: dotenv
sources:
  dotenv:
    type: file
    path: .sparkwing/secrets.local.env
  shell:
    type: env
    prefix: SW_
`), 0o644)

	file, err := sources.Load(path)
	if err != nil {
		fmt.Println("load:", err)
		return
	}
	fmt.Println("default:", file.Default)
	fmt.Println("dotenv path:", file.Sources["dotenv"].Path)
	// Output:
	// default: dotenv
	// dotenv path: .sparkwing/secrets.local.env
}
