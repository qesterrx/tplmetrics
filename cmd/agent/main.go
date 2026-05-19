package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/qesterrx/tplmetrics/internal/agent"
	"github.com/qesterrx/tplmetrics/internal/config"
	"github.com/qesterrx/tplmetrics/internal/logger"
	"github.com/qesterrx/tplmetrics/internal/model"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

func main() {
	logger.InitLogger()
	zerolog.SetGlobalLevel(zerolog.DebugLevel) //Этот левел для меня )

	config, err := config.ParseParamsAgent()
	if err != nil {
		panic(err)
	}

	RunAgent(config)
}

func RunAgent(config *config.ConfigAgent) {

	ctx, cancel := context.WithCancel(context.Background())
	queueToGroup := make(chan model.Metrica, 1000) //Количество ~= (reportInterval/pollInterval+1)*Количество метрик
	queueToSend := make(chan []byte, config.RateLimit)

	g, ctx := errgroup.WithContext(ctx)

	//Сборщик записывает метрики в queueToGroup
	g.Go(func() error {
		agent.Collector(ctx, queueToGroup, config.PoolInterval)
		return nil
	})

	//Дополнительный сборщик записывает метрики в queueToGroup
	g.Go(func() error {
		agent.CollectorAdd(ctx, queueToGroup, config.PoolInterval)
		return nil
	})

	//Репортер берет метрики из queueToGroup, группирует, сериализует и пытается отправить
	g.Go(func() error {
		agent.Reporter(ctx, queueToGroup, queueToSend, config.ReportInterval)
		return nil
	})

	//Пул сендеров занимается отправкой
	for i := 1; i <= config.RateLimit; i++ {
		g.Go(func() error {
			url := fmt.Sprintf("http://%s/updates/", config.ServerHost.String())
			agent.Sender(ctx, queueToSend, i, url, config.SecretKeyForSign)
			return nil
		})
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	cancel()
	logger.Log.Debug().Msg("Ожидание завершения программы")

}
