package middleware

import (
	"context"
	"net/http"
)

const ContextIP string = "ContextIP"

// Простейший вариант, если надо, можно усложнить анализом X-FORWARDED-FROM
func IPContext(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), ContextIP, r.RemoteAddr)
		h.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
