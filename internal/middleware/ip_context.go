package middleware

import (
	"context"
	"net/http"
	"strings"
)

// ContextValue тип для внедрения занчений в контекст
type ContextValue string

// ContextIP - Имя для параметра IP клиента, записанного в контекст запроса
const ContextIP ContextValue = "ContextIP"

// IPContext - Middleware-фунция обогащения контекста запроса значением IP клиента
func IPContext(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		//Добавлем анализ на прокидывание запроса через промежуточные хосты
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ips := strings.Split(forwarded, ",")
			if len(ips) > 0 {
				ip = ips[0]
			}
		}

		ctx := context.WithValue(r.Context(), ContextIP, strings.TrimSpace(ip))
		h.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
