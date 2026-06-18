package agent

import (
	"context"
	"testing"
	"time"

	"github.com/qesterrx/tplmetrics/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestCollector(t *testing.T) {
	tests := []struct {
		name         string
		pollInterval int
		waitDuration time.Duration
		expectedMin  int // минимальное ожидаемое количество метрик
	}{
		{
			name:         "collect metrics with interval 1 second",
			pollInterval: 1,
			waitDuration: 2 * time.Second,
			expectedMin:  29 * 2, //должны успеть собрать метрики два раза
		},
		{
			name:         "collect metrics with interval 2 seconds",
			pollInterval: 2,
			waitDuration: 3 * time.Second,
			expectedMin:  29, //успеваем собрать только один раз
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toGroup := make(chan model.Metrica, 100)
			ctx, cancel := context.WithCancel(context.Background())

			// Запускаем коллектор
			go Collector(ctx, toGroup, tt.pollInterval)

			// Ждем некоторое время
			time.Sleep(tt.waitDuration)

			// Останавливаем коллектор
			cancel()

			// Даем время на завершение
			time.Sleep(100 * time.Millisecond)
			close(toGroup)

			// Собираем все метрики
			var metrics []model.Metrica
			for metric := range toGroup {
				metrics = append(metrics, metric)
			}

			// Проверяем, что метрики собраны
			assert.NotEmpty(t, metrics, "Должны быть собраны метрики")
			assert.GreaterOrEqual(t, len(metrics), tt.expectedMin, "Количество метрик меньше ожидаемого")

		})
	}
}
