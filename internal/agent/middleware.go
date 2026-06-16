// Пакет agent содержит в себе основные функции для работы клиента по сбору метрик
package agent

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/qesterrx/tplmetrics/pkg/aes"
)

// GzipCompressMiddleware - Middleware процедура обеспечивающая архивацию тела http запроса перед отправкой на сервер
func GzipCompressMiddleware(c *resty.Client, r *resty.Request) error {
	if r.Body != nil {

		//r.Body это интерфейс, очередной type assertion
		var srcBody []byte
		switch tmp := r.Body.(type) {
		case string:
			srcBody = []byte(tmp)
		case []byte:
			srcBody = tmp
		default:
			//Если тело не то что мы предпологали то просто ничего не делаем
			return nil
		}

		var compressed bytes.Buffer

		gzWriter := gzip.NewWriter(&compressed)

		_, err := gzWriter.Write(srcBody)
		if err != nil {
			return fmt.Errorf("sender ошибка компрессии gzip %w", err)
		}

		// Важно! Закрываем writer, чтобы сбросить все данные в буфер - эх время мое время
		gzWriter.Close()

		//Добавляем заголовок, переписываем Body
		r.SetHeader("Content-Encoding", "gzip")
		r.SetBody(compressed.Bytes())

	}

	return nil

}

// HMACSignMiddleware - Middleware процедура шифрующая тело http запроса HMAC алгоритмом с переданным в виде параметра ключем
func HMACSignMiddleware(key string) resty.RequestMiddleware {

	secret := []byte(key)

	return func(c *resty.Client, r *resty.Request) error {
		if r.Body != nil {

			//r.Body это интерфейс - type assertion
			var srcBody []byte
			switch tmp := r.Body.(type) {
			case string:
				srcBody = []byte(tmp)
			case []byte:
				srcBody = tmp
			default:
				//Если тело не то что мы предпологали то просто ничего не делаем
				return nil
			}

			//считаем хешь, по идее нужна общая функция для клиента и сервереа но и таааак сойдет
			hash := hmac.New(sha256.New, secret)
			hash.Write(srcBody)
			sign := hash.Sum(nil)

			//Записываем заголовок
			r.SetHeader("HashSHA256", base64.StdEncoding.EncodeToString(sign))

		}

		return nil
	}

}

// RSAEncrypt - Шифрование с помощью RSA тела запроса
func RSAEncrypt(PublicKeyRSA *rsa.PublicKey) resty.RequestMiddleware {

	return func(c *resty.Client, r *resty.Request) error {
		if r.Body != nil {

			//r.Body это интерфейс, очередной type assertion
			var srcBody []byte
			switch tmp := r.Body.(type) {
			case string:
				srcBody = []byte(tmp)
			case []byte:
				srcBody = tmp
			default:
				//Если тело не то что мы предпологали то просто ничего не делаем
				return nil
			}

			//Получаем случайный набор байт - ключ для AES
			AESKey, err := aes.GenAESKey()
			if err != nil {
				return fmt.Errorf("sender ошибка получения AES ключа")
			}

			//Шифруем ключ AES публичным ключем RSA
			encryptedKey, err := rsa.EncryptOAEP(sha256.New(), cryptorand.Reader, PublicKeyRSA, AESKey, []byte{})
			if err != nil {
				return fmt.Errorf("sender ошибка шифрования AES ключа с помощью RSA %w", err)
			}

			//Шифруем тело через AES
			encryptedBody, err := aes.EncryptGCM(srcBody, AESKey)
			if err != nil {
				return fmt.Errorf("sender ошибка шифрования сообщения с помощью AES %w", err)
			}

			r.Header.Set("AES", base64.StdEncoding.EncodeToString(encryptedKey))
			r.SetBody(encryptedBody)

		}

		return nil

	}

}
