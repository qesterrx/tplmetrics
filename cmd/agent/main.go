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
	queueToSend := make(chan []byte, 1000)         //Количество ~= Количество метрик*2

	wg := sync.WaitGroup{}

	//Сборщик записывает метрики в queueToGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		agent.Collector(ctx, queueToGroup, config.PoolInterval)
		cancel()
	}()

	//Группиратор берет метрики в queueToGroup, группирует, сериализует и записывает в queueToSend в виде []byte
	wg.Add(1)
	go func() {
		defer wg.Done()
		agent.Compressor(ctx, queueToGroup, queueToSend, config.ReportInterval)
		cancel()
	}()

	//Отправщик, работает со слайсом байт, ему все равно что отправлять, отправляет сразу же как только появился элемент в queueToSend
	wg.Add(1)
	go func() {
		defer wg.Done()
		url := fmt.Sprintf("http://%s/update/", config.ServerHost.String())
		agent.Sender(ctx, url, queueToSend)
		cancel()
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	cancel()
	logger.Log.Debug().Msg("Ожидание завершения программы")
	wg.Wait()

}
