package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/qesterrx/tplmetrics/internal/config"
	"github.com/qesterrx/tplmetrics/internal/logger"
	"github.com/qesterrx/tplmetrics/internal/model"
	"github.com/qesterrx/tplmetrics/internal/retry"
)

/*Реализация интерфейса MetricaStorage для хранения данных в БД PostgreSQL*/

type PGStorage struct {
	MemStorage
	mode       config.MetricaStorageMode
	hasChanged bool
	db         *sql.DB
}

// Функция проверяет ошибку на возможность retry
func checkRetryPg(err error) bool {

	//По идее тут тоже могут быть ошибки сетевого взаимодействия, если БД вдруг упала скорее всего будет так же connect: connection refused

	//пу-пу-пу
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if len(pgErr.Code) >= 2 && pgErr.Code[:2] == "08" {
			logger.Log.Error().Msg("ошибка выполнения в postgresql, code:" + pgErr.Code)
			return true
		}
	}

	return false

}

// Фабрика
func NewPGStorage(ms *MemStorage, db *sql.DB, mode config.MetricaStorageMode) (*PGStorage, error) {

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

		err = pgs.MemStorage.UpdateMetrica(mtrk)
		if err != nil {
			return nil, err
		}
	}

	// проверяем на ошибки - не понял зачем это?
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return &pgs, nil
}

// Получение метрики по имени
func (pgs *PGStorage) GetMetrica(name string, kind string) (model.Metrica, error) {
	//т.к. в памяти у нас кеш то тут напрямую к БД не обращаемся
	//главное чтобы экзепляр приложения был один и никто другой метрики не менял в БД/ФАЙЛЕ
	return pgs.MemStorage.GetMetrica(name, kind)
}

// Обновление метрики
func (pgs *PGStorage) UpdateMetrica(mtrk model.Metrica) error {

	//Если асинхрон - то просто признак и выходим
	if pgs.mode != config.MetricaStorageModeSync {

		err := pgs.MemStorage.UpdateMetrica(mtrk)
		if err != nil {
			return err
		}

		pgs.hasChanged = true
		return nil
	}

	touched, err := pgs.MemStorage.startUpdateMetricaBatch([]model.Metrica{mtrk})
	if err != nil {
		pgs.restoreUpdateMetricaBatch(touched)
		return err
	}

	//Замыкание для повторов
	fn := func() error {

		ctxTo, cancelCtxTo := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancelCtxTo()

		//По идее тут будут накладные расходы на транзакцию, писать два метода не захотел, мешать в один тоже, может придумаю что получше попозже
		tx, err := pgs.db.BeginTx(ctxTo, nil)
		if err != nil {
			return err
		}

		//Точно знаем что в touched один элемент
		mtrkCurr := (*touched)[0]
		err = pgs.sqlUpdateMetricaTx(ctxTo, tx, *mtrkCurr, "new")
		if err != nil {
			tx.Rollback()
			return err
		}

		tx.Commit()

		return nil

	}

	//Выполнение метода
	err = retry.RetryFunc(context.Background(), fn, checkRetryPg, 3, 1*time.Second, 2*time.Second)
	if err != nil {
		//Откатить изменения в памяти
		pgs.MemStorage.restoreUpdateMetricaBatch(touched)
		return err
	}

	//Зафиксировать изменения в памяти
	pgs.MemStorage.confirmUpdateMetricaBatch(touched)
	return nil

}

// Обновление массива метрик
func (pgs *PGStorage) UpdateMetricaBatch(mtrks []model.Metrica) error {

	if len(mtrks) == 0 {
		return nil
	}

	//Если асинхрон - то просто признак и выходим
	if pgs.mode != config.MetricaStorageModeSync {

		err := pgs.MemStorage.UpdateMetricaBatch(mtrks)
		if err != nil {
			return err
		}

		pgs.hasChanged = true
		return nil
	}

	//Обновляем метрики в памяти без повторов
	touched, err := pgs.MemStorage.startUpdateMetricaBatch(mtrks)
	if err != nil {
		pgs.restoreUpdateMetricaBatch(touched)
		return err
	}

	//Замыкание для повторов
	fn := func() error {

		ctxTo, cancelCtxTo := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancelCtxTo()

		tx, err := pgs.db.BeginTx(ctxTo, nil)
		if err != nil {
			return err
		}

		//Раз у нас есть ссылки на затронутые объекты, почему бы этим не воспользоваться
		for _, mtrk := range *touched {

			err = pgs.sqlUpdateMetricaTx(ctxTo, tx, *mtrk, "new")
			if err != nil {
				tx.Rollback()
				return err
			}

		}

		tx.Commit()
		return nil

	}

	//Выполнение метода
	err = retry.RetryFunc(context.Background(), fn, checkRetryPg, 3, 1*time.Second, 2*time.Second)
	if err != nil {
		//Откатить изменения в памяти
		pgs.MemStorage.restoreUpdateMetricaBatch(touched)
		return err
	}

	//Зафиксировать изменения в памяти
	pgs.MemStorage.confirmUpdateMetricaBatch(touched)
	return nil

}

// Получение всех сохраненных, с сортировкой по имени
func (pgs *PGStorage) GetAllMetrics() []model.Metrica {
	return pgs.MemStorage.GetAllMetrics()
}

// Показываем текущее состояние в output
func (pgs *PGStorage) Debug() {
	pgs.MemStorage.Debug()
}

// Сброc сданных из памяти в БД - ну извращение же, хотя для метрик может быть и норм
func (pgs *PGStorage) WriteMetrics() error {

	if pgs.hasChanged {

		logger.Log.Debug().Msg("Синхронизация данных в БД")

		//Замыкание для повторов
		fn := func() error {
			ctxTo, cancelCtxTo := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancelCtxTo()

			tx, err := pgs.db.BeginTx(ctxTo, nil)
			if err != nil {
				return err
			}

			mtrks := pgs.MemStorage.GetAllMetrics()
			for _, mtrk := range mtrks {
				err := pgs.sqlUpdateMetricaTx(ctxTo, tx, mtrk, "current")
				if err != nil {
					tx.Rollback()
					return err
				}
			}

			tx.Commit()
			pgs.hasChanged = false
			return nil
		}

		//Выполнение метода
		return retry.RetryFunc(context.Background(), fn, checkRetryPg, 3, 1*time.Second, 2*time.Second)

	}

	return nil
}

// Пинг для 10 инкремента
func (pgs *PGStorage) Check() error {
	ctxTo, cancelCtxTo := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancelCtxTo()
	return pgs.db.PingContext(ctxTo)
}

// функция для обновления метрики
func (pgs *PGStorage) sqlUpdateMetricaTx(ctx context.Context, tx *sql.Tx, mtrk model.Metrica, env string) error {

	if tx == nil {
		return fmt.Errorf("процедура sqlUpdateMetricaTx ожидает ссылку на транзакцию Tx")
	}

	var value float64
	var delta int64

	//Ну вот и все.. все-таки пришлось использовать type assertion хотя я всеми силами пытался этого избежать
	switch m := mtrk.(type) {
	case *model.MetricaCounter:
		delta = m.SrcValue(env)
	case *model.MetricaGauge:
		value = m.SrcValue(env)
	default:
		return fmt.Errorf("неизвестный тип метрики в методе sqlUpdateMetrica")
	}

	_, err := tx.ExecContext(ctx, `
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
