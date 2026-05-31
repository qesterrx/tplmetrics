// Пакет service предназначен для описания основной логики приложения
package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/qesterrx/tplmetrics/internal/config"
	"github.com/qesterrx/tplmetrics/internal/logger"
	"github.com/qesterrx/tplmetrics/internal/middleware"
	"github.com/qesterrx/tplmetrics/internal/model"
	"github.com/qesterrx/tplmetrics/internal/repository"
)

// MetricaStorage - Интерфейс описывающий методы для работы с долговременным хранилищем данных
type MetricaStorage interface {

	//UpdateMetrica - Метод для обновления данных метрики
	UpdateMetrica(ctx context.Context, mtrk model.Metrica) error

	//UpdateMetricaBatch - Обновление массива метрик
	UpdateMetricaBatch(ctx context.Context, mtrks []model.Metrica) error

	//GetMetrica - Метод получения экземпляра метрики по имени
	GetMetrica(ctx context.Context, name string, kind string) (model.Metrica, error)

	//GetAllMetrics - Получение всех метрик
	GetAllMetrics(ctx context.Context) []model.Metrica

	//WriteMetrics - Метод для "сброса" накопившихся записей в долговременное хранилище
	WriteMetrics(ctx context.Context) error

	//Check - Проверка хранилища
	Check(ctx context.Context) error

	//Debug - Отладочный вызов
	Debug(ctx context.Context)
}

// UpdateMetricaSubscriber - интерфейс подписчика аудита
type UpdateMetricaSubscriber interface {
	PushNotify(ctx context.Context, msg []byte)
}

// TCLService - Структура объекта содержащего сервисный слой обработчиков
//
// generate:reset
type TCLService struct {
	config  *config.ConfigServer
	storage MetricaStorage
	subs    []UpdateMetricaSubscriber
}

// NewTCLService - Функция возвращает новый экземпляр TCLService.
// На вход передается ссылка на конфигурацию сервера
// Внутри метода происходит выбор постоянного хранилища данных на основе переданных параметров по следующей логике
// Если передан DatabaseDSN то исползуется postgres
// Иначе, если передан FileStorageName то используется файл
// Иначе данные хранятся только в ОЗУ
//
// Так же на основе параметров AuditFile/AuditURL определяются подписчики на аудит изменения метрик
func NewTCLService(config *config.ConfigServer) (*TCLService, error) {

	var storage MetricaStorage

	//Всегда создаем memStorage
	memStorage := repository.NewMemStorage()

	//Дальше пытаемся подобрать реальный Storage по параметрам
	if config.DatabaseDSN != "" {
		//Создаем подключение
		conn, err := sql.Open("pgx", config.DatabaseDSN)
		if err != nil {
			return nil, err
		}

		//Проверяем подключение
		if err := conn.Ping(); err != nil {
			return nil, err
		}

		//Создаем driver для migrate используя существующее подключение
		driver, err := postgres.WithInstance(conn, &postgres.Config{})
		if err != nil {
			return nil, err
		}

		//Создаем экземпляр migrate
		m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
		if err != nil {
			return nil, err
		}

		//Запускаем миграции
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return nil, err
		}

		//Хранение в БД постгри
		pgStorage, err := repository.NewPGStorage(memStorage, conn, config.StorageMode)
		if err != nil {
			logger.Log.Error().Err(err)
			return nil, err
		}
		storage = pgStorage

	} else if config.FileStorageName != "" {
		//Хранение в файле
		fileStorage, err := repository.NewFileStorage(memStorage, config.FileStorageName, config.StorageMode, config.RestoreFromFileStorage)
		if err != nil {
			logger.Log.Error().Err(err)
			return nil, err
		}
		storage = fileStorage

	} else {
		//Хранение в памяти
		storage = memStorage

	}

	/*Раз уж мы передали всю конфигурацию сюда, то и подписчиков создадим тут*/
	subs := []UpdateMetricaSubscriber{}
	if config.AuditFile != "" {
		subs = append(subs, &UpdateMetricaSubscriberFile{FileName: config.AuditFile})
	}
	if config.AuditURL != "" {
		subs = append(subs, &UpdateMetricaSubscriberClient{URL: config.AuditURL})
	}

	return &TCLService{config: config, storage: storage, subs: subs}, nil
}

