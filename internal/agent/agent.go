package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/http"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/qesterrx/tplmetrics/internal/logger"
	"github.com/qesterrx/tplmetrics/internal/model"
	"github.com/qesterrx/tplmetrics/internal/retry"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

// Если честно - не понятно какие ошибки переотправлять а какие нет. Если это важно почему этому не научили или хотя бы не тыкнули носом во что-то полезное.
// Не магия конечно но так... пальцем в небо
var retryableErrors = []error{
	syscall.ECONNREFUSED, //Соединение отклонено
	syscall.ECONNRESET,   //Соединение сброшено
	syscall.ETIMEDOUT,    //Таймаут операции
	syscall.EHOSTUNREACH, //Хост недоступен
	syscall.ENETUNREACH,  //Сеть недоступна
	syscall.EPIPE,        //Разорванный канал
	syscall.EAGAIN,       //Ресурс временно недоступен
	syscall.EWOULDBLOCK,  //Ресурс временно недоступен
}

// Функция проверки ошибки на необходимость повтора
func checkRetryRequest(err error) bool {

	for _, v := range retryableErrors {
		if errors.Is(err, v) {
			//Повторяемая ошибка
			return true
		}
	}
	return false
}

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

			logger.Log.Debug().Msg("Collector Метрики собраны")
		}
	}

}

/*Мда...*/
func CollectorAdd(ctx context.Context, toGroup chan<- model.Metrica, pollInterval int) {

	logger.Log.Debug().Msg("Запуск CollectorAdd")

	ticker := time.NewTicker(time.Second * time.Duration(pollInterval))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			//Если получили сигнал завершения - останавливаемся
			logger.Log.Debug().Msg("Остановка CollectorAdd по контексту")
			return
		case <-ticker.C:
			mmInfo, err := mem.VirtualMemory()
			if err != nil {
				logger.Log.Error().Msg("Ошибка сборка метрик mem.VirtualMemory:" + err.Error())
			}

			cpuCount, err := cpu.Counts(true)
			if err != nil {
				logger.Log.Error().Msg("Ошибка сборка метрик cpu.Counts:" + err.Error())
			}

			toGroup <- model.NewMetricaGauge("TotalMemory", float64(mmInfo.Total))
			toGroup <- model.NewMetricaGauge("FreeMemory", float64(mmInfo.Free))
			toGroup <- model.NewMetricaGauge("CPUutilization1", float64(cpuCount))

			logger.Log.Debug().Msg("CollectorAdd Метрики собраны")
		}

	}

}

