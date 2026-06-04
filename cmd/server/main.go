package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/qesterrx/tplmetrics/internal/config"
	"github.com/qesterrx/tplmetrics/internal/handler"
	"github.com/qesterrx/tplmetrics/internal/logger"
	"github.com/qesterrx/tplmetrics/internal/service"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"

	_ "net/http/pprof"
)

var buildVersion string
var buildDate string
var buildCommit string

func main() {

	nvl := func(str string) string {
		if str == "" {
			return "N/A"
		}
		return str
	}

	fmt.Println("Build version:", nvl(buildVersion))
	fmt.Println("Build date:", nvl(buildDate))
	fmt.Println("Build commit:", nvl(buildCommit))

	// Запускаем HTTP сервер для pprof
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger.InitLogger()
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	//Конфигурация
	cfg, err := config.ParseParamsServer()
	if err != nil {
		logger.Log.Error().Err(err)
		return err
	}

	//Сервис
	tcl, err := service.NewTCLService(cfg)
	if err != nil {
		logger.Log.Error().Err(err)
		return err
	}

	//Объект с хендлерами
	hc := handler.NewHandlerContainer(tcl, cfg.SecretKeyForSign)

	g, ctx := errgroup.WithContext(ctx)

	//На самом деле этот кусочек имеет смысл только если у storage есть куда сохранять данные
	//А вообще конечно передаю привет тому извращенцу который придумал эту логику, а так же наставикам курса которые не могут сказать как это предпологалось сделать
	if cfg.StorageMode == config.MetricaStorageModeAsync {
		g.Go(func() error {
			logger.Log.Debug().Msg("Запуск TickerWriteMetrics")
			tcl.TickerWriteMetrics(ctx)
			return nil
		})
	}

	server := &http.Server{
		Addr:         cfg.ServerHost.String(),
		Handler:      hc.GetRouter(),
		ReadTimeout:  2 * time.Second,  // Максимальное время на чтение запроса
		WriteTimeout: 4 * time.Second,  // Максимальное время на запись ответа
		IdleTimeout:  10 * time.Second, // Таймаут для keep-alive соединений
	}

	g.Go(func() error {
		logger.Log.Debug().Msg("Запуск HttpServer")
		err := server.ListenAndServe()
		if ctx.Err() == nil {
			//Ошибку отображаем только если контекст не завершен
			logger.Log.Error().Msg("Ошибка в работе сервера ListenAndServe:" + err.Error())
			return err
		}
		return nil
	})

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

	return nil
}
