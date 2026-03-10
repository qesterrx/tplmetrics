package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/qesterrx/tplmetrics/internal/logger"
	"github.com/qesterrx/tplmetrics/internal/model"
)

/*Реализация интерфейса MetricaStorage для хранения данных в БД PostgreSQL*/

type PGStorage struct {
	MemStorage
	mode       MetricaStorageMode
	hasChanged bool
	db         *sql.DB
}

// Фабрика
func NewPGStorage(ms *MemStorage, db *sql.DB, mode MetricaStorageMode) (*PGStorage, error) {

	logger.Log.Debug().Msg("Создание PGStorage")

	pgs := PGStorage{
		MemStorage: *ms,
		mode:       mode,
		hasChanged: false,
		db:         db,
	}

	ctxTo, cancelCtxTo := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelCtxTo()

	//Загружаем данные по сохраненным метрикам
	rows, err := db.QueryContext(ctxTo, "select id, kind, delta, value from metrics")
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
			return nil, fmt.Errorf("при загрузке данных из БД обнаружен неизвестный тип метрики")
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

// Получение метрики по имени
func (pgs *PGStorage) Metrica(name string, kind string) (model.Metrica, error) {
	//т.к. в памяти у нас хеш то тут напрямую к БД не обращаемся
	return pgs.MemStorage.Metrica(name, kind)
}

// Обновление метрики
func (pgs *PGStorage) UpdateMetrica(mtrk model.Metrica) error {

	//Я бы конечно использовал WriteMetrics но для чистоты экскримента сделаю тут по другому
	ctxTo, cancelCtxTo := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancelCtxTo()

	err := pgs.MemStorage.UpdateMetrica(mtrk)
	if err != nil {
		return err
	}

	if pgs.mode == MetricaStorageModeSync {
		m, err := pgs.MemStorage.Metrica(mtrk.Name(), string(mtrk.Kind()))
		if err != nil {
			return err
		}
		return pgs.sqlUpdtaeMetrica(ctxTo, m)
	} else {
		pgs.hasChanged = true
		return nil
	}

}

// Обновление массива метрик
func (pgs *PGStorage) UpdateMetricaBatch(mtrks []model.Metrica) error {

	if len(mtrks) == 0 {
		return nil
	}

	//Я бы конечно использовал WriteMetrics но для чистоты экскримента сделаю тут по другому
	ctxTo, cancelCtxTo := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancelCtxTo()

	tx, err := pgs.db.BeginTx(ctxTo, nil)
	if err != nil {
		return err
	}

	for _, mtrk := range mtrks {
		err := pgs.MemStorage.UpdateMetrica(mtrk)
		if err != nil {
			tx.Rollback()
			return err
		}

		if pgs.mode == MetricaStorageModeSync {
			err := pgs.sqlUpdtaeMetrica(ctxTo, mtrk)
			if err != nil {
				tx.Rollback()
				return err
			}
		} else {
			pgs.hasChanged = true
		}
	}

	tx.Commit()

	return nil

}

// Получение всех сохраненных, с сортировкой по имени
func (pgs *PGStorage) AllMetrics() []model.Metrica {
	return pgs.MemStorage.AllMetrics()
}

// Показываем текущее состояние в output
func (pgs *PGStorage) Debug() {
	pgs.MemStorage.Debug()
}

// Сброc сданных из памяти в БД - ну извращение же, хотя для метрик может быть и норм
func (pgs *PGStorage) WriteMetrics() error {

	if pgs.hasChanged {

		logger.Log.Debug().Msg("Синхронизация данных в БД")

		ctxTo, cancelCtxTo := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancelCtxTo()

		tx, err := pgs.db.BeginTx(ctxTo, nil)
		if err != nil {
			return err
		}

		mtrks := pgs.MemStorage.AllMetrics()
		for _, mtrk := range mtrks {
			err := pgs.sqlUpdtaeMetrica(ctxTo, mtrk)
			if err != nil {
				tx.Rollback()
				return err
			}
		}

		tx.Commit()
		pgs.hasChanged = false

	}

	return nil
}

// Пинг для 10 инкремента
func (pgs *PGStorage) PingDB() error {
	ctxTo, cancelCtxTo := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancelCtxTo()
	return pgs.db.PingContext(ctxTo)
}

// функция для обновления метрики
func (pgs *PGStorage) sqlUpdtaeMetrica(ctx context.Context, mtrk model.Metrica) error {

	var value float64
	var delta int64

	//Ну вот и все.. все-таки пришлось использовать type assertion хотя я всеми силами пытался этого избежать
	switch m := mtrk.(type) {
	case *model.MetricaCounter:
		delta = m.SrcValue()
	case *model.MetricaGauge:
		value = m.SrcValue()
	default:
		return fmt.Errorf("неизвестный тип метрики в методе sqlUpdtaeMetrica")
	}

	_, err := pgs.db.ExecContext(ctx, `
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