// Check - метод проверки готовности хранилища
func (tcl *TCLService) Check(ctx context.Context) error {
	return tcl.storage.Check(ctx)
}

// UpdateMetrica - метод обновления значения метрики
func (tcl *TCLService) UpdateMetrica(ctx context.Context, mtrk model.Metrica) error {
	res := tcl.storage.UpdateMetrica(ctx, mtrk)

	//Аудит
	mtrks := []model.Metrica{}
	mtrks = append(mtrks, mtrk)
	tcl.notifyUpdateMetrica(ctx, mtrks)

	return res
}

// UpdateMetricaBatch - метод для пакетного обновления метрик
func (tcl *TCLService) UpdateMetricaBatch(ctx context.Context, mtrks []model.Metrica) error {
	res := tcl.storage.UpdateMetricaBatch(ctx, mtrks)

	//Аудит
	tcl.notifyUpdateMetrica(ctx, mtrks)

	return res
}

// GetMetrica - метод получения экземпляра метрики по имени
func (tcl *TCLService) GetMetrica(ctx context.Context, name string, kind string) (model.Metrica, error) {
	return tcl.storage.GetMetrica(ctx, name, kind)
}

// GetAllMetrics - метод получения всех метрик
func (tcl *TCLService) GetAllMetrics(ctx context.Context) []model.Metrica {
	return tcl.storage.GetAllMetrics(ctx)
}

// AddUpdateMetricaSubscriber - Добавляет подписчика аудита для уведомления об обновлении метрик
func (tcl *TCLService) AddUpdateMetricaSubscriber(sub UpdateMetricaSubscriber) {
	//TODO тут нет проверки на то что такой подписчки уже есть в массиве
	if sub != nil {
		tcl.subs = append(tcl.subs, sub)
	}
}

// notifyUpdateMetrica - Функция для оповещения подпичиков об изменении метрик
func (tcl *TCLService) notifyUpdateMetrica(ctx context.Context, mtrks []model.Metrica) {

	if len(tcl.subs) == 0 || len(mtrks) == 0 {
		return
	}

	ipInCtx := ctx.Value(middleware.ContextIP)
	ip := ""

	if ipInCtx != nil {
		tmp, ok := ipInCtx.(string)
		if ok {
			ip = tmp
		}
	}

	mNames := []string{}
	for k := range mtrks {
		mNames = append(mNames, mtrks[k].Name())
	}

	msg := model.UpdateMetricaSubscriberMsg{TS: time.Now().Unix(), Metrics: mNames, IP: ip}
	json, err := json.MarshalIndent(msg, "", " ")
	if err != nil {
		logger.Log.Error().Msg("NotifyUpdateMetrica: Ошибка сериализации UpdateMetricaSubscriberMsg")
	}

	for _, sub := range tcl.subs {
		sub.PushNotify(ctx, json)
	}

}

// TickerWriteMetrics - Переодическая запись текущего состояния в хранилище, в случае асинхронного режима работы сервиса
func (tcl *TCLService) TickerWriteMetrics(ctx context.Context) {

	ticker := time.NewTicker(time.Second * time.Duration(tcl.config.StoreInterval))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Debug().Msg("Завершение TickerWriteMetrics по контексту")
			//Перед тем как выйти сделаем запись
			err := tcl.storage.WriteMetrics(ctx)
			if err != nil {
				logger.Log.Error().Msg("TickerWriteMetrics Ошибка при сохранении данных в файл" + err.Error())
			}
			return
		case <-ticker.C:
			err := tcl.storage.WriteMetrics(ctx)
			if err != nil {
				logger.Log.Error().Msg("TickerWriteMetrics Ошибка при сохранении данных в файл" + err.Error())
			}
			logger.Log.Debug().Msg("TickerWriteMetrics сброс данных в хранилище")
		}
	}
}
