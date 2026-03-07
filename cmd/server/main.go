package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/qesterrx/tplmetrics/internal/config"
	"github.com/qesterrx/tplmetrics/internal/handler"
	"github.com/qesterrx/tplmetrics/internal/logger"
	"github.com/qesterrx/tplmetrics/internal/repository"
	"github.com/rs/zerolog"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger.InitLogger()
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	config, err := config.ParseParamsServer()
	if err != nil {
		logger.Log.Error().Err(err)
		return err
	}

	var storage repository.MetricaStorage
	var mode repository.MetricaStorageMode

	if config.StoreInterval == 0 {
		mode = repository.MetricaStorageModeSync
	} else {
		mode = repository.MetricaStorageModeAsync
	}

	//Всегда создаем memStorage
	memStorage := repository.NewMemStorage()

	//Дальше пытаемся подобрать реальный Storage по параметрам
	if config.DatabaseDSN != "" {
		//Хранение в БД постгри

		//Сначала запускаем миграции
		if config.DatabaseURL != "" {
			m, err := migrate.New("file://migrations", config.DatabaseURL)
			if err != nil {
				return err
			}

			if err := m.Up(); err != nil && err != migrate.ErrNoChange {
				return err
			}
		}

		pgStorage, err := repository.NewPGStorage(memStorage, config.DatabaseDSN, mode)
		if err != nil {
			logger.Log.Error().Err(err)
			return err
		}
		defer pgStorage.Close()
		storage = pgStorage

	} else if config.FileStorageName != "" {
		//Хранение в файле
		fileStorage, err := repository.NewFileStorage(memStorage, config.FileStorageName, mode, config.RestoreFromFileStorage)
		if err != nil {
			logger.Log.Error().Err(err)
			return err
		}
		defer fileStorage.Close()
		storage = fileStorage

	} else {
		//Хранение в памяти
		storage = memStorage

	}

	var wg sync.WaitGroup

	//На самом деле этот кусочек имеет смысл только если у storage есть куда сохранять данные
	//А вообще конечно передаю привет тому извращенцу который придумал эту логику, а так же наставикам курса которые не могут сказать как это предпологалось сделать
	if mode == repository.MetricaStorageModeAsync {
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger.Log.Debug().Msg("Запуск TickerWriteMetrics")
			repository.TickerWriteMetrics(ctx, storage, config.StoreInterval)
			cancel()
		}()
	}

	server := &http.Server{
		Addr:    config.ServerHost.String(),
		Handler: handler.GetRouter(storage),
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Log.Debug().Msg("Запуск HttpServer")
		err := server.ListenAndServe()
		if ctx.Err() == nil {
			//Ошибку отображаем только если контекст не завершен
			logger.Log.Error().Msg("Ошибка в работе сервера ListenAndServe:" + err.Error())
		}
		cancel()
	}()

	// Канал для сигналов ОС
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Ждем сигнал завершения
	<-sigChan
	cancel()

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	// Пытаемся остановить сервер gracefully
	if err := server.Shutdown(ctxShutdown); err != nil {
		logger.Log.Error().Msg("Ошибка остановки работы сервера:" + err.Error())
	}

	logger.Log.Info().Msg("Сервер HttpServer остановлен")
	wg.Wait()

	return nil
}
