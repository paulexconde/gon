package static

import (
	"net/http"
	"strings"

	"github.com/paulexconde/gon/gon"
)

func Static(staticDir string) gon.Middleware {
	fileServer := http.FileServer(http.Dir(staticDir))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			if strings.HasPrefix(path, "/static/") {
				http.StripPrefix("/static/", fileServer).ServeHTTP(w, r)
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}
