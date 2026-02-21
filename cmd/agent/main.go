package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/qesterrx/tplmetrics/internal/agent"
	"github.com/qesterrx/tplmetrics/internal/config"
	"github.com/qesterrx/tplmetrics/internal/model"
)

func main() {
	config, err := config.ParseParamsAgent()
	if err != nil {
		panic(err)
	}

	RunAgent(config)
}

func RunAgent(config *config.ConfigAgent) {

	ctx, cancel := context.WithCancel(context.Background())
	queueCnah := make(chan model.Metrica, 1000) //Количество ~= (reportInterval/pollInterval+1)*Количество метрик

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		agent.Collector(ctx, queueCnah, config.PoolInterval)
		cancel()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		agent.Sender(ctx, queueCnah, config.ReportInterval, config.ServerHost.String(), config.ClientErrorCount)
		cancel()
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	cancel()
	wg.Wait()

}
