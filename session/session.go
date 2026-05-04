package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"
)

const DefaultTTL = 7 * 24 * time.Hour

var (
	ErrStoreRequired            = errors.New("session store is required")
	ErrTokenCodecRequired       = errors.New("session token codec is required")
	ErrValidatorRequired        = errors.New("session validator is required")
	ErrTTLInvalid               = errors.New("session ttl must be positive")
	ErrSessionIDGeneratorNeeded = errors.New("session id generator is required")
	ErrSessionUserIDRequired    = errors.New("session user id is required")
	ErrSessionIDRequired        = errors.New("session id is required")
	ErrSessionInvalid           = errors.New("session is invalid")
	ErrSessionNotFound          = errors.New("session not found")
	ErrSessionExpired           = errors.New("session expired")
	ErrSessionRevoked           = errors.New("session revoked")
	ErrSessionStateMismatch     = errors.New("session state does not match token claims")
	ErrRevocationUnsupported    = errors.New("session store does not support revocation")
)

type Subject struct {
	UserID string `json:"user_id"`
	AppID  int32  `json:"app_id,omitempty"`
}

type Session struct {
	ID        string    `json:"id"`
	Subject   Subject   `json:"subject"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
	RevokedAt time.Time `json:"revoked_at,omitempty"`
}

type IssuedSession struct {
	Token   string   `json:"token"`
	Session *Session `json:"session"`
}

type TokenClaims struct {
	SessionID string `json:"sid"`
	UserID    string `json:"uid"`
	AppID     int32  `json:"aid,omitempty"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

type Issuer interface {
	Issue(ctx context.Context, subject Subject) (*IssuedSession, error)
}

type Validator interface {
	Validate(ctx context.Context, token string) (*Session, error)
}

type Revoker interface {
	Revoke(ctx context.Context, sessionID string) error
}

type Store interface {
	Save(ctx context.Context, session *Session) error
	Find(ctx context.Context, sessionID string) (*Session, error)
}

type RevocationStore interface {
	Revoke(ctx context.Context, sessionID string, revokedAt time.Time) error
}

type TokenCodec interface {
	Encode(ctx context.Context, claims TokenClaims) (string, error)
	Decode(ctx context.Context, token string) (*TokenClaims, error)
}

type Option func(*config) error

type config struct {
	ttl         time.Duration
	now         func() time.Time
	idGenerator func() (string, error)
}

type Manager struct {
	store      Store
	codec      TokenCodec
	ttl        time.Duration
	now        func() time.Time
	generateID func() (string, error)
}

func WithTTL(ttl time.Duration) Option {
	return func(cfg *config) error {
		if ttl <= 0 {
			return ErrTTLInvalid
		}
		cfg.ttl = ttl
		return nil
	}
}

func WithClock(now func() time.Time) Option {
	return func(cfg *config) error {
		if now == nil {
			return nil
		}
		cfg.now = now
		return nil
	}
}

func WithIDGenerator(generator func() (string, error)) Option {
	return func(cfg *config) error {
		if generator == nil {
			return ErrSessionIDGeneratorNeeded
		}
		cfg.idGenerator = generator
		return nil
	}
}

func NewManager(store Store, codec TokenCodec, opts ...Option) (*Manager, error) {
	if store == nil {
		return nil, ErrStoreRequired
	}
	if codec == nil {
		return nil, ErrTokenCodecRequired
	}

	cfg, err := newConfig(opts...)
	if err != nil {
		return nil, err
	}

	return &Manager{
		store:      store,
		codec:      codec,
		ttl:        cfg.ttl,
		now:        cfg.now,
		generateID: cfg.idGenerator,
	}, nil
}

func (m *Manager) Issue(ctx context.Context, subject Subject) (*IssuedSession, error) {
	if subject.UserID == "" {
		return nil, ErrSessionUserIDRequired
	}

	sessionID, err := m.generateID()
	if err != nil {
		return nil, fmt.Errorf("generate session id: %w", err)
	}
	if sessionID == "" {
		return nil, ErrSessionIDRequired
	}

	now := m.now().UTC().Truncate(time.Second)
	session := &Session{
		ID:        sessionID,
		Subject:   subject,
		IssuedAt:  now,
		ExpiresAt: now.Add(m.ttl),
	}

	token, err := m.codec.Encode(ctx, claimsFromSession(session))
	if err != nil {
		return nil, fmt.Errorf("encode session token: %w", err)
	}

	if err := m.store.Save(ctx, session); err != nil {
		return nil, fmt.Errorf("save session: %w", err)
	}

	return &IssuedSession{
		Token:   token,
		Session: cloneSession(session),
	}, nil
}

func (m *Manager) Validate(ctx context.Context, token string) (*Session, error) {
	claims, err := m.codec.Decode(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("decode session token: %w", err)
	}

	session, err := m.store.Find(ctx, claims.SessionID)
	if err != nil {
		return nil, fmt.Errorf("find session: %w", err)
	}

	if err := validateSession(claims, session, m.now().UTC()); err != nil {
		return nil, err
	}

	return cloneSession(session), nil
}

func (m *Manager) Revoke(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return ErrSessionIDRequired
	}

	store, ok := m.store.(RevocationStore)
	if !ok {
		return ErrRevocationUnsupported
	}

	if err := store.Revoke(ctx, sessionID, m.now().UTC().Truncate(time.Second)); err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}

	return nil
}

func newConfig(opts ...Option) (*config, error) {
	cfg := &config{
		ttl:         DefaultTTL,
		now:         time.Now,
		idGenerator: randomSessionID,
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.ttl <= 0 {
		return nil, ErrTTLInvalid
	}
	if cfg.idGenerator == nil {
		return nil, ErrSessionIDGeneratorNeeded
	}

	return cfg, nil
}

func claimsFromSession(session *Session) TokenClaims {
	return TokenClaims{
		SessionID: session.ID,
		UserID:    session.Subject.UserID,
		AppID:     session.Subject.AppID,
		IssuedAt:  session.IssuedAt.Unix(),
		ExpiresAt: session.ExpiresAt.Unix(),
	}
}

func validateSession(claims *TokenClaims, session *Session, now time.Time) error {
	switch {
	case claims == nil:
		return ErrSessionInvalid
	case session == nil:
		return ErrSessionNotFound
	case session.ID == "":
		return ErrSessionInvalid
	case session.Subject.UserID == "":
		return ErrSessionInvalid
	case !session.RevokedAt.IsZero():
		return ErrSessionRevoked
	case now.After(session.ExpiresAt):
		return ErrSessionExpired
	case claims.SessionID != session.ID:
		return ErrSessionStateMismatch
	case claims.UserID != session.Subject.UserID:
		return ErrSessionStateMismatch
	case claims.AppID != session.Subject.AppID:
		return ErrSessionStateMismatch
	case claims.IssuedAt != session.IssuedAt.Unix():
		return ErrSessionStateMismatch
	case claims.ExpiresAt != session.ExpiresAt.Unix():
		return ErrSessionStateMismatch
	default:
		return nil
	}
}

func randomSessionID() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func cloneSession(session *Session) *Session {
	if session == nil {
		return nil
	}

	cloned := *session
	return &cloned
}
