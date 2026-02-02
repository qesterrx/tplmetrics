package main

import (
	"sync"

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

	queue := make(chan model.Metrica, 1000) //Количество ~= (reportInterval/pollInterval+1)*Количество метрик
	var wg sync.WaitGroup
	wg.Add(1)

	go agent.MetricCollector(queue, config.PoolInterval)
	go agent.Sender(queue, config.ReportInterval, config.ServerHost.String())

	wg.Wait()

}
