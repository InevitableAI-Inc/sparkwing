package localws_test

import (
	"context"

	"github.com/sparkwing-dev/sparkwing/pkg/localws"
)

// ExampleRun shows the smallest call site for the local dev server.
// In real use, [localws.Run] is invoked by `sparkwing dashboard start`;
// it blocks until ctx is cancelled. This example is compile-only
// because [localws.Run] requires the embedded web bundle, which is a
// build-time artifact produced by `bash bin/build-web.sh`.
func ExampleRun() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_ = func() error {
		return localws.Run(ctx, localws.Options{
			Addr: "127.0.0.1:4343",
			Home: "/tmp/sparkwing-home",
		})
	}
}
