package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/qesterrx/tplmetrics/internal/logger"
	"github.com/qesterrx/tplmetrics/internal/model"
)

/*Процедура собирает метрики через интервал pollInterval, записывает в очередь queue*/
func Collector(ctx context.Context, toGroup chan<- model.Metrica, pollInterval int) {

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

			toGroup <- model.NewMetricaGauge("Alloc", float64(m.Alloc))
			toGroup <- model.NewMetricaGauge("BuckHashSys", float64(m.BuckHashSys))
			toGroup <- model.NewMetricaGauge("Frees", float64(m.Frees))
			toGroup <- model.NewMetricaGauge("GCCPUFraction", float64(m.GCCPUFraction))
			toGroup <- model.NewMetricaGauge("GCSys", float64(m.GCSys))
			toGroup <- model.NewMetricaGauge("HeapAlloc", float64(m.HeapAlloc))
			toGroup <- model.NewMetricaGauge("HeapIdle", float64(m.HeapIdle))
			toGroup <- model.NewMetricaGauge("HeapInuse", float64(m.HeapInuse))
			toGroup <- model.NewMetricaGauge("HeapObjects", float64(m.HeapObjects))
			toGroup <- model.NewMetricaGauge("HeapReleased", float64(m.HeapReleased))
			toGroup <- model.NewMetricaGauge("HeapSys", float64(m.HeapSys))
			toGroup <- model.NewMetricaGauge("LastGC", float64(m.LastGC))
			toGroup <- model.NewMetricaGauge("Lookups", float64(m.Lookups))
			toGroup <- model.NewMetricaGauge("MCacheInuse", float64(m.MCacheInuse))
			toGroup <- model.NewMetricaGauge("MCacheSys", float64(m.MCacheSys))
			toGroup <- model.NewMetricaGauge("MSpanInuse", float64(m.MSpanInuse))
			toGroup <- model.NewMetricaGauge("MSpanSys", float64(m.MSpanSys))
			toGroup <- model.NewMetricaGauge("Mallocs", float64(m.Mallocs))
			toGroup <- model.NewMetricaGauge("NextGC", float64(m.NextGC))
			toGroup <- model.NewMetricaGauge("NumForcedGC", float64(m.NumForcedGC))
			toGroup <- model.NewMetricaGauge("NumGC", float64(m.NumGC))
			toGroup <- model.NewMetricaGauge("OtherSys", float64(m.OtherSys))
			toGroup <- model.NewMetricaGauge("PauseTotalNs", float64(m.PauseTotalNs))
			toGroup <- model.NewMetricaGauge("StackInuse", float64(m.StackInuse))
			toGroup <- model.NewMetricaGauge("StackSys", float64(m.StackSys))
			toGroup <- model.NewMetricaGauge("Sys", float64(m.Sys))
			toGroup <- model.NewMetricaGauge("TotalAlloc", float64(m.TotalAlloc))

			toGroup <- model.NewMetricaGauge("RandomValue", float64(rand.ExpFloat64()))
			toGroup <- model.NewMetricaCounter("PollCount", 1)

			counter++

			logger.Log.Debug().Msg("Метрики собраны")
		}
	}

}

/*Процедура через reportInterval вычитывает очередь toGroup, группирует gauge метрики и ставит в очередь на отправку toSend в виде []byte*/
func Compressor(ctx context.Context, toGroup <-chan model.Metrica, toSend chan<- []byte, reportInterval int) {
	logger.Log.Debug().Msg("Запуск Compressor")

	ticker := time.NewTicker(time.Second * time.Duration(reportInterval))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			//Если получили сигнал завершения останавливаемся
			logger.Log.Debug().Msg("Остановка Compressor по контексту")
			return
		case <-ticker.C:

			logger.Log.Debug().Msg("Compressor запуск группировки данных из очереди toGroup")
			groupMap := map[string]model.Metrica{}

		loop:
			for {
				select {
				case <-ctx.Done():
					//Если получили сигнал завершения останавливаемся
					logger.Log.Debug().Msg("Остановка Compressor по контексту")
					return
				case metrica := <-toGroup:
					//Группировка
					switch metrica.Kind() {
					case model.Gauge:
						groupMap[metrica.Name()] = metrica
					case model.Counter:
						oldMetrica, ok := groupMap[metrica.Name()]
						if ok {
							err := oldMetrica.UpdateValue(metrica)
							if err != nil {
								logger.Log.Error().Msg("Ошибка обновления метрики " + err.Error())
								continue
							}
						} else {
							groupMap[metrica.Name()] = metrica
						}
					default:
						logger.Log.Error().Msg("Неизвестный тип метрики " + metrica.Name() + string(metrica.Kind()))
					}
				default:
					//Канал вычитан до конца - выходим из цикла
					break loop
				}
			}

			//Сериализация и ставим в очередь на отправку - Ну вот и пригодилось

			mtrks := []model.Metrica{}

			for _, metrica := range groupMap {
				mtrks = append(mtrks, metrica)
			}

			body, err := json.Marshal(mtrks)
			if err != nil {
				logger.Log.Error().Msg("Ошибка сериализации метрик " + err.Error())
				continue
			}

			logger.Log.Debug().Msg(fmt.Sprintf("В очередь на отправку добавлена массив метрик в количестве %d", len(mtrks)))
			toSend <- body

		}
	}

}

/*Процедура отвечает только за отправку уже сериализованных данных, отправляет сразу же как в toSend появляются данные*/
func Sender(ctx context.Context, url string, toSend <-chan []byte) {
	logger.Log.Debug().Msg("Запуск Sender")

	client := resty.New()
	clentErrorCounter := 0

	for {
		select {
		case <-ctx.Done():
			//Если получили сигнал завершения останавливаемся
			logger.Log.Debug().Msg("Остановка Sender по контексту")
			return
		case srcBody := <-toSend:
			var compressed bytes.Buffer

			gzWriter := gzip.NewWriter(&compressed)

			_, err := gzWriter.Write(srcBody)
			if err != nil {
				logger.Log.Error().Msg("Sender ошибка компрессии gzip " + err.Error())
				clentErrorCounter++
				continue
			}

			// Важно! Закрываем writer, чтобы сбросить все данные в буфер - эх время мое время
			gzWriter.Close()

			resp, err := client.R().
				SetHeader("Content-Encoding", "gzip").
				SetHeader("Content-Type", "application/json").
				SetBody(compressed.Bytes()).
				Post(url)

			if err != nil {
				//Получили ошибку при выполнении запроса
				logger.Log.Error().Msg("Sender ошибка выполнения запроса на сервер " + err.Error())
				clentErrorCounter++
				continue
			}

			if resp.StatusCode() != http.StatusOK {
				//Получили от сервера код который не ожидали
				logger.Log.Error().Msg("Sender сервер не принял сообщение StatusCode!=OK")
				clentErrorCounter++
				continue
			}

			logger.Log.Debug().Msg("Успешная отправка: " + string(srcBody))
		}
	}
}

/*Хм все таки насколько критично то что метрики не дошли до сервера?
Я ушел от сейвМапы потому что подход мне не нравился, хотя наверное он был и ничего
Сейчас при ошибке отправки я получается теряю запись о обновлении метрики
Может быть для Gauge это не критично, но вот Counter при ошибке во времени отправки сразу начнет брехать и со временем это может стать значимо
Все таки какой путь тут выбрать? очень нужен комментарий!*/
