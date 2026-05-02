package session

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContextHelpers(t *testing.T) {
	ctx := WithToken(context.Background(), "token-1")

	token, ok := TokenFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, "token-1", token)

	session := &Session{ID: "session-1", Subject: Subject{UserID: "user-1"}}
	ctx = WithSession(ctx, session)

	storedSession, ok := SessionFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, session, storedSession)

	user := &testUser{ID: "user-1", Name: "John Doe"}
	ctx = WithUser(ctx, user)

	anyUser, ok := AnyUserFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, user, anyUser)

	storedUser, ok := UserFromContext[*testUser](ctx)
	require.True(t, ok)
	require.Equal(t, user, storedUser)
}

type testUser struct {
	ID   string
	Name string
}
