package csrf

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
)

type contextKey string

const CSRFTokenKey contextKey = "csrf_token"

func generateCSRFToken() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(b)
}

func CSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var csrfToken string

		if cookie, err := r.Cookie("csrf_token"); err == nil {
			csrfToken = cookie.Value
		} else {
			csrfToken = generateCSRFToken()

			http.SetCookie(w, &http.Cookie{
				Name:     "csrf_token",
				Value:    csrfToken,
				Path:     "/",
				HttpOnly: true,
			})
		}

		if r.Method == "GET" {

			ctx := context.WithValue(r.Context(), CSRFTokenKey, csrfToken)
			r = r.WithContext(ctx)

		} else {
			formToken := r.FormValue("csrf_token")

			if formToken != csrfToken {
				http.Error(w, "Invalid CSRF Token", http.StatusForbidden)
				return
			}

		}

		next.ServeHTTP(w, r)
	})
}
