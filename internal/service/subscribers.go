package service

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/qesterrx/tplmetrics/internal/logger"
)

// UpdateMetricaSubscriberFile - Подписчик сохраняющий данные в файл, использует модель [UpdateMetricaSubscriberMsg]
type UpdateMetricaSubscriberFile struct {

	//FileName - имя файла для сохранения данных
	FileName string
}

// PushNotify - метод для добавления аудита в файл
func (umsf *UpdateMetricaSubscriberFile) PushNotify(msg []byte) {

	file, err := os.OpenFile(umsf.FileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Log.Error().Msg("UpdateMetricaSubscriberFile: Ошибка открытия файла для записи")
		return
	}
	defer file.Close()

	_, err = file.Write(msg)
	if err != nil {
		logger.Log.Error().Msg("UpdateMetricaSubscriberFile: Ошибка записи данных в файл")
		return
	}

}

// UpdateMetricaSubscriberClient - Подписчик отправляющий даные на URL, использует модель [UpdateMetricaSubscriberMsg]
type UpdateMetricaSubscriberClient struct {
	client *resty.Client

	//URL - адрес для отправки данных
	URL string
}

// PushNotify - метод для отправки аудита на указанный URL
func (umsc *UpdateMetricaSubscriberClient) PushNotify(msg []byte) {
	if umsc.client == nil {
		umsc.client = resty.New()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	resp, err := umsc.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(msg).
		Post(umsc.URL)

	if err != nil {
		logger.Log.Error().Msg("UpdateMetricaSubscriberClient: Ошибка при отправке данных по подписчику " + umsc.URL)
		return
	}

	if resp.StatusCode() != http.StatusOK {
		logger.Log.Error().Msg("UpdateMetricaSubscriberClient: Подписчик ответил ошибкой")
		return
	}

}
