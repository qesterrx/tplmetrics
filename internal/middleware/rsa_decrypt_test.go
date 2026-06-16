package middleware

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/qesterrx/tplmetrics/pkg/aes"
	"github.com/stretchr/testify/assert"
)

// Генерация тестовых ключей RSA
func generateTestKeys(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}
	return privateKey, &privateKey.PublicKey
}

func TestRSADecrypt(t *testing.T) {
	// Генерируем тестовые ключи
	privateKey, publicKey := generateTestKeys(t)

	//Получаем случайный набор байт - ключ для AES
	AESKey, err := aes.GenAESKey()
	assert.NoError(t, err)

	//Шифруем ключ AES публичным ключем RSA
	encryptedKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, AESKey, []byte{})
	assert.NoError(t, err)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Write(body)
		w.WriteHeader(http.StatusOK)
	})

	testCases := []struct {
		name          string
		body          []byte
		expectedError bool
	}{
		{
			name: "Valid encryption",
			body: []byte(`{"username": "testuser", "password": "testpass"}`),
		},
		{
			name: "Empty message",
			body: []byte(``),
		},
		{
			name: "Large message",
			body: []byte(strings.Repeat(`{"test": "data"}`, 1024*1024)),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			// Шифруем сообщение
			encryptedBody, err := aes.EncryptGCM(tc.body, AESKey)
			assert.NoError(t, err)

			// Применяем middleware
			middleware := RSADecrypt(privateKey)
			handlerWM := middleware(handler)

			// Создаем запрос
			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(encryptedBody))
			len := len(encryptedBody)
			req.Header.Set("Content-Length", strconv.Itoa(len))
			req.Header.Set("AES", base64.StdEncoding.EncodeToString(encryptedKey))
			rec := httptest.NewRecorder()

			// Выполняем запрос
			handlerWM.ServeHTTP(rec, req)

			// Проверяем результат
			assert.Equal(t, http.StatusOK, rec.Code)

			//В теле будет результат расшифровки
			receivedBody, err := io.ReadAll(rec.Body)
			assert.NoError(t, err)
			assert.Equal(t, tc.body, receivedBody)
		})
	}
}
