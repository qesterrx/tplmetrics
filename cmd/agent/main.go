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

	wg := sync.WaitGroup{}

	//Сборщик записывает метрики в queueToGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		agent.Collector(ctx, queueToGroup, config.PoolInterval)
		cancel()
	}()

	//Репортер берет метрики из queueToGroup, группирует, сериализует и пытается отправить
	wg.Add(1)
	go func() {
		defer wg.Done()
		url := fmt.Sprintf("http://%s/updates/", config.ServerHost.String())
		agent.Reporter(ctx, queueToGroup, config.ReportInterval, url, config.SecretKeyForSign)
		cancel()
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	cancel()
	logger.Log.Debug().Msg("Ожидание завершения программы")
	wg.Wait()

}
