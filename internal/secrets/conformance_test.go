package secrets_test

import (
	"testing"

	"github.com/sparkwing-dev/sparkwing/internal/secrets"
	"github.com/sparkwing-dev/sparkwing/pkg/controller"
	"github.com/sparkwing-dev/sparkwing/pkg/controller/ciphertest"
)

// TestConformance_Cipher wires the shared cipher contract suite
// against the internal/secrets AEAD implementation. New ciphers
// (KMS-backed, HSM-backed, custom AEAD) can drop into the same
// pattern from their own *_test.go.
func TestConformance_Cipher(t *testing.T) {
	key, err := secrets.GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	ciphertest.TestCipher(t, func() controller.Cipher {
		c, err := secrets.NewCipher(key)
		if err != nil {
			t.Fatalf("NewCipher: %v", err)
		}
		return c
	})
}
