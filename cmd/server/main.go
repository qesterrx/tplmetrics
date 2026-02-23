package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/qesterrx/tplmetrics/internal/config"
	"github.com/qesterrx/tplmetrics/internal/handler"
	"github.com/qesterrx/tplmetrics/internal/logger"
	"github.com/qesterrx/tplmetrics/internal/repository"
	"github.com/rs/zerolog"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger.InitLogger()
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	config, err := config.ParseParamsServer()
	if err != nil {
		panic(err)
	}

	storage := repository.NewMemStorage()

	if config.RestoreFromFileStorage {
		err = repository.LoadDataFromFile(config.FileStorageName, storage)
		if err != nil {
			logger.Log.Error().Msg("Ошибка при загрузке даннных из файла  " + config.FileStorageName + ":" + err.Error())
		}
	}

	var wg sync.WaitGroup

	if config.StoreInterval == 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger.Log.Debug().Msg("Запуск EventSaver")
			repository.EventSaver(ctx, config.FileStorageName, storage)
			cancel()
		}()
	}

	if config.StoreInterval > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger.Log.Debug().Msg("Запуск TimeSaver")
			repository.TimeSaver(ctx, config.FileStorageName, config.StoreInterval, storage)
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
		logger.Log.Error().Msg("Ошибка в работе сервера ListenAndServe:" + err.Error())
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
}
