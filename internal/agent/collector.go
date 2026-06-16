package agent

import (
	"context"
	"math/rand/v2"
	"runtime"
	"time"

	"github.com/qesterrx/tplmetrics/internal/logger"
	"github.com/qesterrx/tplmetrics/internal/model"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

// Collector Процедура(горутина) собирает основные метрики через переданный интервал pollInterval, записывает в очередь toGroup
func Collector(ctx context.Context, toGroup chan<- model.Metrica, pollInterval int) {

	logger.Log.Debug().Msg("Запуск Collector")

	ticker := time.NewTicker(time.Second * time.Duration(pollInterval))
	defer ticker.Stop()

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

			logger.Log.Debug().Msg("Collector Метрики собраны")
		}
	}

}

// CollectorAdd Процедура(горутина) - собирает дополнительные метрики через pollInterval, отправляет в очередь toGroup
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
