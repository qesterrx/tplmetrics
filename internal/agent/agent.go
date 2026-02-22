package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/qesterrx/tplmetrics/internal/logger"
	"github.com/qesterrx/tplmetrics/internal/model"
)

var storage *SafeMap = NewSafeMap()

/*
По заданию не ясно как надо делать этого клиента.
Все таки если мы опрашиваем метрики каждые N минут а отправляем реже - не понятно надо ли отправлять метрики gauge (мы же все равно будем сохранять толкьо последнее значение?)
*/
func Collector(ctx context.Context, pollInterval int) {

	logger.Log.Debug().Msg("Запуск Collector")

	ticker := time.NewTicker(time.Second * time.Duration(pollInterval))
	defer ticker.Stop()

	counter := 0

	for {
		select {
		case <-ctx.Done():
			//Если получили сигнал завершения - останавливаемся
			logger.Log.Debug().Msg("Остановка Collector по контексту")
			return
		case <-ticker.C:
			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			storage.SetValue("Alloc", float64(m.Alloc))
			storage.SetValue("BuckHashSys", float64(m.BuckHashSys))
			storage.SetValue("Frees", float64(m.Frees))
			storage.SetValue("GCCPUFraction", float64(m.GCCPUFraction))
			storage.SetValue("GCSys", float64(m.GCSys))
			storage.SetValue("HeapAlloc", float64(m.HeapAlloc))
			storage.SetValue("HeapIdle", float64(m.HeapIdle))
			storage.SetValue("HeapInuse", float64(m.HeapInuse))
			storage.SetValue("HeapObjects", float64(m.HeapObjects))
			storage.SetValue("HeapReleased", float64(m.HeapReleased))
			storage.SetValue("HeapSys", float64(m.HeapSys))
			storage.SetValue("LastGC", float64(m.LastGC))
			storage.SetValue("Lookups", float64(m.Lookups))
			storage.SetValue("MCacheInuse", float64(m.MCacheInuse))
			storage.SetValue("MCacheSys", float64(m.MCacheSys))
			storage.SetValue("MSpanInuse", float64(m.MSpanInuse))
			storage.SetValue("MSpanSys", float64(m.MSpanSys))
			storage.SetValue("Mallocs", float64(m.Mallocs))
			storage.SetValue("NextGC", float64(m.NextGC))
			storage.SetValue("NumForcedGC", float64(m.NumForcedGC))
			storage.SetValue("NumGC", float64(m.NumGC))
			storage.SetValue("OtherSys", float64(m.OtherSys))
			storage.SetValue("PauseTotalNs", float64(m.PauseTotalNs))
			storage.SetValue("StackInuse", float64(m.StackInuse))
			storage.SetValue("StackSys", float64(m.StackSys))
			storage.SetValue("Sys", float64(m.Sys))
			storage.SetValue("TotalAlloc", float64(m.TotalAlloc))

			storage.SetValue("RandomValue", float64(rand.ExpFloat64()))

			storage.AddDelta("PollCount", 1)

			counter++

			logger.Log.Debug().Msg("Метрики собраны")
		}
	}

}

func CallServer(client *resty.Client, host string, metrica model.Metrica) error {

	body, err := json.Marshal(metrica)
	if err != nil {
		return fmt.Errorf("CallServer ошибка сериализация метрики %s", err.Error())
	}

	var compressed bytes.Buffer

	gzWriter := gzip.NewWriter(&compressed)

	_, err = gzWriter.Write(body)
	if err != nil {
		log.Fatal("Error writing to gzip:", err)
	}

	// Важно! Закрываем writer, чтобы сбросить все данные в буфер - эх время мое время
	gzWriter.Close()

	url := fmt.Sprintf("http://%s/update/", host)
	resp, err := client.R().
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Content-Type", "application/json").
		SetBody(compressed.Bytes()).
		Post(url)

	if err != nil {
		//Получили ошибку при выполнении запроса
		return fmt.Errorf("CallServer ошибка выполнения запроса %s", err.Error())
	}

	if resp.StatusCode() != http.StatusOK {
		//Получили от сервера код который не ожидали
		return fmt.Errorf("CallServer сервер не принял сообщение StatusCode!=OK")
	}

	return nil
}

func Sender(ctx context.Context, reportInterval int, host string, clientErrorCount int) {

	logger.Log.Debug().Msg("Запуск Sender")

	ticker := time.NewTicker(time.Second * time.Duration(reportInterval))
	defer ticker.Stop()

	client := resty.New()
	clentErrorCounter := 0

	for {
		select {
		case <-ctx.Done():
			//Если получили сигнал завершения останавливаемся
			logger.Log.Debug().Msg("Остановка Sender по контексту")
			return
		case <-ticker.C:

			keys := storage.GetKeys()
			for _, key := range keys {

				if clentErrorCounter > clientErrorCount {
					//Если количество ошибок превысило лимит - завершаем работу
					logger.Log.Error().Msg("Sender превышено допустимое количество ошибок отправки")
					return
				}

				//Метрики типа Value-Gauge
				value, ok := storage.GetValue(key)
				if ok {
					mtrk := model.NewMetricaGauge(key, value)
					err := CallServer(client, host, mtrk)
					if err != nil {
						logger.Log.Error().Msg(fmt.Sprintf("Sender MetricaGauge ошибка обращения к серверу: %s", err.Error()))
						clentErrorCounter++
					}
				}

				//Метрики типа Delta-Counter Все приседание ради них
				delta, ok := storage.GetDelta(key)
				if ok {
					mtrk := model.NewMetricaCounter(key, delta)
					err := CallServer(client, host, mtrk)
					if err != nil {
						logger.Log.Error().Msg(fmt.Sprintf("Sender MetricaCounter ошибка обращения к серверу: %s", err.Error()))
						clentErrorCounter++
					}
				}

				//Вдруг у нас много метрик но пришла команда завершения по контексту
				select {
				case <-ctx.Done():
					logger.Log.Error().Msg("Остановка Sender цикла по контексту")
					return
				default:
				}

			}

			logger.Log.Debug().Msg("Sender все метрики отправлены")

		}
	}
}
