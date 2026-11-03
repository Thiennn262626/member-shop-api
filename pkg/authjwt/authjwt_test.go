package authjwt

import (
	"testing"
	"time"
)

func TestECDSAKeyManager(t *testing.T) {
	// 1. Auto generate keypair
	km, err := NewECDSAKeyManager("", "")
	if err != nil {
		t.Fatalf("failed to create ECDSA key manager: %v", err)
	}

	// 2. Sign token
	tokenStr, jti, exp, err := km.SignAccessToken("staff-123", "cashier01", "cashier", "Nguyen Van A", 15*time.Minute)
	if err != nil {
		t.Fatalf("failed to sign access token: %v", err)
	}
	if tokenStr == "" || jti == "" {
		t.Fatalf("expected non-empty token string and jti")
	}
	if exp.Before(time.Now()) {
		t.Fatalf("expected future expiration")
	}

	// 3. Validate token
	claims, err := km.ValidateAccessToken(tokenStr)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}
	if claims.JTI != jti {
		t.Errorf("expected JTI %s, got %s", jti, claims.JTI)
	}
	if claims.UserID != "staff-123" {
		t.Errorf("expected UserID staff-123, got %s", claims.UserID)
	}
	if claims.Username != "cashier01" {
		t.Errorf("expected Username cashier01, got %s", claims.Username)
	}

	// 4. Test Export / Re-import PEM
	privPEM, err := ExportECPrivateKeyPEM(km.PrivateKey)
	if err != nil {
		t.Fatalf("failed to export private key PEM: %v", err)
	}
	pubPEM, err := ExportECPublicKeyPEM(km.PublicKey)
	if err != nil {
		t.Fatalf("failed to export public key PEM: %v", err)
	}

	km2, err := NewECDSAKeyManager(privPEM, pubPEM)
	if err != nil {
		t.Fatalf("failed to reload from exported PEM: %v", err)
	}

	claims2, err := km2.ValidateAccessToken(tokenStr)
	if err != nil {
		t.Fatalf("failed to validate token with reloaded key manager: %v", err)
	}
	if claims2.JTI != jti {
		t.Errorf("expected JTI %s, got %s", jti, claims2.JTI)
	}
}

func TestGenerateRandomRefreshToken(t *testing.T) {
	t1, err := GenerateRandomRefreshToken()
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}
	t2, err := GenerateRandomRefreshToken()
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}

	if len(t1) != 64 {
		t.Errorf("expected 64 hex characters (32 bytes), got length %d", len(t1))
	}
	if t1 == t2 {
		t.Errorf("expected distinct random tokens, but got identical: %s", t1)
	}
}
