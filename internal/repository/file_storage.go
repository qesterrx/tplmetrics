package repository

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/qesterrx/tplmetrics/internal/logger"
	"github.com/qesterrx/tplmetrics/internal/model"
)

/*Реализация интерфейса MetricaStorage для хранения данных в файле*/

type FileStorage struct {
	MemStorage
	filename   string
	mode       MetricaStorageMode
	restore    bool
	hasChanged bool
}

// Фабрика
func NewFileStorage(ms *MemStorage, filename string, mode MetricaStorageMode, restore bool) (*FileStorage, error) {

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
		MemStorage: *ms,
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
				fs.MemStorage.UpdateMetrica(mtrk)
			}
		}

	}

	return &fs, nil
}

// Завершаем работу с файлом
func (fs *FileStorage) Close() {
	logger.Log.Debug().Msg("Закрытие FileStorage")

	err := fs.WriteMetrics()
	if err != nil {
		logger.Log.Debug().Msg("Ошибка сохранения данных FileStorage " + err.Error())
	}
}

// Получение метрики по имени
func (fs *FileStorage) Metrica(name string, kind string) (model.Metrica, error) {
	return fs.MemStorage.Metrica(name, kind)
}

// Обновление метрики
func (fs *FileStorage) UpdateMetrica(mtrk model.Metrica) error {

	err := fs.MemStorage.UpdateMetrica(mtrk)
	if err != nil {
		return err
	}

	fs.hasChanged = true
	if fs.mode == MetricaStorageModeSync {
		return fs.WriteMetrics()
	} else {
		return nil
	}

}

// Обновление массива метрик
func (fs *FileStorage) UpdateMetricaBatch(mtrks []model.Metrica) error {
	err := fs.MemStorage.UpdateMetricaBatch(mtrks)
	if err != nil {
		return err
	}

	fs.hasChanged = true
	if fs.mode == MetricaStorageModeSync {
		return fs.WriteMetrics()
	} else {
		return nil
	}

}

// Получение всех сохраненных, с сортировкой по имени
func (fs *FileStorage) AllMetrics() []model.Metrica {
	return fs.MemStorage.AllMetrics()
}

// Показываем текущее состояние в output
func (fs *FileStorage) Debug() {
	fs.MemStorage.Debug()
}

// Метод для записи данных в хранилище
func (fs *FileStorage) WriteMetrics() error {
	if fs.hasChanged {

		logger.Log.Debug().Msg("Синхронизация данных в файл")

		mtrks := fs.MemStorage.AllMetrics()

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
