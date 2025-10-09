package session

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chaos-io/chaos/pkg/logs"
)

const (
	SessionKey     = "session_key"
	SessionExpires = 7 * 24 * time.Hour
)

// 用于签名的密钥（在实际应用中应从配置中读取或使用环境变量）
var hmacSecret = []byte("my-session-hmac-key")

type Session struct {
	UserID    string
	SessionID int64
	CreatedAt time.Time
	ExpiresAt time.Time
}

//go:generate mockgen -destination=mocks/session.go -package=mocks . ISession
type ISession interface {
	GenerateSessionKey(ctx context.Context, session *Session) (string, error)
	ValidateSession(ctx context.Context, sessionID string) (*Session, error)
}

type sessionImpl struct{}

func NewSessionImpl() ISession {
	return &sessionImpl{}
}

func (s *sessionImpl) GenerateSessionKey(ctx context.Context, session *Session) (string, error) {
	now := time.Now()
	session.CreatedAt = now
	session.ExpiresAt = now.Add(SessionExpires)

	data, err := json.Marshal(session)
	if err != nil {
		return "", err
	}

	// 计算HMAC签名以确定完整性
	h := hmac.New(sha256.New, hmacSecret)
	h.Write(data)
	sign := h.Sum(nil)

	finalData := append(data, sign...)

	// base64编码最终结果
	return base64.RawURLEncoding.EncodeToString(finalData), nil
}

func (s *sessionImpl) ValidateSession(ctx context.Context, sessionID string) (*Session, error) {
	logs.Debugw("validate session", "sessionID", sessionID)

	data, err := base64.RawURLEncoding.DecodeString(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to decode session id (%s): %w", sessionID, err)
	}

	if len(data) < 32 {
		return nil, fmt.Errorf("invalid session id (%s)", sessionID)
	}

	// 分离会话数据和签名
	signature := data[len(data)-32:]
	sessionData := data[:len(data)-32]

	h := hmac.New(sha256.New, hmacSecret)
	h.Write(sessionData)
	expectedSignature := h.Sum(nil)

	if !hmac.Equal(signature, expectedSignature) {
		return nil, fmt.Errorf("invalid session signature (%s)", sessionID)
	}

	var session Session
	if err := json.Unmarshal(sessionData, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data (%s): %w", sessionID, err)
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, fmt.Errorf("session expired")
	}

	return &session, nil
}
