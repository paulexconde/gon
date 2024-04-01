package render

import (
	"encoding/json"
	"net/http"

	"github.com/a-h/templ"
)

func renderComponent(w http.ResponseWriter, r *http.Request, component templ.Component, options ...func(*templ.ComponentHandler)) error {
	handler := templ.Handler(component)

	for _, option := range options {
		option(handler)
	}

	handler.ServeHTTP(w, r)

	return nil
}

func RenderTempl(w http.ResponseWriter, r *http.Request, component templ.Component, options ...func(*templ.ComponentHandler)) error {
	return renderComponent(w, r, component, options...)
}

func JSON(w http.ResponseWriter, data interface{}, statusCode int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		return err
	}

	return nil
}
