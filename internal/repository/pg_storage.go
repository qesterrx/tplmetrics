package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/qesterrx/tplmetrics/internal/logger"
	"github.com/qesterrx/tplmetrics/internal/model"
)

/*Реализация интерфейса MetricaStorage для хранения данных в БД PostgreSQL*/

type PGStorage struct {
	MemStorage
	databaseDSN string
	mode        MetricaStorageMode
	hasChanged  bool
	pool        *pgxpool.Pool
}

// Фабрика
func NewPGStorage(ms *MemStorage, databaseDSN string, mode MetricaStorageMode) (*PGStorage, error) {

	logger.Log.Debug().Msg("Создание PGStorage")
	//На подключение к БД, пинг, загрузку данных даем 10 секунд
	ctxto, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
	rows, err := pool.Query(ctxto, "select id, kind, delta, value from metrics")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var kindSrc string
		var delta int64
		var value float64

		err = rows.Scan(&name, &kindSrc, &delta, &value)
		if err != nil {
			return nil, err
		}

		kind, err := model.GetKindValue(kindSrc)
		if err != nil {
			return nil, err
		}

		var mtrk model.Metrica
		switch kind {
		case model.Gauge:
			mtrk = model.NewMetricaGauge(name, value)
		case model.Counter:
			mtrk = model.NewMetricaCounter(name, delta)
		default:
			return nil, fmt.Errorf("При загрузке данных из БД обнаружен неизвестный тип метрики")
		}

		pgs.MemStorage.UpdateMetrica(mtrk)
	}

	// проверяем на ошибки - не понял зачем это?
	err = rows.Err()
	if err != nil {
		return nil, err
	}

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
	//т.к. в памяти у нас хеш то тут напрямую к БД не обращаемся
	return pgs.MemStorage.Metrica(name, kind)
}

// Обновление метрики
func (pgs *PGStorage) UpdateMetrica(mtrk model.Metrica) error {

	err := pgs.MemStorage.UpdateMetrica(mtrk)
	if err != nil {
		return err
	}

	if pgs.mode == MetricaStorageModeSync {
		m, err := pgs.MemStorage.Metrica(mtrk.Name(), string(mtrk.Kind()))
		if err != nil {
			return err
		}
		return pgs.sqlUpdtaeMetrica(m)
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

// Сброc сданных из памяти в БД
func (pgs *PGStorage) WriteMetrics() error {

	if pgs.hasChanged {

		logger.Log.Debug().Msg("Синхронизация данных в БД")

		mtrks := pgs.MemStorage.AllMetrics()
		for _, mtrk := range mtrks {
			err := pgs.sqlUpdtaeMetrica(mtrk)
			if err != nil {
				return err
			}
		}

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

// функция для обновления метрики
func (pgs *PGStorage) sqlUpdtaeMetrica(mtrk model.Metrica) error {

	ctxto, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	var value float64
	var delta int64

	//Ну вот и все.. все-таки пришлось использовать type assertion хотя я всеми силами пытался этого избежать
	switch m := mtrk.(type) {
	case *model.MetricaCounter:
		delta = m.SrcValue()
	case *model.MetricaGauge:
		value = m.SrcValue()
	default:
		return fmt.Errorf("Неизвестный тип метрики в методе sqlUpdtaeMetrica")
	}

	_, err := pgs.pool.Exec(ctxto, `
INSERT INTO metrics (id, kind, delta, value, updated)
VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
ON CONFLICT (id, kind)
DO UPDATE SET
    delta = EXCLUDED.delta,
    value = EXCLUDED.value,
    updated = EXCLUDED.updated`, mtrk.Name(), mtrk.Kind(), delta, value)

	if err != nil {
		return err
	}

	return nil
}
