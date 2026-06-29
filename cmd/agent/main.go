package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/qesterrx/tplmetrics/internal/agent"
	cfg "github.com/qesterrx/tplmetrics/internal/config"
	"github.com/qesterrx/tplmetrics/internal/logger"
	"github.com/qesterrx/tplmetrics/internal/model"
	"github.com/qesterrx/tplmetrics/pkg/defval"
	"github.com/rs/zerolog"
)

var buildVersion string
var buildDate string
var buildCommit string

func main() {

	logger.InitLogger()
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	logger.Log.Info().Str("Build version:", defval.DVR(buildVersion, "N/A")).Msg("")
	logger.Log.Info().Str("Build date:", defval.DVR(buildDate, "N/A")).Msg("")
	logger.Log.Info().Str("Build commit:", defval.DVR(buildCommit, "N/A")).Msg("")

	config, err := cfg.ParseParamsAgent()
	if err != nil {
		logger.Log.Fatal().Msg(err.Error())
	}

	//Загружаем публичный ключ
	err = config.LoadPublicKey()
	if err != nil {
		logger.Log.Fatal().Msg(err.Error())
	}

	RunAgent(config)
}

func RunAgent(config *cfg.ConfigAgent) {

	ctx, cancel := context.WithCancel(context.Background())
	queueToGroup := make(chan model.Metrica, 1000) //Количество ~= (reportInterval/pollInterval+1)*Количество метрик
	queueToSend := make(chan []byte, config.RateLimit)

	wg := sync.WaitGroup{}

	//Сборщик записывает метрики в queueToGroup
	wg.Go(func() {
		agent.Collector(ctx, queueToGroup, config.PoolInterval)
	})

	//Дополнительный сборщик записывает метрики в queueToGroup
	wg.Go(func() {
		agent.CollectorAdd(ctx, queueToGroup, config.PoolInterval)
	})

	//Репортер берет метрики из queueToGroup, группирует, сериализует и пытается отправить
	wg.Go(func() {
		agent.Reporter(ctx, queueToGroup, queueToSend, config.ReportInterval)
		//Надо закрыть канал, чтобы отправщики могли остановиться
		close(queueToSend)
	})

	switch config.Protocol {
	case string(cfg.ProtocolServerHTTP):
		//Пул сендеров занимается отправкой
		for i := 1; i <= config.RateLimit; i++ {
			wg.Go(func() {
				url := fmt.Sprintf("http://%s/updates/", config.ServerHost.String())
				agent.Sender(ctx, queueToSend, i, url, config.SecretKeyForSign, config.PublicKeyRSA)
			})
		}
	case string(cfg.ProtocolServerGRPC):
		wg.Go(func() {
			url := config.ServerHost.String()
			agent.SenderGRPC(ctx, queueToSend, url)
		})
	default:
		logger.Log.Fatal().Msg("Неизвестный вариант протокола")
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-sigChan
	cancel()

	logger.Log.Debug().Msg("Ожидание завершения программы")
	wg.Wait()

}
