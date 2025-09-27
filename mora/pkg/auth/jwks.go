package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// ErrJWKSFetch represents an error fetching JWKS
	ErrJWKSFetch = errors.New("failed to fetch JWKS")
	// ErrKeyNotFound represents a key not found error
	ErrKeyNotFound = errors.New("key not found in JWKS")
	// ErrInvalidKeyType represents an invalid key type error
	ErrInvalidKeyType = errors.New("invalid key type")
)

// JWK represents a JSON Web Key
type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKS represents a JSON Web Key Set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWKSValidator handles JWKS-based token validation
type JWKSValidator struct {
	jwksURL    string
	httpClient *http.Client
	cache      map[string]*rsa.PublicKey
	cacheTime  time.Time
	cacheTTL   time.Duration
}

// NewJWKSValidator creates a new JWKS validator
func NewJWKSValidator(jwksURL string) *JWKSValidator {
	return &JWKSValidator{
		jwksURL: jwksURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		cache:    make(map[string]*rsa.PublicKey),
		cacheTTL: 1 * time.Hour, // Cache keys for 1 hour
	}
}

// ValidateTokenWithJWKS validates a JWT token using JWKS
func (v *JWKSValidator) ValidateTokenWithJWKS(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, ErrInvalidToken
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Get the key ID from token header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("missing key ID in token header")
		}

		// Get the public key for this key ID
		publicKey, err := v.getPublicKey(kid)
		if err != nil {
			return nil, err
		}

		return publicKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, ErrMalformedToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.IsExpired() {
		return nil, ErrExpiredToken
	}

	return claims, nil
}

// ValidateTokenWithPublicKey validates a JWT token using a public key
func ValidateTokenWithPublicKey(tokenString, publicKeyPEM string) (*Claims, error) {
	if tokenString == "" {
		return nil, ErrInvalidToken
	}

	// Parse the public key
	publicKey, err := parsePublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, ErrMalformedToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.IsExpired() {
		return nil, ErrExpiredToken
	}

	return claims, nil
}

// getPublicKey retrieves a public key by key ID, using cache if available
func (v *JWKSValidator) getPublicKey(kid string) (*rsa.PublicKey, error) {
	// Check cache first
	if time.Since(v.cacheTime) < v.cacheTTL {
		if key, exists := v.cache[kid]; exists {
			return key, nil
		}
	}

	// Fetch JWKS
	jwks, err := v.fetchJWKS()
	if err != nil {
		return nil, err
	}

	// Find the key with matching kid
	for _, jwk := range jwks.Keys {
		if jwk.Kid == kid {
			publicKey, err := v.jwkToPublicKey(jwk)
			if err != nil {
				return nil, err
			}

			// Update cache
			v.cache[kid] = publicKey
			v.cacheTime = time.Now()

			return publicKey, nil
		}
	}

	return nil, ErrKeyNotFound
}

// fetchJWKS fetches the JWKS from the configured URL
func (v *JWKSValidator) fetchJWKS() (*JWKS, error) {
	resp, err := v.httpClient.Get(v.jwksURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrJWKSFetch, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: HTTP %d", ErrJWKSFetch, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrJWKSFetch, err)
	}

	var jwks JWKS
	if err := json.Unmarshal(body, &jwks); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrJWKSFetch, err)
	}

	return &jwks, nil
}

// jwkToPublicKey converts a JWK to an RSA public key
func (v *JWKSValidator) jwkToPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	if jwk.Kty != "RSA" {
		return nil, ErrInvalidKeyType
	}

	// Decode the modulus (n)
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}

	// Decode the exponent (e)
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	// Convert bytes to big integers
	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	// Create the RSA public key
	publicKey := &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}

	return publicKey, nil
}

// parsePublicKeyFromPEM parses a public key from PEM format
func parsePublicKeyFromPEM(publicKeyPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	var publicKey *rsa.PublicKey
	var err error

	switch block.Type {
	case "PUBLIC KEY":
		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKIX public key: %w", err)
		}
		var ok bool
		publicKey, ok = pub.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("public key is not RSA")
		}
	case "RSA PUBLIC KEY":
		publicKey, err = x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKCS1 public key: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported PEM block type: %s", block.Type)
	}

	return publicKey, nil
}

// GenerateTokenWithPrivateKey generates a JWT token using RSA private key
func GenerateTokenWithPrivateKey(userID, username, privateKeyPEM string, ttl time.Duration) (string, error) {
	claims := NewClaims(userID, username, ttl)

	// Parse the private key
	privateKey, err := parsePrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	// Create token with RSA256 signing method
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Sign the token
	return token.SignedString(privateKey)
}

// parsePrivateKeyFromPEM parses a private key from PEM format
func parsePrivateKeyFromPEM(privateKeyPEM string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	var privateKey *rsa.PrivateKey
	var err error

	switch block.Type {
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKCS8 private key: %w", err)
		}
		var ok bool
		privateKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("private key is not RSA")
		}
	case "RSA PRIVATE KEY":
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKCS1 private key: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported PEM block type: %s", block.Type)
	}

	return privateKey, nil
}

// SetKeyID sets the key ID in the token header for JWKS validation
func SetKeyID(token *jwt.Token, keyID string) {
	token.Header["kid"] = keyID
}

// GetKeyIDFromToken extracts the key ID from token header
func GetKeyIDFromToken(tokenString string) (string, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return "", ErrMalformedToken
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", ErrMalformedToken
	}

	var header map[string]interface{}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return "", ErrMalformedToken
	}

	kid, ok := header["kid"].(string)
	if !ok {
		return "", errors.New("missing key ID in token header")
	}

	return kid, nil
}
