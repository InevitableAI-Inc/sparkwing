package color_test

import (
	"fmt"

	"github.com/sparkwing-dev/sparkwing/pkg/color"
)

// ExampleGreen forces color off so the example output is the bare
// formatted string. In a real terminal session the same call would
// be wrapped in ANSI escape codes.
func ExampleGreen() {
	color.SetEnabled(false)
	fmt.Println(color.Green("deployed %s", "v1.5.5"))
	// Output: deployed v1.5.5
}
