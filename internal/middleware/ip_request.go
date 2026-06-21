package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"
)

// ContextValue тип для внедрения занчений в контекст
type contextIpValue string

// ContextIP - Имя для параметра IP клиента, записанного в контекст запроса
const ContextIP contextIpValue = "ContextIP"

// IPRequest - Middleware-фунция проверки ип адреса по маске и обогащения контекста запроса значением IP клиента
func IPRequest(subnet *net.IPNet) func(http.Handler) http.Handler {

	return func(h http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr

			//Обрезаем порт
			ip, _, err := net.SplitHostPort(ip)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			//Добавлем анализ на прокидывание запроса через промежуточные хосты
			hasForwarded := false
			if forwarded := r.Header.Get("X-Real-IP"); forwarded != "" {
				ips := strings.Split(forwarded, ",")
				if len(ips) > 0 {
					hasForwarded = true
					ip = ips[0]
				}
			}

			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				ips := strings.Split(forwarded, ",")
				if len(ips) > 0 {
					//Непонятно на какой заголовок ориентироваться
					if hasForwarded {
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					ip = ips[0]
				}
			}

			//Проверим что выбранная строка это ИП
			netIP := net.ParseIP(ip)
			if netIP == nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			//Проверка адреса по маске
			if subnet != nil {
				if !subnet.Contains(netIP) {
					w.WriteHeader(http.StatusForbidden)
					return
				}
			}

			//Добавляем ИП в контекст для дальнейших нужд
			ctx := context.WithValue(r.Context(), ContextIP, strings.TrimSpace(ip))
			h.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}
