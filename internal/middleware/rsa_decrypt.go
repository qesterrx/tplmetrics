package middleware

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"io"
	"net/http"
)

func RSADecrypt(privateKeyPem *rsa.PrivateKey) func(http.Handler) http.Handler {

	return func(h http.Handler) http.Handler {
		funcRSADecrypt := func(w http.ResponseWriter, r *http.Request) {

			contentLength := r.Header.Get("Content-Length")

			if contentLength != "0" && contentLength != "" {

				//Чиатем тело
				body, err := io.ReadAll(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				r.Body.Close()

				//Расшифровываем секретным ключем
				decryptedBody, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKeyPem, body, []byte{})
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
				}

				//Восстанавливаем для дальнейшей работы
				r.Body = io.NopCloser(bytes.NewBuffer(decryptedBody))

			}

			h.ServeHTTP(w, r)
		}

		return http.HandlerFunc(funcRSADecrypt)

	}

}
