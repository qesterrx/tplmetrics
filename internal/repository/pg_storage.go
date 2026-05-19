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

// PGStorage - структура, реализующая интерфейс [service.MetricaStorage]
// обеспечивает сохранение метрик в БД
// поддерживает асинхронный режим работы
type PGStorage struct {
	*MemStorage
	mode       config.MetricaStorageMode
	hasChanged bool
	db         *sql.DB
}

// checkRetryPg - функци проверки ошибки на необходимость переотправки запроса в БД
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

// NewPGStorage - функция возвращает новый экземпляр PGStorage
// На вход ожидает
// ms - адрес экземпляра MemStorage (для хранения изменений в памяти)
// db - ссылка на подключение к БД
// mode - режим работы хранилища MetricaStorageMode
// * Всегда восстанавливает данные из БД в память перед началом работы
func NewPGStorage(ms *MemStorage, db *sql.DB, mode config.MetricaStorageMode) (*PGStorage, error) {

	logger.Log.Debug().Msg("Создание PGStorage")

	pgs := PGStorage{
		MemStorage: ms,
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

		err = pgs.MemStorage.UpdateMetrica(context.Background(), mtrk)
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

// GetMetrica - возвращает метрику по имени и типу
func (pgs *PGStorage) GetMetrica(ctx context.Context, name string, kind string) (model.Metrica, error) {
	//т.к. в памяти у нас кеш то тут напрямую к БД не обращаемся
	//главное чтобы экзепляр приложения был один и никто другой метрики не менял в БД/ФАЙЛЕ
	return pgs.MemStorage.GetMetrica(ctx, name, kind)
}

// UpdateMetrica - обновляет метрику
func (pgs *PGStorage) UpdateMetrica(ctx context.Context, mtrk model.Metrica) error {

	//Если асинхрон - то просто признак и выходим
	if pgs.mode != config.MetricaStorageModeSync {

		err := pgs.MemStorage.UpdateMetrica(ctx, mtrk)
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

		ctxTo, cancelCtxTo := context.WithTimeout(ctx, 1*time.Second)
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
	err = retry.RetryFunc(ctx, fn, checkRetryPg, 3, 1*time.Second, 2*time.Second)
	if err != nil {
		//Откатить изменения в памяти
		pgs.MemStorage.restoreUpdateMetricaBatch(touched)
		return err
	}

	//Зафиксировать изменения в памяти
	pgs.MemStorage.confirmUpdateMetricaBatch(touched)
	return nil

}

// UpdateMetricaBatch Обновляет массив метрик
func (pgs *PGStorage) UpdateMetricaBatch(ctx context.Context, mtrks []model.Metrica) error {

	if len(mtrks) == 0 {
		return nil
	}

	//Если асинхрон - то просто признак и выходим
	if pgs.mode != config.MetricaStorageModeSync {

		err := pgs.MemStorage.UpdateMetricaBatch(ctx, mtrks)
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

		ctxTo, cancelCtxTo := context.WithTimeout(ctx, 1*time.Second)
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
	err = retry.RetryFunc(ctx, fn, checkRetryPg, 3, 1*time.Second, 2*time.Second)
	if err != nil {
		//Откатить изменения в памяти
		pgs.MemStorage.restoreUpdateMetricaBatch(touched)
		return err
	}

	//Зафиксировать изменения в памяти
	pgs.MemStorage.confirmUpdateMetricaBatch(touched)
	return nil

}

// AllMetrics - Получение всех сохраненных метрик с сортировкой по имени
func (pgs *PGStorage) GetAllMetrics(ctx context.Context) []model.Metrica {
	return pgs.MemStorage.GetAllMetrics(ctx)
}

// Debug - Показываем текущее состояние в output
func (pgs *PGStorage) Debug(ctx context.Context) {
	pgs.MemStorage.Debug(ctx)
}

// WriteMetrics - Метод для записи данных в хранилище
// Обновление файла происходит только если есть хоть одна метрика которая была изменена после последнего сохранения
func (pgs *PGStorage) WriteMetrics(ctx context.Context) error {

	if pgs.hasChanged {

		logger.Log.Debug().Msg("Синхронизация данных в БД")

		//Замыкание для повторов
		fn := func() error {
			ctxTo, cancelCtxTo := context.WithTimeout(ctx, 1*time.Second)
			defer cancelCtxTo()

			tx, err := pgs.db.BeginTx(ctxTo, nil)
			if err != nil {
				return err
			}

			mtrks := pgs.MemStorage.GetAllMetrics(ctx)
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
		return retry.RetryFunc(ctx, fn, checkRetryPg, 3, 1*time.Second, 2*time.Second)

	}

	return nil
}

// Check - Проверка хранилища
func (pgs *PGStorage) Check(ctx context.Context) error {
	ctxTo, cancelCtxTo := context.WithTimeout(ctx, 1*time.Second)
	defer cancelCtxTo()
	return pgs.db.PingContext(ctxTo)
}

// sqlUpdateMetricaTx - дополнительная функция для изменения массива метрик в рамках одной транзакции БД
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
