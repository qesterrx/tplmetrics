package agent

import (
	"context"
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
func Collector(ctx context.Context, queue chan<- model.Metrica, pollInterval int) {

	counter := 0

	for {
		select {
		case <-ctx.Done():
			//Если получили сигнал завершения - останавливаемся
			return
		default:
			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			queue <- model.NewMetricaGauge("Alloc", float64(m.Alloc))
			queue <- model.NewMetricaGauge("BuckHashSys", float64(m.BuckHashSys))
			queue <- model.NewMetricaGauge("Frees", float64(m.Frees))
			queue <- model.NewMetricaGauge("GCCPUFraction", float64(m.GCCPUFraction))
			queue <- model.NewMetricaGauge("GCSys", float64(m.GCSys))
			queue <- model.NewMetricaGauge("HeapAlloc", float64(m.HeapAlloc))
			queue <- model.NewMetricaGauge("HeapIdle", float64(m.HeapIdle))
			queue <- model.NewMetricaGauge("HeapInuse", float64(m.HeapInuse))
			queue <- model.NewMetricaGauge("HeapObjects", float64(m.HeapObjects))
			queue <- model.NewMetricaGauge("HeapReleased", float64(m.HeapReleased))
			queue <- model.NewMetricaGauge("HeapSys", float64(m.HeapSys))
			queue <- model.NewMetricaGauge("LastGC", float64(m.LastGC))
			queue <- model.NewMetricaGauge("Lookups", float64(m.Lookups))
			queue <- model.NewMetricaGauge("MCacheInuse", float64(m.MCacheInuse))
			queue <- model.NewMetricaGauge("MCacheSys", float64(m.MCacheSys))
			queue <- model.NewMetricaGauge("MSpanInuse", float64(m.MSpanInuse))
			queue <- model.NewMetricaGauge("MSpanSys", float64(m.MSpanSys))
			queue <- model.NewMetricaGauge("Mallocs", float64(m.Mallocs))
			queue <- model.NewMetricaGauge("NextGC", float64(m.NextGC))
			queue <- model.NewMetricaGauge("NumForcedGC", float64(m.NumForcedGC))
			queue <- model.NewMetricaGauge("NumGC", float64(m.NumGC))
			queue <- model.NewMetricaGauge("OtherSys", float64(m.OtherSys))
			queue <- model.NewMetricaGauge("PauseTotalNs", float64(m.PauseTotalNs))
			queue <- model.NewMetricaGauge("StackInuse", float64(m.StackInuse))
			queue <- model.NewMetricaGauge("StackSys", float64(m.StackSys))
			queue <- model.NewMetricaGauge("Sys", float64(m.Sys))
			queue <- model.NewMetricaGauge("TotalAlloc", float64(m.TotalAlloc))

			queue <- model.NewMetricaGauge("RandomValue", float64(rand.ExpFloat64()))
			queue <- model.NewMetricaCounter("PollCount", 1)

			counter++

			//Метрики собраны, засыпаем на pollInterval
			time.Sleep(time.Second * time.Duration(pollInterval))
		}
	}

}

func Sender(ctx context.Context, queue <-chan model.Metrica, reportInterval int, host string, clientErrorCount int) {

	client := resty.New()
	clentErrorCounter := 0

	for {
		select {
		case <-ctx.Done():
			//Если получили сигнал завершения останавливаемся
			return
		case metrica := <-queue:

			if clentErrorCounter > clientErrorCount {
				//Если количество ошибок превысило лимит - завершаем работу
				fmt.Printf("Sender превышено допустимое количество ошибок отправки")
				return
			}

			url := fmt.Sprintf("http://%s/update/%s/%s/%s", host, metrica.Kind(), metrica.Name(), metrica.Value())
			resp, err := client.R().
				SetHeader("Content-Type", "text/plain").
				Post(url)

			if err != nil {
				//Получили ошибку при выполнении запроса
				clentErrorCounter++
				continue
			}

			if resp.StatusCode() != http.StatusOK {
				//Получили от сервера код который не ожидали
				clentErrorCounter++
				continue
			}

		default:
			//Канал пуст, отправили все что было в канале а знчит засыпаем на reportInterval секунд
			time.Sleep(time.Second * time.Duration(reportInterval))
		}
	}

}
