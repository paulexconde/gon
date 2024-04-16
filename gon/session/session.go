package session

import (
	"context"
	"net/http"
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
	defer store.lock.Unlock()
	store.Sessions[sessionID] = session

	return session
}

func (store *SessionStore) GetSession(sessionID string) (*Session, bool) {
	// Fix mutex
	store.lock.RLock()
	defer store.lock.RUnlock()
	session, exists := store.Sessions[sessionID]

	return session, exists
}

func (store *SessionStore) DeleteSession(sessionID string) {
	store.lock.Lock()
	defer store.lock.Unlock()
	delete(store.Sessions, sessionID)
}

func SessionMiddleware(store *SessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionCookie, err := r.Cookie("session_id")

			if err == http.ErrNoCookie {
				// no session, handlers are remain functional even without a session
			} else {
				sess, _ := store.GetSession(sessionCookie.Value)

				if sess != nil && sess.ExpiresAt.After(time.Now()) {
					ctx := context.WithValue(r.Context(), Key, sess)
					r = r.WithContext(ctx)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
