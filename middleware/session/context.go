package session

import "context"

type tokenContextKey struct{}
type sessionContextKey struct{}
type userContextKey struct{}

func WithToken(ctx context.Context, token string) context.Context {
	if token == "" {
		return ctx
	}
	return context.WithValue(ctx, tokenContextKey{}, token)
}

func TokenFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(tokenContextKey{}).(string)
	return token, ok && token != ""
}

func WithSession(ctx context.Context, session *Session) context.Context {
	if session == nil {
		return ctx
	}
	return context.WithValue(ctx, sessionContextKey{}, session)
}

func SessionFromContext(ctx context.Context) (*Session, bool) {
	session, ok := ctx.Value(sessionContextKey{}).(*Session)
	if !ok || session == nil {
		return nil, false
	}
	return session, true
}

func WithUser(ctx context.Context, user any) context.Context {
	if user == nil {
		return ctx
	}
	return context.WithValue(ctx, userContextKey{}, user)
}

func AnyUserFromContext(ctx context.Context) (any, bool) {
	user := ctx.Value(userContextKey{})
	if user == nil {
		return nil, false
	}
	return user, true
}

func UserFromContext[T any](ctx context.Context) (T, bool) {
	user, ok := ctx.Value(userContextKey{}).(T)
	if !ok {
		var zero T
		return zero, false
	}
	return user, true
}
