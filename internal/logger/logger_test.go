package logger

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestLoggingMiddleware(t *testing.T) {

	var buf bytes.Buffer
	log := zerolog.New(&buf).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	Log = log

	// Тестовый обработчик
	handler := LoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Выполняем запрос
	handler.ServeHTTP(w, req)

	logString := buf.String()
	assert.Contains(t, logString, `"level":"info"`)
	assert.Contains(t, logString, `"URI":"/"`)
	assert.Contains(t, logString, `"method":"GET"`)
	assert.Contains(t, logString, `"duration":"`)
	assert.Contains(t, logString, `"code":200`)
	assert.Contains(t, logString, `"size request":0`)
	assert.Contains(t, logString, `"size response":2`)

}
