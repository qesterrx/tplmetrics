package agent

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/qesterrx/tplmetrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReporter(t *testing.T) {
	tests := []struct {
		name           string
		metrics        []model.Metrica
		reportInterval int
		expectedCount  int
		expectedTypes  map[string]string // имя -> тип
	}{
		{
			name: "single gauge metric",
			metrics: []model.Metrica{
				model.NewMetricaGauge("cpu", 85.5),
			},
			reportInterval: 1,
			expectedCount:  1,
			expectedTypes: map[string]string{
				"cpu": "gauge",
			},
		},
		{
			name: "single counter metric",
			metrics: []model.Metrica{
				model.NewMetricaCounter("requests", 100),
			},
			reportInterval: 1,
			expectedCount:  1,
			expectedTypes: map[string]string{
				"requests": "counter",
			},
		},
		{
			name: "multiple different metrics",
			metrics: []model.Metrica{
				model.NewMetricaGauge("cpu", 85.5),
				model.NewMetricaCounter("requests", 100),
				model.NewMetricaGauge("memory", 2048),
			},
			reportInterval: 1,
			expectedCount:  3,
			expectedTypes: map[string]string{
				"cpu":      "gauge",
				"requests": "counter",
				"memory":   "gauge",
			},
		},
		{
			name: "duplicate gauge metrics - last wins",
			metrics: []model.Metrica{
				model.NewMetricaGauge("cpu", 85.5),
				model.NewMetricaGauge("cpu", 95.5),
				model.NewMetricaGauge("cpu", 75.5),
			},
			reportInterval: 1,
			expectedCount:  1,
			expectedTypes: map[string]string{
				"cpu": "gauge",
			},
		},
		{
			name: "duplicate counter metrics - sum",
			metrics: []model.Metrica{
				model.NewMetricaCounter("requests", 100),
				model.NewMetricaCounter("requests", 50),
				model.NewMetricaCounter("requests", 25),
			},
			reportInterval: 1,
			expectedCount:  1,
			expectedTypes: map[string]string{
				"requests": "counter",
			},
		},
		{
			name: "mixed duplicate metrics",
			metrics: []model.Metrica{
				model.NewMetricaGauge("cpu", 85.5),
				model.NewMetricaCounter("requests", 100),
				model.NewMetricaGauge("cpu", 95.5),
				model.NewMetricaCounter("requests", 50),
				model.NewMetricaGauge("memory", 2048),
			},
			reportInterval: 1,
			expectedCount:  3,
			expectedTypes: map[string]string{
				"cpu":      "gauge",
				"requests": "counter",
				"memory":   "gauge",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toGroup := make(chan model.Metrica, len(tt.metrics))
			toSend := make(chan []byte, 1)
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			// Запускаем Reporter
			go Reporter(ctx, toGroup, toSend, tt.reportInterval)

			// Отправляем метрики
			for _, m := range tt.metrics {
				toGroup <- m
			}

			// Ждем отправки данных
			select {
			case data := <-toSend:
				// Проверяем полученные данные
				var metrics []model.MetricaJSONAdapter
				err := json.Unmarshal(data, &metrics)
				require.NoError(t, err, "Failed to unmarshal metrics")

				assert.Equal(t, tt.expectedCount, len(metrics), "Wrong number of metrics")

				// Проверяем типы метрик
				for _, m := range metrics {
					expectedType, ok := tt.expectedTypes[m.Name]
					assert.True(t, ok, "Unexpected metric: %s", m.Name)
					assert.Equal(t, expectedType, string(m.Kind), "Wrong type for metric: %s", m.Name)
				}

			case <-time.After(3 * time.Second):
				t.Fatal("Timeout waiting for data")
			}
		})
	}
}
