package session

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestManagerIssueValidateRevoke(t *testing.T) {
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	manager := newTestManager(t, NewMemoryStore(), testClock(&now), testIDSequence("session-1"))

	issued, err := manager.Issue(context.Background(), Subject{UserID: "user-1", AppID: 7})
	require.NoError(t, err)
	require.Equal(t, "session-1", issued.Session.ID)
	require.Equal(t, now, issued.Session.IssuedAt)
	require.Equal(t, now.Add(DefaultTTL), issued.Session.ExpiresAt)

	validated, err := manager.Validate(context.Background(), issued.Token)
	require.NoError(t, err)
	require.Equal(t, issued.Session, validated)

	err = manager.Revoke(context.Background(), issued.Session.ID)
	require.NoError(t, err)

	_, err = manager.Validate(context.Background(), issued.Token)
	require.ErrorIs(t, err, ErrSessionRevoked)
}

func TestManagerValidatesAgainstStoreState(t *testing.T) {
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	store := NewMemoryStore()
	manager := newTestManager(t, store, testClock(&now), testIDSequence("session-1"))

	issued, err := manager.Issue(context.Background(), Subject{UserID: "user-1"})
	require.NoError(t, err)

	stored, err := store.Find(context.Background(), issued.Session.ID)
	require.NoError(t, err)
	stored.Subject.UserID = "other-user"
	require.NoError(t, store.Save(context.Background(), stored))

	_, err = manager.Validate(context.Background(), issued.Token)
	require.ErrorIs(t, err, ErrSessionStateMismatch)
}

func TestManagerSupportsSingleSessionPerUserViaStore(t *testing.T) {
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	store := NewMemoryStore(WithSingleSessionPerUser())
	manager := newTestManager(t, store, testClock(&now), testIDSequence("session-1", "session-2"))

	first, err := manager.Issue(context.Background(), Subject{UserID: "user-1"})
	require.NoError(t, err)

	now = now.Add(time.Minute)
	second, err := manager.Issue(context.Background(), Subject{UserID: "user-1"})
	require.NoError(t, err)

	_, err = manager.Validate(context.Background(), first.Token)
	require.ErrorIs(t, err, ErrSessionRevoked)

	active := store.activeSessionIDs[newSubjectKey(Subject{UserID: "user-1"})]
	require.Len(t, active, 1)
	_, hasFirst := active[first.Session.ID]
	require.False(t, hasFirst)
	_, hasSecond := active[second.Session.ID]
	require.True(t, hasSecond)

	validated, err := manager.Validate(context.Background(), second.Token)
	require.NoError(t, err)
	require.Equal(t, second.Session, validated)
}

func TestManagerRejectsExpiredSession(t *testing.T) {
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	manager := newTestManager(
		t,
		NewMemoryStore(),
		testClock(&now),
		testIDSequence("session-1"),
		WithTTL(time.Minute),
	)

	issued, err := manager.Issue(context.Background(), Subject{UserID: "user-1"})
	require.NoError(t, err)

	now = now.Add(2 * time.Minute)

	_, err = manager.Validate(context.Background(), issued.Token)
	require.ErrorIs(t, err, ErrSessionExpired)
}

func TestNewManagerValidatesDependencies(t *testing.T) {
	keyring, err := NewStaticKeyring(testActiveKey())
	require.NoError(t, err)
	codec, err := NewHMACCodec(keyring)
	require.NoError(t, err)

	_, err = NewManager(nil, codec)
	require.ErrorIs(t, err, ErrStoreRequired)

	_, err = NewManager(NewMemoryStore(), nil)
	require.ErrorIs(t, err, ErrTokenCodecRequired)

	_, err = NewManager(NewMemoryStore(), codec, WithTTL(0))
	require.ErrorIs(t, err, ErrTTLInvalid)

	_, err = NewManager(NewMemoryStore(), codec, WithIDGenerator(nil))
	require.ErrorIs(t, err, ErrSessionIDGeneratorNeeded)
}

func TestManagerRevokeRequiresRevocationStore(t *testing.T) {
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	manager := newTestManager(t, stubStore{}, testClock(&now), testIDSequence("session-1"))

	err := manager.Revoke(context.Background(), "session-1")
	require.ErrorIs(t, err, ErrRevocationUnsupported)
}

type stubStore struct{}

func (stubStore) Save(ctx context.Context, session *Session) error {
	return nil
}

func (stubStore) Find(ctx context.Context, sessionID string) (*Session, error) {
	return nil, errors.New("not implemented")
}

func newTestManager(t *testing.T, store Store, opts ...Option) *Manager {
	t.Helper()

	keyring, err := NewStaticKeyring(testActiveKey(), testFallbackKey())
	require.NoError(t, err)

	codec, err := NewHMACCodec(keyring)
	require.NoError(t, err)

	manager, err := NewManager(store, codec, opts...)
	require.NoError(t, err)

	return manager
}

func testActiveKey() Key {
	return Key{
		ID:     "active",
		Secret: []byte("0123456789abcdef0123456789abcdef"),
	}
}

func testFallbackKey() Key {
	return Key{
		ID:     "fallback",
		Secret: []byte("fedcba9876543210fedcba9876543210"),
	}
}

func testClock(now *time.Time) Option {
	return WithClock(func() time.Time {
		return *now
	})
}

func testIDSequence(ids ...string) Option {
	index := 0
	return WithIDGenerator(func() (string, error) {
		if index >= len(ids) {
			return "", errors.New("no more ids")
		}
		id := ids[index]
		index++
		return id, nil
	})
}
