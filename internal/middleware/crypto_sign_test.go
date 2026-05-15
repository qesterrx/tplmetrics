package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHMACSignMiddleware(t *testing.T) {

	msg := "A not so long message for response"
	secret := "!ToPSecretKey31415"

	//Получим ожидаемый результат
	hash := hmac.New(sha256.New, []byte(secret))
	hash.Write([]byte(msg))
	sign := hash.Sum(nil)
	sign64 := base64.StdEncoding.EncodeToString(sign)

	//Получим middleware
	middlFunc := HMACSignMiddleware(secret)

	// Тестовый обработчик
	handler := middlFunc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(msg))
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Выполняем запрос
	handler.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, sign64, w.Header().Get("HashSHA256"))

}

//TODO набросок. Тут надо проверять не только ответ от сервера, но и проверку подписи в запросе
