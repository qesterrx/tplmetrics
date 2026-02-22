package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/qesterrx/tplmetrics/internal/agent"
	"github.com/qesterrx/tplmetrics/internal/config"
	"github.com/qesterrx/tplmetrics/internal/logger"
)

func main() {
	logger.InitLogger()

	config, err := config.ParseParamsAgent()
	if err != nil {
		panic(err)
	}

	RunAgent(config)
}

func RunAgent(config *config.ConfigAgent) {

	ctx, cancel := context.WithCancel(context.Background())

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		agent.Collector(ctx, config.PoolInterval)
		cancel()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		agent.Sender(ctx, config.ReportInterval, config.ServerHost.String(), config.ClientErrorCount)
		cancel()
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	cancel()
	logger.Log.Debug().Msg("Ожидание завершения программы")
	wg.Wait()

}
