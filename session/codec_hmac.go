package session

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const (
	MinSigningSecretLength = 32
	tokenAlgorithm         = "HS256"
	tokenType              = "gokit-session.v1"
)

var (
	ErrKeyringRequired       = errors.New("session keyring is required")
	ErrSigningKeyIDRequired  = errors.New("session signing key id is required")
	ErrSigningSecretRequired = errors.New("session signing secret is required")
	ErrWeakSigningSecret     = fmt.Errorf("session signing secret must be at least %d bytes", MinSigningSecretLength)
	ErrTokenRequired         = errors.New("session token is required")
	ErrTokenMalformed        = errors.New("session token is malformed")
	ErrTokenHeaderInvalid    = errors.New("session token header is invalid")
	ErrTokenClaimsInvalid    = errors.New("session token claims are invalid")
	ErrTokenKeyNotFound      = errors.New("session signing key was not found")
	ErrTokenSignatureInvalid = errors.New("session token signature is invalid")
)

type Key struct {
	ID     string
	Secret []byte
}

type Keyring interface {
	Active(ctx context.Context) (Key, error)
	Lookup(ctx context.Context, keyID string) (Key, error)
}

type StaticKeyring struct {
	active string
	keys   map[string]Key
}

type HMACCodec struct {
	keyring Keyring
}

type tokenHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
	KeyID     string `json:"kid"`
}

func NewStaticKeyring(active Key, fallback ...Key) (*StaticKeyring, error) {
	keys := make(map[string]Key, 1+len(fallback))

	register := func(key Key) error {
		normalized, err := normalizeKey(key)
		if err != nil {
			return err
		}
		if _, exists := keys[normalized.ID]; exists {
			return fmt.Errorf("duplicate session signing key id %q", normalized.ID)
		}
		keys[normalized.ID] = normalized
		return nil
	}

	if err := register(active); err != nil {
		return nil, err
	}
	for _, key := range fallback {
		if err := register(key); err != nil {
			return nil, err
		}
	}

	return &StaticKeyring{
		active: active.ID,
		keys:   keys,
	}, nil
}

func NewHMACCodec(keyring Keyring) (*HMACCodec, error) {
	if keyring == nil {
		return nil, ErrKeyringRequired
	}

	return &HMACCodec{keyring: keyring}, nil
}

func (k *StaticKeyring) Active(ctx context.Context) (Key, error) {
	_ = ctx

	key, ok := k.keys[k.active]
	if !ok {
		return Key{}, ErrTokenKeyNotFound
	}

	return cloneKey(key), nil
}

func (k *StaticKeyring) Lookup(ctx context.Context, keyID string) (Key, error) {
	_ = ctx

	key, ok := k.keys[keyID]
	if !ok {
		return Key{}, ErrTokenKeyNotFound
	}

	return cloneKey(key), nil
}

func (c *HMACCodec) Encode(ctx context.Context, claims TokenClaims) (string, error) {
	if err := validateClaims(claims); err != nil {
		return "", err
	}

	key, err := c.keyring.Active(ctx)
	if err != nil {
		return "", err
	}

	header := tokenHeader{
		Algorithm: tokenAlgorithm,
		Type:      tokenType,
		KeyID:     key.ID,
	}

	headerPart, err := encodeTokenPart(header)
	if err != nil {
		return "", err
	}

	claimsPart, err := encodeTokenPart(claims)
	if err != nil {
		return "", err
	}

	signingInput := headerPart + "." + claimsPart
	signature := signHMAC(key.Secret, []byte(signingInput))

	return signingInput + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func (c *HMACCodec) Decode(ctx context.Context, token string) (*TokenClaims, error) {
	if token == "" {
		return nil, ErrTokenRequired
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrTokenMalformed
	}

	var header tokenHeader
	if err := decodeTokenPart(parts[0], &header); err != nil {
		return nil, ErrTokenHeaderInvalid
	}
	if header.Algorithm != tokenAlgorithm || header.Type != tokenType || header.KeyID == "" {
		return nil, ErrTokenHeaderInvalid
	}

	key, err := c.keyring.Lookup(ctx, header.KeyID)
	if err != nil {
		return nil, err
	}

	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, ErrTokenMalformed
	}

	signingInput := parts[0] + "." + parts[1]
	expectedSignature := signHMAC(key.Secret, []byte(signingInput))
	if !hmac.Equal(signature, expectedSignature) {
		return nil, ErrTokenSignatureInvalid
	}

	var claims TokenClaims
	if err := decodeTokenPart(parts[1], &claims); err != nil {
		return nil, ErrTokenClaimsInvalid
	}
	if err := validateClaims(claims); err != nil {
		return nil, err
	}

	return &claims, nil
}

func validateClaims(claims TokenClaims) error {
	switch {
	case claims.SessionID == "":
		return ErrTokenClaimsInvalid
	case claims.UserID == "":
		return ErrTokenClaimsInvalid
	case claims.IssuedAt <= 0:
		return ErrTokenClaimsInvalid
	case claims.ExpiresAt <= 0:
		return ErrTokenClaimsInvalid
	case claims.ExpiresAt <= claims.IssuedAt:
		return ErrTokenClaimsInvalid
	default:
		return nil
	}
}

func normalizeKey(key Key) (Key, error) {
	switch {
	case key.ID == "":
		return Key{}, ErrSigningKeyIDRequired
	case len(key.Secret) == 0:
		return Key{}, ErrSigningSecretRequired
	case len(key.Secret) < MinSigningSecretLength:
		return Key{}, ErrWeakSigningSecret
	default:
		return Key{
			ID:     key.ID,
			Secret: bytes.Clone(key.Secret),
		}, nil
	}
}

func cloneKey(key Key) Key {
	key.Secret = bytes.Clone(key.Secret)
	return key
}

func encodeTokenPart(value any) (string, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decodeTokenPart(part string, target any) error {
	payload, err := base64.RawURLEncoding.DecodeString(part)
	if err != nil {
		return err
	}
	return json.Unmarshal(payload, target)
}

func signHMAC(secret, payload []byte) []byte {
	h := hmac.New(sha256.New, secret)
	h.Write(payload)
	return h.Sum(nil)
}
