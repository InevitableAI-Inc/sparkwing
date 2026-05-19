package runners_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sparkwing-dev/sparkwing/pkg/runners"
)

// ExampleLoad reads a small runners.yaml and lists the declared
// runners with their labels. Production code uses [runners.Resolve]
// to apply the repo / user precedence and the implicit-local synth.
func ExampleLoad() {
	dir, _ := os.MkdirTemp("", "sparkwing-runners-")
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "runners.yaml")
	_ = os.WriteFile(path, []byte(`
runners:
  local:
    type: local
    labels: [local, "os=linux"]
  cloud:
    type: kubernetes
    controller: shared
    labels: [cloud, "os=linux"]
`), 0o644)

	file, err := runners.Load(path)
	if err != nil {
		fmt.Println("load:", err)
		return
	}
	names := []string{"cloud", "local"} // sorted for deterministic output
	for _, n := range names {
		r := file.Runners[n]
		fmt.Printf("%s (%s): %s\n", n, r.Type, strings.Join(r.Labels, ","))
	}
	// Output:
	// cloud (kubernetes): cloud,os=linux
	// local (local): local,os=linux
}
