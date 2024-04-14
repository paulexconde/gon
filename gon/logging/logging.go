package logging

import (
	"net/http"
	"strings"
	"time"

	"github.com/paulexconde/gon/gon/logger"
)

type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *wrappedWriter) Write(b []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}

	return w.ResponseWriter.Write(b)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		if isWebSocketUpgrade(r) {
			// Handle WebSocket separately
			next.ServeHTTP(w, r)
			// Log WebSocket upgrade - not interfering with headers or status code
			logger.INFO("WebSocket upgrade requested", r.URL.Path, time.Since(start))
			return
		}

		wrapped := &wrappedWriter{
			ResponseWriter: w,
		}

		next.ServeHTTP(wrapped, r)

		statusCode := wrapped.statusCode

		if statusCode == 0 {
			statusCode = http.StatusOK
		}
		//	log.Printf("%s %s %d", r.Method, r.URL.Path, statusCode)
		if wrapped.statusCode >= 200 && wrapped.statusCode < 300 {
			logger.LOG(wrapped.statusCode, r.URL.Path, r.URL.Query().Encode(), time.Since(start))
		} else if wrapped.statusCode >= 400 && wrapped.statusCode < 500 {
			logger.BAD(wrapped.statusCode, r.URL.Path, r.URL.Query().Encode(), time.Since(start))
		} else if wrapped.statusCode >= 500 {
			logger.ERROR(wrapped.statusCode, r.URL.Path, r.URL.Query().Encode(), time.Since(start))
		} else {
			logger.WARN(wrapped.statusCode, r.URL.Path, r.URL.Query().Encode(), time.Since(start))
		}
	})
}

func isWebSocketUpgrade(r *http.Request) bool {
	connHdr := ""
	connHdrs := r.Header["Connection"]
	if len(connHdrs) > 0 {
		connHdr = connHdrs[0]
	}

	upgradeWebsocket := false
	upgradeHdrs := r.Header["Upgrade"]
	if len(upgradeHdrs) > 0 {
		upgradeWebsocket = (strings.ToLower(upgradeHdrs[0]) == "websocket")
	}

	return connHdr == "Upgrade" && upgradeWebsocket
}
