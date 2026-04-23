package session

import "context"

type sessionContextKey struct{}

func WithSessionID(ctx context.Context, sessionID string) context.Context {
	if len(sessionID) == 0 {
		return ctx
	}

	ctx = context.WithValue(ctx, sessionContextKey{}, sessionID)
	return context.WithValue(ctx, SessionKey, sessionID)
}

func SessionIDFromContext(ctx context.Context) (string, bool) {
	if sessionID, ok := ctx.Value(sessionContextKey{}).(string); ok && len(sessionID) > 0 {
		return sessionID, true
	}

	sessionID, ok := ctx.Value(SessionKey).(string)
	return sessionID, ok && len(sessionID) > 0
}
