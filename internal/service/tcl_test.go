package service

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/qesterrx/tplmetrics/internal/config"
	"github.com/qesterrx/tplmetrics/internal/model"
	"github.com/qesterrx/tplmetrics/internal/service/mocks"
	"github.com/stretchr/testify/assert"
)

func TestTCLService_WithMock(t *testing.T) {
	tests := []struct {
		name          string
		metric        model.Metrica
		mockSetup     func(*mocks.MockMetricaStorage)
		expectedError bool
	}{
		{
			name:   "successful update gauge metric",
			metric: model.NewMetricaGauge("cpu", 85.5),
			mockSetup: func(mock *mocks.MockMetricaStorage) {
				mock.EXPECT().
					UpdateMetrica(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedError: false,
		},
		{
			name:   "successful update counter metric",
			metric: model.NewMetricaCounter("requests", 100),
			mockSetup: func(mock *mocks.MockMetricaStorage) {
				mock.EXPECT().
					UpdateMetrica(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedError: false,
		},
		{
			name:   "update with storage error",
			metric: model.NewMetricaGauge("memory", 2048),
			mockSetup: func(mock *mocks.MockMetricaStorage) {
				mock.EXPECT().
					UpdateMetrica(gomock.Any(), gomock.Any()).
					Return(errors.New("storage error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем контроллер
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Создаем мок
			mockStorage := mocks.NewMockMetricaStorage(ctrl)

			// Настраиваем мок
			tt.mockSetup(mockStorage)

			// Создаем сервис
			service := &TCLService{
				config:  &config.ConfigServer{},
				storage: mockStorage,
				subs:    []UpdateMetricaSubscriber{},
			}

			// Вызываем метод
			err := service.UpdateMetrica(context.Background(), tt.metric)

			// Проверяем результат
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
