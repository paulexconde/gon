package cache

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"sync"
	"time"
)

type CachedResponse struct {
	ContentType string
	Data        []byte
}

type Cache interface {
	Set(key string, value CachedResponse, duration time.Duration) error
	Get(key string) (CachedResponse, bool)
	Delete(key string) error
	Clear() error
}

type MemoryCache struct {
	expiration map[string]time.Time
	cache      map[string]CachedResponse
	mutex      sync.Mutex
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		cache:      make(map[string]CachedResponse),
		expiration: make(map[string]time.Time),
	}
}

func (m *MemoryCache) Set(key string, value CachedResponse, duration time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.cache[key] = value

	if duration > 0 {
		m.expiration[key] = time.Now().Add(duration)
	}

	return nil
}

func (m *MemoryCache) Get(key string) (CachedResponse, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if exp, found := m.expiration[key]; found && time.Now().After(exp) {
		delete(m.cache, key)
		delete(m.expiration, key)
		return CachedResponse{}, false
	}

	val, found := m.cache[key]
	return val, found
}

func (m *MemoryCache) Delete(key string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.cache, key)
	delete(m.expiration, key)
	return nil
}

func (m *MemoryCache) Clear() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.cache = make(map[string]CachedResponse)
	m.expiration = make(map[string]time.Time)

	return nil
}

type FileCache struct {
	baseDir string
}

func NewFileCache(baseDir string) *FileCache {
	return &FileCache{
		baseDir: baseDir,
	}
}

func (f *FileCache) Set(key string, value CachedResponse, duration time.Duration) error {
	filePath := path.Join(f.baseDir, key)
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

func (f *FileCache) Get(key string) (CachedResponse, bool) {
	filePath := path.Join(f.baseDir, key)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return CachedResponse{}, false
	}

	var value CachedResponse

	if err = json.Unmarshal(data, &value); err != nil {
		return CachedResponse{}, false
	}

	return value, true
}

func (f *FileCache) Delete(key string) error {
	filePath := path.Join(f.baseDir, key)

	return os.Remove(filePath)
}

func (f *FileCache) Clear() error {
	return os.RemoveAll(f.baseDir)
}

func CacheMiddleware(cache Cache) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.RequestURI()

			if cachedResponse, found := cache.Get(key); found {
				w.Header().Set("Content-Type", cachedResponse.ContentType)
				w.Write(cachedResponse.Data)
				return
			}

			rec := httptest.NewRecorder()
			next.ServeHTTP(rec, r)

			cachedResponse := CachedResponse{
				ContentType: rec.Header().Get("Content-Type"),
				Data:        rec.Body.Bytes(),
			}

			cache.Set(key, cachedResponse, 10*time.Minute)

			w.Header().Set("Content-Type", cachedResponse.ContentType)

			for k, v := range rec.Header() {
				w.Header()[k] = v
			}
			w.WriteHeader(rec.Code)
			w.Write(cachedResponse.Data)
		})
	}
}
