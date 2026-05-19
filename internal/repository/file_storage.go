// Пакет repository - Содержит логику работы со всеми возможным вариантами хранения данных о метриках
package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/qesterrx/tplmetrics/internal/config"
	"github.com/qesterrx/tplmetrics/internal/logger"
	"github.com/qesterrx/tplmetrics/internal/model"
)

// FileStorage - структура, реализующая интерфейс [service.MetricaStorage]
// обеспечивает сохранение метрик в указанных в настройках файл
// поддерживает асинхронный режим работы
type FileStorage struct {
	*MemStorage
	filename   string
	mode       config.MetricaStorageMode
	restore    bool
	hasChanged bool
}

// NewFileStorage - функция возвращает новый экземпляр FileStorage
// На вход ожидает
// ms - адрес экземпляра MemStorage (для хранения изменений в памяти)
// filename - имя файла для сохранения данных
// mode - режим работы хранилища MetricaStorageMode
// restore - флаг необходимости восстановления данных из указанного файла (при запуске)
func NewFileStorage(ms *MemStorage, filename string, mode config.MetricaStorageMode, restore bool) (*FileStorage, error) {

	logger.Log.Debug().Msg("Создание FileStorage")
	//Проверяем существование файла, если файла нет надо его создать
	_, err := os.Stat(filename)
	if err != nil {
		err := os.WriteFile(filename, []byte(""), 0666)
		if err != nil {
			return nil, fmt.Errorf("ошика создания файла %w", err)
		}
	}

	fs := FileStorage{
		MemStorage: ms,
		mode:       mode,
		hasChanged: false,
		filename:   filename,
		restore:    restore,
	}

	//Если необходимо восстановить данные из файла
	if restore {

		logger.Log.Debug().Msg("Загрузка данных из файла")

		data, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}

		if len(data) > 0 {
			arrMetrica := []model.MetricaJSONAdapter{}
			err = json.Unmarshal(data, &arrMetrica)
			if err != nil {
				return nil, err
			}

			for _, v := range arrMetrica {
				mtrk, err := v.Metrica()
				if err != nil {
					return nil, err
				}
				fs.MemStorage.UpdateMetrica(context.Background(), mtrk)
			}
		}

	}

	return &fs, nil
}

// GetMetrica - возвращает метрику по имени и типу
func (fs *FileStorage) GetMetrica(ctx context.Context, name string, kind string) (model.Metrica, error) {
	return fs.MemStorage.GetMetrica(ctx, name, kind)
}

// UpdateMetrica - обновляет метрику
func (fs *FileStorage) UpdateMetrica(ctx context.Context, mtrk model.Metrica) error {

	err := fs.MemStorage.UpdateMetrica(ctx, mtrk)
	if err != nil {
		return err
	}

	fs.hasChanged = true
	if fs.mode == config.MetricaStorageModeSync {
		err := fs.WriteMetrics(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// UpdateMetricaBatch Обновляет массив метрик
func (fs *FileStorage) UpdateMetricaBatch(ctx context.Context, mtrks []model.Metrica) error {
	err := fs.MemStorage.UpdateMetricaBatch(ctx, mtrks)
	if err != nil {
		return err
	}

	fs.hasChanged = true
	if fs.mode == config.MetricaStorageModeSync {
		err := fs.WriteMetrics(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// AllMetrics - Получение всех сохраненных метрик с сортировкой по имени
func (fs *FileStorage) GetAllMetrics(ctx context.Context) []model.Metrica {
	return fs.MemStorage.GetAllMetrics(ctx)
}

// Debug - Показываем текущее состояние в output
func (fs *FileStorage) Debug(ctx context.Context) {
	fs.MemStorage.Debug(ctx)
}

// WriteMetrics - Метод для записи данных в хранилище
// Обновление файла происходит только если есть хоть одна метрика которая была изменена после последнего сохранения
func (fs *FileStorage) WriteMetrics(ctx context.Context) error {
	if fs.hasChanged {

		logger.Log.Debug().Msg("Синхронизация данных в файл")

		mtrks := fs.MemStorage.GetAllMetrics(ctx)

		bytes, err := json.Marshal(&mtrks)
		if err != nil {
			return err
		}

		err = os.WriteFile(fs.filename, bytes, 0666)
		if err != nil {
			return err
		}

		fs.hasChanged = false

	}

	return nil
}

// Check - Проверка хранилища, заглушка, FileStorage всегда готов к работе
func (fs *FileStorage) Check(ctx context.Context) error {
	return nil
}
