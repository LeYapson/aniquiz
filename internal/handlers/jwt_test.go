package handlers_test

import (
	"testing"
	"time"

	"github.com/LeYapson/aniquiz/internal/handlers"
	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateToken_ReturnsNonEmptyString(t *testing.T) {
	tok, err := handlers.GenerateToken(42, "testuser")
	if err != nil {
		t.Fatalf("erreur inattendue: %v", err)
	}
	if tok == "" {
		t.Error("le token ne doit pas être vide")
	}
}

func TestValidateToken_ValidToken(t *testing.T) {
	tok, _ := handlers.GenerateToken(42, "testuser")

	claims, err := handlers.ValidateToken(tok)
	if err != nil {
		t.Fatalf("token valide rejeté: %v", err)
	}
	if claims.UserID != 42 {
		t.Errorf("UserID: got %d, want 42", claims.UserID)
	}
	if claims.Username != "testuser" {
		t.Errorf("Username: got %s, want testuser", claims.Username)
	}
}

func TestValidateToken_InvalidString(t *testing.T) {
	_, err := handlers.ValidateToken("not-a-jwt")
	if err == nil {
		t.Error("un token invalide doit retourner une erreur")
	}
}

func TestValidateToken_WrongSignature(t *testing.T) {
	// Token signé avec un secret différent
	claims := handlers.Claims{
		UserID:   1,
		Username: "hacker",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	tok, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("wrong-secret"))

	_, err := handlers.ValidateToken(tok)
	if err == nil {
		t.Error("un token avec une signature incorrecte doit être rejeté")
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	claims := handlers.Claims{
		UserID:   1,
		Username: "testuser",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	}
	// On signe avec le même secret que le code (vide en dev = "dev-secret-change-in-prod")
	tok, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("dev-secret-change-in-prod"))

	_, err := handlers.ValidateToken(tok)
	if err == nil {
		t.Error("un token expiré doit être rejeté")
	}
}
