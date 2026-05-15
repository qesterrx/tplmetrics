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

// Интерфейс хранилки
type MetricaStorage interface {
	//Метод для обновления данных метрики
	UpdateMetrica(model.Metrica) error
	//Обновление массива метрик
	UpdateMetricaBatch([]model.Metrica) error
	//Метод получения экземпляра метрики по имени
	GetMetrica(name string, kind string) (model.Metrica, error)
	//Получение всех метрик
	GetAllMetrics() []model.Metrica
	//Метод для "сброса" накопившихся записей в долговременное хранилище
	WriteMetrics() error
	//Проверка хранилища
	Check() error
	//Отладочный вызов
	Debug()
}

type UpdateMetricaSubscriber interface {
	PushNotify(msg []byte)
}

// Структура под логику
type TCLService struct {
	config  *config.ConfigServer
	storage MetricaStorage
	subs    []UpdateMetricaSubscriber
}

// Конструктор
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

	storage.Debug()

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

// Проверка хранилища
func (tcl *TCLService) Check(ctx context.Context) error {
	return tcl.storage.Check()
}

// Метод для обновления данных метрики
func (tcl *TCLService) UpdateMetrica(ctx context.Context, mtrk model.Metrica) error {
	res := tcl.storage.UpdateMetrica(mtrk)
	mtrks := []model.Metrica{}
	mtrks = append(mtrks, mtrk)

	tcl.NotifyUpdateMetrica(ctx, mtrks)
	return res
}

// Метод для обновления данных метрик
func (tcl *TCLService) UpdateMetricaBatch(ctx context.Context, mtrks []model.Metrica) error {
	res := tcl.storage.UpdateMetricaBatch(mtrks)
	tcl.NotifyUpdateMetrica(ctx, mtrks)
	return res
}

// Метод получения экземпляра метрики по имени
func (tcl *TCLService) GetMetrica(ctx context.Context, name string, kind string) (model.Metrica, error) {
	return tcl.storage.GetMetrica(name, kind)
}

// Получение всех метрик
func (tcl *TCLService) GetAllMetrics(ctx context.Context) []model.Metrica {
	return tcl.storage.GetAllMetrics()
}

// Добавляем подписчика
func (tcl *TCLService) AddUpdateMetricaSubscriber(sub UpdateMetricaSubscriber) {
	//TODO тут нет проверки на то что такой подписчки уже есть в массиве
	if sub != nil {
		tcl.subs = append(tcl.subs, sub)
	}
}

// Оповещаем подписчиков о обновлении метрик
func (tcl *TCLService) NotifyUpdateMetrica(ctx context.Context, mtrks []model.Metrica) {

	if len(tcl.subs) == 0 || len(mtrks) == 0 {
		return
	}

	ip_key := ctx.Value(middleware.ContextIP)
	ip := ""

	if ip_key != nil {
		tmp, ok := ip_key.(string)
		if ok {
			ip = tmp
		}
	}

	mNames := []string{}
	for k := range mtrks {
		mNames = append(mNames, mtrks[k].Name())
	}

	msg := model.UpdateMetricaSubscriberMsg{Ts: time.Now().Unix(), Metrics: mNames, IP: ip}
	json, err := json.MarshalIndent(msg, "", " ")
	if err != nil {
		logger.Log.Error().Msg("NotifyUpdateMetrica: Ошибка сериализации UpdateMetricaSubscriberMsg")
	}

	for _, sub := range tcl.subs {
		sub.PushNotify(json)
	}

}

// Переодическая запись текущего состояния в хранилище
func (tcl *TCLService) TickerWriteMetrics(ctx context.Context) {

	ticker := time.NewTicker(time.Second * time.Duration(tcl.config.StoreInterval))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Debug().Msg("Завершение TickerWriteMetrics по контексту")
			//Перед тем как выйти сделаем запись
			err := tcl.storage.WriteMetrics()
			if err != nil {
				logger.Log.Error().Msg("TickerWriteMetrics Ошибка при сохранении данных в файл" + err.Error())
			}
			return
		case <-ticker.C:
			err := tcl.storage.WriteMetrics()
			if err != nil {
				logger.Log.Error().Msg("TickerWriteMetrics Ошибка при сохранении данных в файл" + err.Error())
			}
			logger.Log.Debug().Msg("TickerWriteMetrics сброс данных в хранилище")
		}
	}
}
