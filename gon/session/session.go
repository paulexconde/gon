package session

import (
	"sync"
	"time"
)

type sessionKey string

const Key sessionKey = "session"

type Session struct {
	ExpiresAt time.Time
	Data      map[string]interface{}
	ID        string
}

type SessionStore struct {
	Sessions map[string]*Session
	lock     sync.RWMutex
}

func NewSessionStore() *SessionStore {
	return &SessionStore{
		Sessions: make(map[string]*Session),
	}
}

func (store *SessionStore) CreateSession() *Session {
	sessionID := time.Now().Format("20060102150405")

	session := &Session{
		ID:        sessionID,
		Data:      make(map[string]interface{}),
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}

	store.lock.Lock()
	store.Sessions[sessionID] = session
	store.lock.Unlock()

	return session
}

func (store *SessionStore) GetSession(sessionID string) (*Session, bool) {
	store.lock.RLock()
	session, exists := store.Sessions[sessionID]
	store.lock.Unlock()

	return session, exists
}

func (store *SessionStore) DeleteSession(sessionID string) {
	store.lock.Lock()
	delete(store.Sessions, sessionID)
	store.lock.Unlock()
}
