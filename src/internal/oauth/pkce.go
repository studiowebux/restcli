package oauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// GenerateCodeVerifier generates a random PKCE code verifier
// The code verifier is a high-entropy cryptographic random string
// with length between 43-128 characters
func GenerateCodeVerifier() (string, error) {
	// Generate 32 random bytes (will be base64url encoded to 43 chars)
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Base64url encode (without padding)
	verifier := base64.RawURLEncoding.EncodeToString(bytes)
	return verifier, nil
}

// GenerateCodeChallenge generates the PKCE code challenge from the verifier
// Uses SHA256 hashing as per RFC 7636
func GenerateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])
	return challenge
}

// PKCEPair holds the verifier and challenge for an OAuth flow
type PKCEPair struct {
	Verifier  string
	Challenge string
}

// GeneratePKCEPair generates both verifier and challenge
func GeneratePKCEPair() (*PKCEPair, error) {
	verifier, err := GenerateCodeVerifier()
	if err != nil {
		return nil, err
	}

	challenge := GenerateCodeChallenge(verifier)

	return &PKCEPair{
		Verifier:  verifier,
		Challenge: challenge,
	}, nil
}
