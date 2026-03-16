package services

import (
	"strings"
	"testing"
)

func TestAPIKeyService(t *testing.T) {
	svc := NewAPIKeyService()

	t.Run("GenerateKey_Format", func(t *testing.T) {
		fullKey, prefix, hash, err := svc.GenerateKey()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !strings.HasPrefix(fullKey, "ll_") {
			t.Errorf("expected key to start with ll_, got %s", fullKey[:10])
		}

		// ll_ (3) + 48 random chars = 51 total
		if len(fullKey) != 51 {
			t.Errorf("expected key length 51, got %d", len(fullKey))
		}

		if len(prefix) != 8 {
			t.Errorf("expected prefix length 8, got %d", len(prefix))
		}

		// Prefix should match first 8 chars after ll_
		if fullKey[3:11] != prefix {
			t.Errorf("prefix %s doesn't match key substring %s", prefix, fullKey[3:11])
		}

		if hash == "" {
			t.Fatal("expected non-empty hash")
		}

		// Hash should be a bcrypt hash (starts with $2a$ or $2b$)
		if !strings.HasPrefix(hash, "$2") {
			t.Errorf("expected bcrypt hash, got %s", hash[:10])
		}
	})

	t.Run("GenerateKey_Uniqueness", func(t *testing.T) {
		keys := make(map[string]bool)
		for i := 0; i < 100; i++ {
			fullKey, _, _, err := svc.GenerateKey()
			if err != nil {
				t.Fatalf("expected no error on iteration %d, got %v", i, err)
			}
			if keys[fullKey] {
				t.Fatalf("duplicate key generated on iteration %d", i)
			}
			keys[fullKey] = true
		}
	})

	t.Run("VerifyKey_Valid", func(t *testing.T) {
		fullKey, _, hash, err := svc.GenerateKey()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !svc.VerifyKey(fullKey, hash) {
			t.Error("expected key to verify successfully")
		}
	})

	t.Run("VerifyKey_Invalid", func(t *testing.T) {
		_, _, hash, err := svc.GenerateKey()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if svc.VerifyKey("ll_wrongkeywrongkeywrongkeywrongkeywrongkeywrongkey1", hash) {
			t.Error("expected verification to fail for wrong key")
		}
	})

	t.Run("ExtractPrefix", func(t *testing.T) {
		prefix := svc.ExtractPrefix("ll_abcdefgh1234567890123456789012345678901234567890")
		if prefix != "abcdefgh" {
			t.Errorf("expected prefix abcdefgh, got %s", prefix)
		}
	})

	t.Run("ExtractPrefix_TooShort", func(t *testing.T) {
		prefix := svc.ExtractPrefix("ll_abc")
		if prefix != "" {
			t.Errorf("expected empty prefix for short key, got %s", prefix)
		}
	})

	t.Run("GenerateKey_AlphanumericOnly", func(t *testing.T) {
		for i := 0; i < 20; i++ {
			fullKey, _, _, err := svc.GenerateKey()
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			// Check that the random part (after ll_) is all alphanumeric
			randomPart := fullKey[3:]
			for _, ch := range randomPart {
				if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')) {
					t.Errorf("unexpected character %c in key", ch)
				}
			}
		}
	})
}
