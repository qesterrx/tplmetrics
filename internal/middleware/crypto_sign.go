package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"
)

type HMACSignWriter struct {
	// Райтер, встраивание
	http.ResponseWriter
	key []byte
}

func calcHash(msg []byte, key []byte) string {
	hash := hmac.New(sha256.New, key)
	hash.Write(msg)
	sign := hash.Sum(nil)
	return base64.StdEncoding.EncodeToString(sign)
}

func (sgnw *HMACSignWriter) Write(msg []byte) (int, error) {

	hash := calcHash(msg, sgnw.key)
	sgnw.ResponseWriter.Header().Set("HashSHA256", hash)
	return sgnw.ResponseWriter.Write(msg)
}

// HMACSignMiddleware - Middleware-фунция проверяющая хеш тела http запроса переданного в заголовке HashSHA256
func HMACSignMiddleware(key string) func(http.Handler) http.Handler {

	secret := []byte(key)

	return func(h http.Handler) http.Handler {
		funcSign := func(w http.ResponseWriter, r *http.Request) {

			//Райтер, если мы сюда попали значит все ответы должны содержать подпись
			sgnw := HMACSignWriter{ResponseWriter: w, key: secret}

			//Запрос проверим без отдельной структуры
			hashIncome := r.Header.Get("HashSHA256")
			if hashIncome != "" {
				//Читаем запрос
				body, err := io.ReadAll(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				r.Body.Close() //Вроде как всегда закрывать надо

				//Проверяем хеш
				hash := calcHash(body, secret)
				if hash != hashIncome {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				//Восстанавливаем тело для дальнейшей работы
				r.Body = io.NopCloser(bytes.NewBuffer(body))
			}

			h.ServeHTTP(&sgnw, r)

		}

		return http.HandlerFunc(funcSign)
	}
}
