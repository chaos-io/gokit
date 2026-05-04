package session

import "context"

type userContextKey struct{}

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
