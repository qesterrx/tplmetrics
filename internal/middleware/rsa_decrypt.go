package middleware

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"

	"github.com/qesterrx/tplmetrics/pkg/aes"
)

func RSADecrypt(privateKeyPem *rsa.PrivateKey) func(http.Handler) http.Handler {

	return func(h http.Handler) http.Handler {
		funcRSADecrypt := func(w http.ResponseWriter, r *http.Request) {

			contentLength := r.Header.Get("Content-Length")
			encryptedAESKeyString := r.Header.Get("AES")

			if contentLength != "0" && contentLength != "" && encryptedAESKeyString != "" {

				encryptedAESKey, err := base64.StdEncoding.DecodeString(encryptedAESKeyString)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				//Расшифровываем секретным ключем
				decryptedAESKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKeyPem, encryptedAESKey, []byte{})
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				//Чиатем тело
				body, err := io.ReadAll(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				r.Body.Close()

				//Расшифровываем тело
				decryptedBody, err := aes.DecryptGCM(body, decryptedAESKey)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				//Восстанавливаем для дальнейшей работы
				r.Body = io.NopCloser(bytes.NewBuffer(decryptedBody))

			}

			h.ServeHTTP(w, r)
		}

		return http.HandlerFunc(funcRSADecrypt)

	}

}
