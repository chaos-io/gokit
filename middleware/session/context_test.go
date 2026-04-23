package session

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithSessionID(t *testing.T) {
	ctx := WithSessionID(context.Background(), "session-1")

	sessionID, ok := SessionIDFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, "session-1", sessionID)

	legacy, legacyOK := ctx.Value(SessionKey).(string)
	require.True(t, legacyOK)
	require.Equal(t, "session-1", legacy)
}
