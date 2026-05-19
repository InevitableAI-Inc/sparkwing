package controller_test

import (
	"encoding/base64"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/sparkwing-dev/sparkwing/pkg/controller"
	"github.com/sparkwing-dev/sparkwing/pkg/store"
)

// ExampleNew shows the laptop-mode wiring: a sqlite-backed [store.Store]
// fronted by a [controller.Server] with no dispatcher, no auth, and no
// secrets cipher. This is the minimum viable controller; pkg/localws
// builds on top by adding the dashboard handler and a log store on the
// same mux.
func ExampleNew() {
	dir, _ := os.MkdirTemp("", "sparkwing-controller-")
	defer os.RemoveAll(dir)
	st, err := store.Open(filepath.Join(dir, "state.db"))
	if err != nil {
		fmt.Println("store:", err)
		return
	}

	srv := controller.New(st, nil)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	fmt.Println("controller routes mounted")
	// Output: controller routes mounted
}

// passthroughCipher is the no-op shape a custom cipher takes when an
// external consumer wants encryption-at-rest off but the route wired.
// Real implementations (AES-GCM, KMS, etc.) follow the same shape.
type passthroughCipher struct{}

func (passthroughCipher) Seal(plain string) (string, error) {
	return "raw:" + base64.StdEncoding.EncodeToString([]byte(plain)), nil
}

func (passthroughCipher) Open(envelope string) (string, error) {
	const prefix = "raw:"
	if len(envelope) < len(prefix) || envelope[:len(prefix)] != prefix {
		return "", fmt.Errorf("not a passthrough envelope")
	}
	b, err := base64.StdEncoding.DecodeString(envelope[len(prefix):])
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ExampleServer_WithSecretsCipher shows wiring a custom [controller.Cipher]
// implementation. The controller calls Seal before writing each secret
// to the store and Open when serving it back; the interface lets you
// plug in KMS, an HSM, or any other key-management approach without
// patching the controller.
func ExampleServer_WithSecretsCipher() {
	dir, _ := os.MkdirTemp("", "sparkwing-cipher-")
	defer os.RemoveAll(dir)
	st, _ := store.Open(filepath.Join(dir, "state.db"))

	srv := controller.New(st, nil).WithSecretsCipher(passthroughCipher{})
	_ = srv
	fmt.Println("server has cipher wired")
	// Output: server has cipher wired
}
