package agent

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/qesterrx/tplmetrics/internal/logger"
	"github.com/qesterrx/tplmetrics/pkg/retry"
)

// retryableErrors - Список ошибок для переотправки, проверяемый через checkRetryRequest функцию
var retryableErrors = []error{
	syscall.ECONNREFUSED, //Соединение отклонено
	syscall.ECONNRESET,   //Соединение сброшено
	syscall.ETIMEDOUT,    //Таймаут операции
	syscall.EHOSTUNREACH, //Хост недоступен
	syscall.ENETUNREACH,  //Сеть недоступна
	syscall.EPIPE,        //Разорванный канал
	syscall.EAGAIN,       //Ресурс временно недоступен
	syscall.EWOULDBLOCK,  //Ресурс временно недоступен
}

// checkRetryRequest- Функция проверки ошибки на необходимость повтора
func checkRetryRequest(err error) bool {

	for _, v := range retryableErrors {
		if errors.Is(err, v) {
			//Повторяемая ошибка
			return true
		}
	}
	return false
}

// Sender - Процедура (горутина) которая вычитывает данные из очереди toSend и отправляет их по url
// В метод дополнительно передан параметр secretKeyForSign, содаржащий ключ для шифрования сообщения через функцию HMACSignMiddleware
//
// В данной процедуре используется resty клиент, и дополнительно заданы middleware функции
// GzipCompressMiddleware
// HMACSignMiddleware
func Sender(ctx context.Context, toSend <-chan []byte, num int, url string, secretKeyForSign string, PublicKeyRSA *x509.Certificate) {

	//??? вот тут с-порно, если бы этот контекст использовался бы Resty то логично, а тут он прокидывается в Retry и по факту используется там только для того чтобы отменить retry
	//ctxSend := context.WithoutCancel(ctx)

	logger.Log.Debug().Msg("Запуск Sender" + strconv.Itoa(num))
	//Клиента создаем один раз
	client := resty.New()

	//Тут определим middleware агента
	client.OnBeforeRequest(GzipCompressMiddleware) //Сначала зипуем
	//TODO вообще если не зиповать то выходит ошибка RSA crypto/rsa: message too long for RSA key size
	if PublicKeyRSA != nil {
		client.OnBeforeRequest(RSAEncrypt(PublicKeyRSA)) //Затем RSA
	}
	if secretKeyForSign != "" {
		client.OnBeforeRequest(HMACSignMiddleware(secretKeyForSign)) //Затем подписываем
	}

	for {
		select {
		case <-ctx.Done():
			//Если получили сигнал завершения останавливаемся
			logger.Log.Debug().Msg("Остановка Sender " + strconv.Itoa(num) + " по контексту, ожидается завершение процесса")

			//Reporter может еще что-то дописывать в канал, ждем закрытия канала
			for msg := range toSend {
				err := Send(ctx, client, url, msg)

				if err != nil {
					logger.Log.Error().Msg(err.Error())
				} else {
					logger.Log.Debug().Msg("Метрики отправлены [Sender " + strconv.Itoa(num) + "]")
				}
			}

			logger.Log.Debug().Msg("Sender " + strconv.Itoa(num) + " завершил работу, остановлен по контексту")
			return

		case msg := <-toSend:
			err := Send(ctx, client, url, msg)

			if err != nil {
				logger.Log.Error().Msg(err.Error())
			} else {
				logger.Log.Debug().Msg("Метрики отправлены [Sender " + strconv.Itoa(num) + "]")
			}
		}
	}

}

// Send - Процедура отвечает за отправку сериализованных данных
func Send(ctx context.Context, client *resty.Client, url string, body []byte) error {

	//замыкание для вызова в retry.RetryFunc
	fn := func() error {

		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(body).
			Post(url)

		if err != nil {
			//Получили ошибку при выполнении запроса
			return fmt.Errorf("sender ошибка выполнения запроса на сервер %w", err)
		}

		if resp.StatusCode() != http.StatusOK {
			//Получили от сервера код который не ожидали
			return fmt.Errorf("sender сервер не принял сообщение StatusCode!=OK")
		}

		return nil
	}

	//Выполнение с повтором
	return retry.RetryFunc(ctx, fn, checkRetryRequest, 3, 1*time.Second, 2*time.Second)

}
