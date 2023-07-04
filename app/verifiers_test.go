// Package app for pkid app
package app

import (
	"testing"

	"github.com/rawdaGastan/pkid/client"
)

func TestVerifiers(t *testing.T) {
	_, publicKey, err := client.GenerateKeyPair()

	if err != nil {
		t.Errorf("error generating keys: %q", err)
	}

	t.Run("test_wrong_encoding", func(t *testing.T) {
		encoded := "XXXXXaGVsbG8="

		_, err := verifySignedData(encoded, publicKey)
		if err == nil {
			t.Error("decoding should fail")
		}
	})

	t.Run("test_wrong_encoding_header", func(t *testing.T) {
		encoded := "XXXXXaGVsbG8="

		_, err := verifySignedHeader(encoded, publicKey)
		if err == nil {
			t.Error("decoding should fail")
		}
	})
}
