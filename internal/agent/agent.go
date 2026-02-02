package agent

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/qesterrx/tplmetrics/internal/model"
)

/*
По заданию не ясно как надо делать этого клиента.
Все таки если мы опрашиваем метрики каждые N минут а отправляем реже - не понятно надо ли отправлять метрики gauge (мы же все равно будем сохранять толкьо последнее значение?)
*/
func MetricCollector(queue chan<- model.Metrica, pollInterval int) {

	counter := 0

	for {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		queue <- *model.NewGaugeMetrica("Alloc", float64(m.Alloc))
		queue <- *model.NewGaugeMetrica("BuckHashSys", float64(m.BuckHashSys))
		queue <- *model.NewGaugeMetrica("Frees", float64(m.Frees))
		queue <- *model.NewGaugeMetrica("GCCPUFraction", float64(m.GCCPUFraction))
		queue <- *model.NewGaugeMetrica("GCSys", float64(m.GCSys))
		queue <- *model.NewGaugeMetrica("HeapAlloc", float64(m.HeapAlloc))
		queue <- *model.NewGaugeMetrica("HeapIdle", float64(m.HeapIdle))
		queue <- *model.NewGaugeMetrica("HeapInuse", float64(m.HeapInuse))
		queue <- *model.NewGaugeMetrica("HeapObjects", float64(m.HeapObjects))
		queue <- *model.NewGaugeMetrica("HeapReleased", float64(m.HeapReleased))
		queue <- *model.NewGaugeMetrica("HeapSys", float64(m.HeapSys))
		queue <- *model.NewGaugeMetrica("LastGC", float64(m.LastGC))
		queue <- *model.NewGaugeMetrica("Lookups", float64(m.Lookups))
		queue <- *model.NewGaugeMetrica("MCacheInuse", float64(m.MCacheInuse))
		queue <- *model.NewGaugeMetrica("MCacheSys", float64(m.MCacheSys))
		queue <- *model.NewGaugeMetrica("MSpanInuse", float64(m.MSpanInuse))
		queue <- *model.NewGaugeMetrica("MSpanSys", float64(m.MSpanSys))
		queue <- *model.NewGaugeMetrica("Mallocs", float64(m.Mallocs))
		queue <- *model.NewGaugeMetrica("NextGC", float64(m.NextGC))
		queue <- *model.NewGaugeMetrica("NumForcedGC", float64(m.NumForcedGC))
		queue <- *model.NewGaugeMetrica("NumGC", float64(m.NumGC))
		queue <- *model.NewGaugeMetrica("OtherSys", float64(m.OtherSys))
		queue <- *model.NewGaugeMetrica("PauseTotalNs", float64(m.PauseTotalNs))
		queue <- *model.NewGaugeMetrica("StackInuse", float64(m.StackInuse))
		queue <- *model.NewGaugeMetrica("StackSys", float64(m.StackSys))
		queue <- *model.NewGaugeMetrica("Sys", float64(m.Sys))
		queue <- *model.NewGaugeMetrica("TotalAlloc", float64(m.TotalAlloc))

		queue <- *model.NewGaugeMetrica("RandomValue", float64(rand.ExpFloat64()))
		queue <- *model.NewCounterMetrica("PollCount", 1)

		counter++
		fmt.Printf("MetricCollector Counter = %d, len chan =%d\n", counter, len(queue))

		time.Sleep(time.Second * time.Duration(pollInterval))
	}

}

func Sender(queue <-chan model.Metrica, reportInterval int, host string) {

	client := resty.New()

	for {
		select {
		case metrica := <-queue:
			//fmt.Printf("Reader: получил %v\n", metrica)
			url := fmt.Sprintf("http://%s/update/%s/%s/%s", host, metrica.Kind, metrica.Name, metrica.GetMetricaValue())
			resp, err := client.R().
				SetHeader("Content-Type", "text/plain").
				Post(url)

			if err != nil {
				fmt.Printf("RESTY HTTP Error = %e\n", err)
				continue
			}

			if resp.StatusCode() != http.StatusOK {
				fmt.Printf("RESTY HTTP StatusCode = %d", resp.StatusCode())
				continue
			}

		default:
			fmt.Println("Канал пуст")
			time.Sleep(time.Second * time.Duration(reportInterval))
		}
	}

}
