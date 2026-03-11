package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func main() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello from inside")
	})

	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ingressPath := r.Header.Get("X-Hass-Ingress-Path")
			if ingressPath != "" {
				http.StripPrefix(ingressPath, next).ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	app := mw(handler)

	req := httptest.NewRequest("GET", "/api/config", nil)
	req.Header.Set("X-Hass-Ingress-Path", "/api/hassio_ingress/123")
	w := httptest.NewRecorder()
	
	app.ServeHTTP(w, req)
	
	fmt.Printf("Status: %d\n", w.Result().StatusCode)
}
