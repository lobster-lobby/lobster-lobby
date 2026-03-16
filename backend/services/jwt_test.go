package services

import (
	"testing"
	"time"
)

func TestJWTService(t *testing.T) {
	svc := NewJWTService("test-secret-key-for-testing")

	t.Run("GenerateAndValidateAccessToken", func(t *testing.T) {
		token, err := svc.GenerateAccessToken("507f1f77bcf86cd799439011", "human")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if token == "" {
			t.Fatal("expected non-empty token")
		}

		claims, err := svc.ValidateAccessToken(token)
		if err != nil {
			t.Fatalf("expected valid token, got error: %v", err)
		}
		if claims.Subject != "507f1f77bcf86cd799439011" {
			t.Errorf("expected sub=507f1f77bcf86cd799439011, got %s", claims.Subject)
		}
		if claims.Type != "human" {
			t.Errorf("expected type=human, got %s", claims.Type)
		}
	})

	t.Run("InvalidTokenRejected", func(t *testing.T) {
		_, err := svc.ValidateAccessToken("not.a.valid.token")
		if err == nil {
			t.Fatal("expected error for invalid token")
		}
	})

	t.Run("GenerateRefreshToken", func(t *testing.T) {
		token, expiresAt, err := svc.GenerateRefreshToken()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if token == "" {
			t.Fatal("expected non-empty refresh token")
		}
		if expiresAt.Before(time.Now().Add(6 * 24 * time.Hour)) {
			t.Error("expected expiry ~7 days from now")
		}
	})

	t.Run("WrongSecretRejected", func(t *testing.T) {
		token, _ := svc.GenerateAccessToken("userid", "human")
		otherSvc := NewJWTService("different-secret")
		_, err := otherSvc.ValidateAccessToken(token)
		if err == nil {
			t.Fatal("expected error for wrong secret")
		}
	})
}
