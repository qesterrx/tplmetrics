package repository

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/qesterrx/tplmetrics/internal/logger"
	"github.com/qesterrx/tplmetrics/internal/model"
)

// Интерфейс используемый в хендлерах
type MetricaStorage interface {
	//Метод для обновления данных метрики
	UpdateMetrica(model.Metrica) error
	//Метод получения экземпляра метрики по имени
	Metrica(name string, kind string) (model.Metrica, error)
	//Получение всех метрик
	AllMetrics() []model.Metrica
	//Отладочный вызов
	Debug()
	//Очередь уведомлений по обновлениям данных для синхронной записи в файл
	GetQueueUpdateEvents() *chan bool
}

// "Синхронная" запись в файл
// На самом деле конечно не синхронная, если это будет проблемой прошу направить меня в сторону решения.
// Записывать напрямую в структуре которая реализует интерфейс MetricaStorage мне показалось не правильным с точки зрения декомпозиции
func EventSaver(ctx context.Context, filename string, storage MetricaStorage) {

	//Получаем очередь из которой будем читать
	queueEvent := *storage.GetQueueUpdateEvents()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Debug().Msg("Завершение EventSaver по контексту")
			return
		case <-queueEvent:
			err := SaveDataToFile(filename, storage)
			if err != nil {
				logger.Log.Error().Msg("EventSaver Ошибка при сохранении данных в файл" + err.Error())
			} else {
				logger.Log.Debug().Msg("EventSaver Данные сохранены в файл")
			}
		}
	}
}

// Переодическая запись текущего состояния в файл
func TimeSaver(ctx context.Context, filename string, interval int, storage MetricaStorage) {

	ticker := time.NewTicker(time.Second * time.Duration(interval))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Debug().Msg("Завершение TimeSaver по контексту")
			return
		case <-ticker.C:
			err := SaveDataToFile(filename, storage)
			if err != nil {
				logger.Log.Error().Msg("TimeSaver Ошибка при сохранении данных в файл" + err.Error())
			} else {
				logger.Log.Debug().Msg(" TimeSaver Данные сохранены в файл")
			}
		}
	}
}

// Функция сохранения данных в файл
func SaveDataToFile(filename string, storage MetricaStorage) error {
	mtrks := storage.AllMetrics()

	bytes, err := json.Marshal(&mtrks)
	if err != nil {
		return err
	}

	err = os.WriteFile(filename, bytes, 0666)
	if err != nil {
		return err
	}

	return nil
}

// Функция чтения сохраненных данных из файла в момент запуска сервера
func LoadDataFromFile(filename string, storage MetricaStorage) error {

	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	arrMetrica := []model.MetricaJSONAdapter{}
	err = json.Unmarshal(data, &arrMetrica)
	if err != nil {
		return err
	}

	for _, v := range arrMetrica {
		mtrk, err := v.Metrica()
		if err != nil {
			return err
		}
		storage.UpdateMetrica(mtrk)
	}

	return nil
}
