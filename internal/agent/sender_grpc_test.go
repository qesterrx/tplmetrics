package agent

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/qesterrx/tplmetrics/internal/model"
	"github.com/stretchr/testify/require"
)

func TestSenderGRPC(t *testing.T) {
	tests := []struct {
		name        string
		metrics     []model.Metrica
		expectError bool
	}{
		{
			name: "send single gauge metric",
			metrics: []model.Metrica{
				model.NewMetricaGauge("cpu", 85.5),
			},
			expectError: false,
		},
		{
			name: "send single counter metric",
			metrics: []model.Metrica{
				model.NewMetricaCounter("requests", 100),
			},
			expectError: false,
		},
		{
			name: "send multiple metrics",
			metrics: []model.Metrica{
				model.NewMetricaGauge("cpu", 85.5),
				model.NewMetricaCounter("requests", 100),
				model.NewMetricaGauge("memory", 2048),
			},
			expectError: false,
		},
		{
			name:        "send empty metrics",
			metrics:     []model.Metrica{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем каналы
			toSend := make(chan []byte, 1)
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			// Сериализуем метрики
			data, err := json.Marshal(tt.metrics)
			require.NoError(t, err)

			//Клиент
			go func() {
				SenderGRPC(ctx, toSend, ":66666")
			}()

			toSend <- data
			close(toSend)

			// Ждем завершения
			select {
			case <-ctx.Done():
				t.Fatal("Test timeout")
			case <-time.After(1 * time.Second):
				// Ожидаем, что горутина завершится
			}
		})
	}
}
