package secret

import (
	"bytes"
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := bytes.Repeat([]byte{1}, 32)
	envelope, err := Encrypt(key, "secret-password", "master-key")
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}
	if envelope.Ciphertext == "secret-password" {
		t.Fatal("ciphertext should not equal plaintext")
	}

	plaintext, err := Decrypt(key, envelope)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}
	if plaintext != "secret-password" {
		t.Fatalf("expected plaintext restored, got %q", plaintext)
	}
}
