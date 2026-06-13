package middleware

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
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

	testCases := []struct {
		name          string
		message       []byte
		expectedError bool
	}{
		{
			name:          "Valid encryption",
			message:       []byte(`{"username": "testuser", "password": "testpass"}`),
			expectedError: false,
		},
		{
			name:          "Empty message",
			message:       []byte(``),
			expectedError: false,
		},
		{
			name:          "Large message",
			message:       bytes.Repeat([]byte("A"), 100),
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Шифруем сообщение
			ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, tc.message, []byte{})
			if err != nil && !tc.expectedError {
				t.Fatalf("Failed to encrypt: %v", err)
			}

			if tc.expectedError {
				return
			}

			// Создаем тестовый обработчик
			var receivedBody []byte
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				receivedBody = body
				w.WriteHeader(http.StatusOK)
			})

			// Применяем middleware
			middleware := RSADecrypt(privateKey)
			handler := middleware(nextHandler)

			// Создаем запрос
			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(ciphertext))
			len := len(ciphertext)
			req.Header.Set("Content-Length", strconv.Itoa(len))
			rec := httptest.NewRecorder()

			// Выполняем запрос
			handler.ServeHTTP(rec, req)

			// Проверяем результат
			if rec.Code != http.StatusOK {
				t.Errorf("Expected status OK, got %d", rec.Code)
			}

			if !bytes.Equal(receivedBody, tc.message) {
				t.Errorf("Body mismatch.\nExpected: %s\nGot: %s", tc.message, receivedBody)
			}
		})
	}
}
