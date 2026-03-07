package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/qesterrx/tplmetrics/internal/logger"
	"github.com/qesterrx/tplmetrics/internal/model"
)

/**/

type PGStorage struct {
	MemStorage
	databaseDSN string
	mode        MetricaStorageMode
	restore     bool
	hasChanged  bool
	pool        *pgxpool.Pool
}

// Фабрика
func NewPGStorage(ms *MemStorage, databaseDSN string, mode MetricaStorageMode) (*PGStorage, error) {

	logger.Log.Debug().Msg("Создание PGStorage")
	//На подключение к БД  и пингдаем 5 секунд
	ctxto, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//Коннект
	pool, err := pgxpool.New(ctxto, databaseDSN)
	if err != nil {
		return nil, err
	}

	//Проверяем соединение
	err = pool.Ping(ctxto)
	if err != nil {
		return nil, err
	}

	pgs := PGStorage{
		MemStorage:  *ms,
		mode:        mode,
		databaseDSN: databaseDSN,
		hasChanged:  false,
		pool:        pool,
	}

	//Загружаем данные по сохраненным метрикам

	return &pgs, nil
}

// Завершаем работу с PG
func (pgs *PGStorage) Close() {

	logger.Log.Debug().Msg("Закрытие PGStorage")

	err := pgs.WriteMetrics()
	if err != nil {
		logger.Log.Debug().Msg("Ошибка сохранения данных PGStorage " + err.Error())
	}
	pgs.pool.Close()
}

// Получение метрики по имени
func (pgs *PGStorage) Metrica(name string, kind string) (model.Metrica, error) {
	return pgs.MemStorage.Metrica(name, kind)
}

// Обновление метрики
func (pgs *PGStorage) UpdateMetrica(mtrk model.Metrica) error {

	err := pgs.MemStorage.UpdateMetrica(mtrk)
	if err != nil {
		return err
	}

	if pgs.mode == MetricaStorageModeSync {
		return pgs.WriteMetrics() //Нет смысла записывать все метрики
	} else {
		pgs.hasChanged = true
		return nil
	}

}

// Получение всех сохраненных, с сортировкой по имени
func (pgs *PGStorage) AllMetrics() []model.Metrica {
	return pgs.MemStorage.AllMetrics()
}

// Показываем текущее состояние в output
func (pgs *PGStorage) Debug() {
	pgs.MemStorage.Debug()
}

// Сбро сданных из памяти в БД
func (pgs *PGStorage) WriteMetrics() error {
	if pgs.hasChanged {

		logger.Log.Debug().Msg("Синхронизация данных в файл")

		mtrks := pgs.MemStorage.AllMetrics()

		_, err := json.Marshal(&mtrks)
		if err != nil {
			return err
		}

		/*err = os.WriteFile(fs.filename, bytes, 0666)
		if err != nil {
			return err
		}*/

		pgs.hasChanged = false

	}

	return nil
}

// Пинг для 10 инкремента
func (pgs *PGStorage) PingDB() error {
	ctxto, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	return pgs.pool.Ping(ctxto)
}
