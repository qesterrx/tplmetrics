package middleware

import (
	"context"
	"net/http"
)

// ContextValue тип для внедрения занчений в контекст
type ContextValue string

// ContextIP - Имя для параметра IP клиента, записанного в контекст запроса
const ContextIP ContextValue = "ContextIP"

// IPContext - Middleware-фунция обогащения контекста запроса значением IP клиента
// TODO: можно усложнить анализом X-FORWARDED-FROM
func IPContext(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), ContextIP, r.RemoteAddr)
		h.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
