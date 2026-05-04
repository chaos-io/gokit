package session

import (
	"context"
	"sync"
	"time"
)

type MemoryStoreOption func(*MemoryStore)

type MemoryStore struct {
	mu                   sync.RWMutex
	sessions             map[string]*Session
	activeSessionIDs     map[subjectKey]map[string]struct{}
	singleSessionPerUser bool
}

type subjectKey struct {
	userID string
	appID  int32
}

func WithSingleSessionPerUser() MemoryStoreOption {
	return func(store *MemoryStore) {
		store.singleSessionPerUser = true
	}
}

func NewMemoryStore(opts ...MemoryStoreOption) *MemoryStore {
	store := &MemoryStore{
		sessions:         make(map[string]*Session),
		activeSessionIDs: make(map[subjectKey]map[string]struct{}),
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(store)
	}

	return store
}

func (s *MemoryStore) Save(ctx context.Context, session *Session) error {
	_ = ctx

	cloned := cloneSession(session)
	if cloned == nil {
		return ErrSessionInvalid
	}
	if cloned.ID == "" {
		return ErrSessionIDRequired
	}
	if cloned.Subject.UserID == "" {
		return ErrSessionUserIDRequired
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.singleSessionPerUser {
		for sessionID := range s.activeSessionIDs[newSubjectKey(cloned.Subject)] {
			s.revokeLocked(sessionID, cloned.IssuedAt)
		}
	}

	if existing, ok := s.sessions[cloned.ID]; ok {
		s.removeSubjectIndex(existing)
	}

	s.sessions[cloned.ID] = cloned
	s.addSubjectIndex(cloned)

	return nil
}

func (s *MemoryStore) Find(ctx context.Context, sessionID string) (*Session, error) {
	_ = ctx

	if sessionID == "" {
		return nil, ErrSessionIDRequired
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[sessionID]
	if !ok {
		return nil, ErrSessionNotFound
	}

	return cloneSession(session), nil
}

func (s *MemoryStore) Revoke(ctx context.Context, sessionID string, revokedAt time.Time) error {
	_ = ctx

	if sessionID == "" {
		return ErrSessionIDRequired
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.sessions[sessionID]; !ok {
		return ErrSessionNotFound
	}

	s.revokeLocked(sessionID, revokedAt)
	return nil
}

func (s *MemoryStore) addSubjectIndex(session *Session) {
	key := newSubjectKey(session.Subject)
	if _, ok := s.activeSessionIDs[key]; !ok {
		s.activeSessionIDs[key] = make(map[string]struct{})
	}
	s.activeSessionIDs[key][session.ID] = struct{}{}
}

func (s *MemoryStore) removeSubjectIndex(session *Session) {
	key := newSubjectKey(session.Subject)
	index, ok := s.activeSessionIDs[key]
	if !ok {
		return
	}

	delete(index, session.ID)
	if len(index) == 0 {
		delete(s.activeSessionIDs, key)
	}
}

func (s *MemoryStore) revokeLocked(sessionID string, revokedAt time.Time) {
	session, ok := s.sessions[sessionID]
	if !ok {
		return
	}

	session.RevokedAt = revokedAt
	s.removeSubjectIndex(session)
}

func newSubjectKey(subject Subject) subjectKey {
	return subjectKey{
		userID: subject.UserID,
		appID:  subject.AppID,
	}
}
