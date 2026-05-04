package session

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserContextHelpers(t *testing.T) {
	user := &testUser{ID: "user-1", Name: "John Doe"}
	ctx := WithUser(context.Background(), user)

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
