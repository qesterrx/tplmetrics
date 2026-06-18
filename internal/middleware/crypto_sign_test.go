package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHMACSignMiddleware(t *testing.T) {

	secret := "!ToPSecretKey31415"

	res := "A not so long message for response"
	resSign := calcHash([]byte(res), []byte(secret))

	//Получим middleware
	middlFunc := HMACSignMiddleware(secret)

	// Тестовый обработчик
	handler := middlFunc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(res))
	}))

	//Проверяем что сервер умеет подписывать сообщения
	t.Run("Response", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		// Выполняем запрос
		handler.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, resSign, w.Header().Get("HashSHA256"))
	})

	req := "A not so long message for request"
	reqSign := calcHash([]byte(req), []byte(secret))

	//Проверяем что сервер умеет проверять подписанные сообщения
	t.Run("Request correct sign", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", strings.NewReader(req))
		req.Header.Set("HashSHA256", reqSign)
		w := httptest.NewRecorder()

		// Выполняем запрос
		handler.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, resSign, w.Header().Get("HashSHA256"))
	})

	t.Run("Request incorrect sign", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", strings.NewReader(req))
		req.Header.Set("HashSHA256", "ABC")
		w := httptest.NewRecorder()

		// Выполняем запрос
		handler.ServeHTTP(w, req)

		res := w.Result()
		defer res.Body.Close()

		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})

}
