//Я еще буду курить это, сдаю на проверку чтобы узнать правильный подход или нет
//fmt тут пока тоже подержу
//прошу понять и простить

package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// EncryptGCM шифрует данные с использованием AES-256-GCM
func EncryptGCM(source []byte, key []byte) ([]byte, error) {

	//fmt.Println("EncryptGCM", "source", source)
	//fmt.Println("EncryptGCM", "key", key)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	encoded := gcm.Seal(nonce, nonce, source, nil) //сделал тут nil в dst и огреб разборку на полтора часа

	//fmt.Println("EncryptGCM", "encoded", encoded)

	return encoded, nil
}

// DecryptGCM расшифровывает данные
func DecryptGCM(source []byte, key []byte) ([]byte, error) {

	//fmt.Println("DecryptGCM", "source", source)
	//fmt.Println("DecryptGCM", "key", key)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(source) < nonceSize {
		return nil, fmt.Errorf("Сообщение недостаточной длины")
	}

	// Извлекаем nonce и зашифрованные данные
	nonce, encodedMsg := source[:nonceSize], source[nonceSize:]

	// Расшифровываем и проверяем аутентификацию
	decoded, err := gcm.Open(nil, nonce, encodedMsg, nil) //на удивление тут нужен именно нул в dst
	if err != nil {
		return nil, err
	}

	//fmt.Println("DecryptGCM", "decoded", decoded)

	return decoded, nil
}

// GenAESKey возвращает случайный ключ нужной длинны
func GenAESKey() ([]byte, error) {
	AESKeyBytes := make([]byte, 32)
	_, err := rand.Read(AESKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("Ошибка получения AES ключа %w", err)
	}

	return AESKeyBytes, nil
}
