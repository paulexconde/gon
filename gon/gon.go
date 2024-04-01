package gon

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/paulexconde/gon/gon/logging"
)

type HttpError struct {
	Message string
	Status  int
}

func (he HttpError) Error() string {
	return fmt.Sprintf("Status: %d, Message: %s", he.Status, he.Message)
}

type (
	Middleware  func(http.Handler) http.Handler
	HandlerFunc func(http.ResponseWriter, *http.Request) error
)

type Gon struct {
	mux         *http.ServeMux
	middlewares []Middleware
}

func NewApp() *Gon {
	return &Gon{
		mux:         http.NewServeMux(),
		middlewares: []Middleware{},
	}
}

func (g *Gon) Use(m Middleware) {
	g.middlewares = append(g.middlewares, m)
}

func (g *Gon) Handle(pattern string, handler HandlerFunc) {
	wrappedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r)
		if err != nil {
			var httpErr HttpError

			if ok := errors.As(err, &httpErr); ok {
				http.Error(w, httpErr.Message, httpErr.Status)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	finalHandler := http.Handler(wrappedHandler)

	for _, m := range g.middlewares {
		finalHandler = m(finalHandler)
	}

	g.mux.Handle(pattern, finalHandler)
}

func (g *Gon) Group(prefix string) *Gon {
	group := &Gon{
		mux:         http.NewServeMux(),
		middlewares: make([]Middleware, len(g.middlewares)),
	}

	copy(group.middlewares, g.middlewares)

	g.mux.Handle(prefix+"/", http.StripPrefix(prefix, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		group.mux.ServeHTTP(w, r)
	})))

	return group
}

func (g *Gon) Start(address string) error {
	server := &http.Server{
		Addr:    address,
		Handler: logging.Logging(g.mux),
	}
	return server.ListenAndServe()
}
