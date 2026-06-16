package agent

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzipCompressMiddleware(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		expectError bool
		expectGzip  bool
	}{
		{
			name:        "compress string body",
			body:        `{"test": "data"}`,
			expectError: false,
			expectGzip:  true,
		},
		{
			name:        "compress empty string",
			body:        "",
			expectError: false,
			expectGzip:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			client := resty.New()
			request := client.R()
			request.SetBody(tt.body)

			err := GzipCompressMiddleware(client, request)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			if tt.expectGzip {
				// Проверяем заголовок
				assert.Equal(t, "gzip", request.Header.Get("Content-Encoding"))

				// Проверяем, что тело сжато
				body, ok := request.Body.([]byte)
				assert.True(t, ok, "Body должен быть []byte")
				assert.NotEmpty(t, body)

				// Пробуем распаковать
				reader, err := gzip.NewReader(bytes.NewReader(body))
				require.NoError(t, err)
				defer reader.Close()

				decompressed, err := io.ReadAll(reader)
				require.NoError(t, err)

				// Проверяем, что распакованное содержимое совпадает с исходным
				assert.Equal(t, []byte(tt.body), decompressed)
			} else {
				// Если сжатие не ожидается, заголовок не должен быть установлен
				assert.Empty(t, request.Header.Get("Content-Encoding"))
			}
		})
	}
}

func TestHMACSignMiddleware(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		body        string
		expectError bool
		expectSign  bool
	}{
		{
			name:        "sign string body with valid key",
			key:         "my-secret-key",
			body:        `{"test": "data"}`,
			expectError: false,
			expectSign:  true,
		},
		{
			name:        "sign empty body",
			key:         "secret",
			body:        "",
			expectError: false,
			expectSign:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			client := resty.New()
			request := client.R()
			request.SetBody(tt.body)

			middleware := HMACSignMiddleware(tt.key)
			err := middleware(client, request)

			assert.NoError(t, err)

			if tt.expectSign {
				// Проверяем наличие заголовка
				signHeader := request.Header.Get("HashSHA256")
				assert.NotEmpty(t, signHeader, "HashSHA256 заголовок должен быть установлен")

				// Вычисляем ожидаемую подпись
				hash := hmac.New(sha256.New, []byte(tt.key))
				hash.Write([]byte(tt.body))
				expectedSign := base64.StdEncoding.EncodeToString(hash.Sum(nil))

				assert.Equal(t, expectedSign, signHeader, "Подпись не совпадает с ожидаемой")
			} else {
				assert.Empty(t, request.Header.Get("HashSHA256"), "HashSHA256 заголовок не должен быть установлен")
			}
		})
	}
}

func TestRSAEncrypt(t *testing.T) {

	// Генерируем приватный ключ для тестов
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	//Кейсы
	tests := []struct {
		name        string
		body        string
		expectError bool
	}{
		{
			name:        "encrypt string body",
			body:        `{"test": "data"}`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			client := resty.New()
			request := client.R()
			request.SetBody(tt.body)

			middleware := RSAEncrypt(&privateKey.PublicKey)
			err := middleware(client, request)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// Проверяем, что тело изменилось
			encryptedBody, ok := request.Body.([]byte)
			assert.True(t, ok, "Body должен быть []byte")
			assert.NotEmpty(t, encryptedBody)

			// Зашифрованное тело не должно совпадать с исходным
			assert.NotEqual(t, []byte(tt.body), encryptedBody, "Зашифрованное тело не должно совпадать с исходным")

			//Проверяем что после расшифровки получилось исходное сообщение
			decryptedBody, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, encryptedBody, []byte{})
			assert.NoError(t, err)
			assert.Equal(t, tt.body, string(decryptedBody))
		})
	}
}
