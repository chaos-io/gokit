package session

import (
	"context"
	"fmt"

	"github.com/chaos-io/chaos/pkg/logs"
	"github.com/go-kit/kit/endpoint"
)

type AuthProvider interface {
	GetLoginUser(ctx context.Context, req any) (*User, error)
}

type AuthProviderFunc func(ctx context.Context, req any) (*User, error)

func (f AuthProviderFunc) GetLoginUser(ctx context.Context, req any) (*User, error) {
	return f(ctx, req)
}

func NewSessionMW(ap AuthProvider, s ISession) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req any) (resp any, err error) {
			sessionID, ok := ctx.Value(SessionKey).(string)
			if !ok || sessionID == "" {
				return nil, fmt.Errorf("session not found")
			}
			logs.Debugw("session", "sessionID", sessionID)

			session, err := s.ValidateSession(ctx, sessionID)
			if err != nil {
				return nil, fmt.Errorf("falied to validate session, error: %v", err)
			}

			user, err := ap.GetLoginUser(ctx, req)
			if err != nil {
				return nil, fmt.Errorf("failed to get login user, error: %v", err)
			}

			// 验证用户ID是否匹配
			if user.ID != session.UserID {
				return nil, fmt.Errorf("user id mismatch")
			}

			// 注入用户信息到上下文
			ctx = WithCtxUser(ctx, user)

			return next(ctx, req)
		}
	}
}
