package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/qesterrx/tplmetrics/internal/agent"
	"github.com/qesterrx/tplmetrics/internal/config"
	"github.com/qesterrx/tplmetrics/internal/logger"
	"github.com/qesterrx/tplmetrics/internal/model"
	"github.com/rs/zerolog"
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

	wg := sync.WaitGroup{}

	//Сборщик записывает метрики в queueToGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		agent.Collector(ctx, queueToGroup, config.PoolInterval)
		cancel()
	}()

	//Дополнительный сборщик записывает метрики в queueToGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		agent.CollectorAdd(ctx, queueToGroup, config.PoolInterval)
		cancel()
	}()

	//Репортер берет метрики из queueToGroup, группирует, сериализует и пытается отправить
	wg.Add(1)
	go func() {
		defer wg.Done()
		agent.Reporter(ctx, queueToGroup, queueToSend, config.ReportInterval)
		cancel()
	}()

	//Пул сендеров занимается отправкой
	for i := 1; i <= config.RateLimit; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			url := fmt.Sprintf("http://%s/updates/", config.ServerHost.String())
			agent.Sender(ctx, queueToSend, i, url, config.SecretKeyForSign)
			cancel()
		}()
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	cancel()
	logger.Log.Debug().Msg("Ожидание завершения программы")
	wg.Wait()

}
