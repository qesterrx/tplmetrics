package middleware

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIPRequest(t *testing.T) {
	tests := []struct {
		name           string
		subnet         string
		remoteAddr     string
		headers        map[string]string
		expectedStatus int
		expectedIP     string
		expectInCtx    bool
	}{
		{
			name:           "Valid IP without subnet",
			subnet:         "",
			remoteAddr:     "192.168.1.100:54321",
			headers:        nil,
			expectedStatus: http.StatusOK,
			expectedIP:     "192.168.1.100",
			expectInCtx:    true,
		},
		{
			name:           "Valid IP in subnet",
			subnet:         "192.168.1.0/24",
			remoteAddr:     "192.168.1.100:54321",
			headers:        nil,
			expectedStatus: http.StatusOK,
			expectedIP:     "192.168.1.100",
			expectInCtx:    true,
		},
		{
			name:           "Invalid IP - outside subnet",
			subnet:         "192.168.1.0/24",
			remoteAddr:     "192.168.2.100:54321",
			headers:        nil,
			expectedStatus: http.StatusForbidden,
			expectedIP:     "",
			expectInCtx:    false,
		},
		{
			name:           "Invalid RemoteAddr format",
			subnet:         "",
			remoteAddr:     "invalid-address",
			headers:        nil,
			expectedStatus: http.StatusBadRequest,
			expectedIP:     "",
			expectInCtx:    false,
		},
		{
			name:           "RemoteAddr without port",
			subnet:         "",
			remoteAddr:     "192.168.1.100",
			headers:        nil,
			expectedStatus: http.StatusBadRequest,
			expectedIP:     "",
			expectInCtx:    false,
		},
		{
			name:       "X-Real-IP header",
			subnet:     "",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Real-IP": "203.0.113.5",
			},
			expectedStatus: http.StatusOK,
			expectedIP:     "203.0.113.5",
			expectInCtx:    true,
		},
		{
			name:       "X-Real-IP with multiple IPs (take first)",
			subnet:     "",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Real-IP": "203.0.113.5, 192.168.1.10",
			},
			expectedStatus: http.StatusOK,
			expectedIP:     "203.0.113.5",
			expectInCtx:    true,
		},
		{
			name:       "Both X-Real-IP and X-Forwarded-For - BadRequest",
			subnet:     "",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Real-IP":       "203.0.113.5",
				"X-Forwarded-For": "203.0.113.5",
			},
			expectedStatus: http.StatusBadRequest,
			expectedIP:     "",
			expectInCtx:    false,
		},
		{
			name:       "Invalid IP in X-Real-IP",
			subnet:     "",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Real-IP": "invalid-ip",
			},
			expectedStatus: http.StatusBadRequest,
			expectedIP:     "",
			expectInCtx:    false,
		},
		{
			name:       "IP in X-Real-IP outside subnet",
			subnet:     "192.168.1.0/24",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Real-IP": "192.168.2.100",
			},
			expectedStatus: http.StatusForbidden,
			expectedIP:     "",
			expectInCtx:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем подсеть, если указана
			var subnet *net.IPNet
			if tt.subnet != "" {
				_, parsed, err := net.ParseCIDR(tt.subnet)
				assert.NoError(t, err)
				subnet = parsed
			}

			// Создаем middleware
			handler := IPRequest(subnet)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Проверяем, что IP есть в контексте
				ip, ok := r.Context().Value(ContextIP).(string)
				if tt.expectInCtx {
					assert.True(t, ok)
					assert.Equal(t, tt.expectedIP, ip)
				}
				w.WriteHeader(http.StatusOK)
			}))

			// Создаем запрос
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr

			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			// Выполняем запрос
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			// Проверяем статус
			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}