/*Процедура через reportInterval вычитывает очередь toGroup, группирует gauge метрики и ставит в очередь на отправку toSend в виде []byte*/
func Reporter(ctx context.Context, toGroup <-chan model.Metrica, toSend chan<- []byte, reportInterval int) {
	logger.Log.Debug().Msg("Запуск Reporter")

	ticker := time.NewTicker(time.Second * time.Duration(reportInterval))
	defer ticker.Stop()

	//в данной структуре будем группировать данные из очереди, кроме того с помощью нее обеспечим транзакционность
	groupMap := map[string]model.Metrica{}

	for {
		select {
		case <-ctx.Done():
			//Если получили сигнал завершения останавливаемся
			logger.Log.Debug().Msg("Остановка Reporter по контексту")
			return
		case <-ticker.C:

			logger.Log.Debug().Msg("Reporter запуск группировки данных из очереди toGroup")

		loop:
			for {
				select {
				case <-ctx.Done():
					//Если получили сигнал завершения останавливаемся
					logger.Log.Debug().Msg("Остановка Reporter по контексту")
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
								oldMetrica.Restore()
								logger.Log.Error().Msg("Ошибка обновления метрики " + err.Error())
								continue
							}
							oldMetrica.Confirm()
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

			logger.Log.Debug().Msg(fmt.Sprintf("Попытка отправить массив метрик, количество %d", len(mtrks)))

			//Отправка данных
			toSend <- body

			//... и каков ответ на главный вопрос жизни, вселенной и всего такого
			groupMap = map[string]model.Metrica{}

		}
	}

}

func Sender(ctx context.Context, toSend <-chan []byte, num int, url string, secretKeyForSign string) {
	logger.Log.Debug().Msg("Запуск Sender")

	//Клиента создаем один раз
	client := resty.New()

	//Тут определим middleware агента
	client.OnBeforeRequest(GzipCompressMiddleware) //Сначала зипуем
	if secretKeyForSign != "" {
		client.OnBeforeRequest(HMACSignMiddleware(secretKeyForSign)) //Затем подписываем
	}

	for {
		select {
		case <-ctx.Done():
			//Если получили сигнал завершения останавливаемся
			logger.Log.Debug().Msg("Остановка Sender по контексту")
			return
		case msg := <-toSend:
			err := Send(ctx, client, url, msg)

			if err != nil {
				logger.Log.Error().Msg(err.Error())
			} else {
				logger.Log.Debug().Msg("Метрики отправлены [Sender " + strconv.Itoa(num) + "]")
			}
		}
	}

}

/*Процедура отвечает только за отправку уже сериализованных данных*/
func Send(ctx context.Context, client *resty.Client, url string, body []byte) error {

	//замыкание для вызова в retry.RetryFunc
	fn := func() error {

		resp, err := client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(body).
			Post(url)

		if err != nil {
			//Получили ошибку при выполнении запроса
			return fmt.Errorf("sender ошибка выполнения запроса на сервер %w", err)
		}

		if resp.StatusCode() != http.StatusOK {
			//Получили от сервера код который не ожидали
			return fmt.Errorf("sender сервер не принял сообщение StatusCode!=OK")
		}

		return nil
	}

	//Выполнение с повтором
	return retry.RetryFunc(ctx, fn, checkRetryRequest, 3, 1*time.Second, 2*time.Second)

}

func GzipCompressMiddleware(c *resty.Client, r *resty.Request) error {
	if r.Body != nil {

		//r.Body это интерфейс, очередной type assertion
		var srcBody []byte
		switch tmp := r.Body.(type) {
		case string:
			srcBody = []byte(tmp)
		case []byte:
			srcBody = tmp
		default:
			//Если тело не то что мы предпологали то просто ничего не делаем
			return nil
		}

		var compressed bytes.Buffer

		gzWriter := gzip.NewWriter(&compressed)

		_, err := gzWriter.Write(srcBody)
		if err != nil {
			return fmt.Errorf("sender ошибка компрессии gzip %w", err)
		}

		// Важно! Закрываем writer, чтобы сбросить все данные в буфер - эх время мое время
		gzWriter.Close()

		//Добавляем заголовок, переписываем Body
		r.SetHeader("Content-Encoding", "gzip")
		r.SetBody(compressed.Bytes())

	}

	return nil

}

func HMACSignMiddleware(key string) resty.RequestMiddleware {

	secret := []byte(key)

	return func(c *resty.Client, r *resty.Request) error {
		if r.Body != nil {

			//r.Body это интерфейс - type assertion
			var srcBody []byte
			switch tmp := r.Body.(type) {
			case string:
				srcBody = []byte(tmp)
			case []byte:
				srcBody = tmp
			default:
				//Если тело не то что мы предпологали то просто ничего не делаем
				return nil
			}

			//считаем хешь, по идее нужна общая функция для клиента и сервереа но и таааак сойдет
			hash := hmac.New(sha256.New, secret)
			hash.Write(srcBody)
			sign := hash.Sum(nil)

			//Записываем заголовок
			r.SetHeader("HashSHA256", base64.StdEncoding.EncodeToString(sign))

		}

		return nil
	}

}

/*Вообще конечно бросается в глаза то что можно было обойтись одной middleware или вообще без них...*/
