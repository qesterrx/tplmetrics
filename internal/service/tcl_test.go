package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/qesterrx/tplmetrics/internal/config"
	"github.com/qesterrx/tplmetrics/internal/model"
	"github.com/qesterrx/tplmetrics/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestTCLService_UpdateMetrica(t *testing.T) {
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

func TestTCLService_UpdateMetricaBatch(t *testing.T) {
	tests := []struct {
		name          string
		metrics       []model.Metrica
		mockSetup     func(*mocks.MockMetricaStorage)
		expectedError bool
	}{
		{
			name: "successful batch update",
			metrics: []model.Metrica{
				model.NewMetricaGauge("cpu", 85.5),
				model.NewMetricaCounter("requests", 100),
			},
			mockSetup: func(mock *mocks.MockMetricaStorage) {
				mock.EXPECT().
					UpdateMetricaBatch(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedError: false,
		},
		{
			name: "batch update with storage error",
			metrics: []model.Metrica{
				model.NewMetricaGauge("memory", 2048),
			},
			mockSetup: func(mock *mocks.MockMetricaStorage) {
				mock.EXPECT().
					UpdateMetricaBatch(gomock.Any(), gomock.Any()).
					Return(errors.New("batch storage error"))
			},
			expectedError: true,
		},
		{
			name:    "empty batch",
			metrics: []model.Metrica{},
			mockSetup: func(mock *mocks.MockMetricaStorage) {
				mock.EXPECT().
					UpdateMetricaBatch(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := mocks.NewMockMetricaStorage(ctrl)
			tt.mockSetup(mockStorage)

			service := &TCLService{
				config:  &config.ConfigServer{},
				storage: mockStorage,
				subs:    []UpdateMetricaSubscriber{},
			}

			err := service.UpdateMetricaBatch(context.Background(), tt.metrics)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTCLService_GetMetrica(t *testing.T) {
	tests := []struct {
		name          string
		metricName    string
		metricKind    string
		expectedValue model.Metrica
		expectedError error
		mockSetup     func(*mocks.MockMetricaStorage)
	}{
		{
			name:          "get existing gauge metric",
			metricName:    "cpu",
			metricKind:    "gauge",
			expectedValue: model.NewMetricaGauge("cpu", 85.5),
			expectedError: nil,
			mockSetup: func(mock *mocks.MockMetricaStorage) {
				mock.EXPECT().
					GetMetrica(gomock.Any(), "cpu", "gauge").
					Return(model.NewMetricaGauge("cpu", 85.5), nil)
			},
		},
		{
			name:          "get existing counter metric",
			metricName:    "requests",
			metricKind:    "counter",
			expectedValue: model.NewMetricaCounter("requests", 100),
			expectedError: nil,
			mockSetup: func(mock *mocks.MockMetricaStorage) {
				mock.EXPECT().
					GetMetrica(gomock.Any(), "requests", "counter").
					Return(model.NewMetricaCounter("requests", 100), nil)
			},
		},
		{
			name:          "get non-existing metric",
			metricName:    "nonexistent",
			metricKind:    "gauge",
			expectedValue: nil,
			expectedError: errors.New("metric not found"),
			mockSetup: func(mock *mocks.MockMetricaStorage) {
				mock.EXPECT().
					GetMetrica(gomock.Any(), "nonexistent", "gauge").
					Return(nil, errors.New("metric not found"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := mocks.NewMockMetricaStorage(ctrl)
			tt.mockSetup(mockStorage)

			service := &TCLService{
				config:  &config.ConfigServer{},
				storage: mockStorage,
				subs:    []UpdateMetricaSubscriber{},
			}

			result, err := service.GetMetrica(context.Background(), tt.metricName, tt.metricKind)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue.Name(), result.Name())
				assert.Equal(t, tt.expectedValue.Kind(), result.Kind())
			}
		})
	}
}

func TestTCLService_GetAllMetrics(t *testing.T) {
	tests := []struct {
		name          string
		expectedCount int
		mockSetup     func(*mocks.MockMetricaStorage)
	}{
		{
			name:          "get all metrics with data",
			expectedCount: 2,
			mockSetup: func(mock *mocks.MockMetricaStorage) {
				mock.EXPECT().
					GetAllMetrics(gomock.Any()).
					Return([]model.Metrica{
						model.NewMetricaGauge("cpu", 85.5),
						model.NewMetricaCounter("requests", 100),
					})
			},
		},
		{
			name:          "get all metrics empty",
			expectedCount: 0,
			mockSetup: func(mock *mocks.MockMetricaStorage) {
				mock.EXPECT().
					GetAllMetrics(gomock.Any()).
					Return([]model.Metrica{})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := mocks.NewMockMetricaStorage(ctrl)
			tt.mockSetup(mockStorage)

			service := &TCLService{
				config:  &config.ConfigServer{},
				storage: mockStorage,
				subs:    []UpdateMetricaSubscriber{},
			}

			result := service.GetAllMetrics(context.Background())
			assert.Len(t, result, tt.expectedCount)
		})
	}
}

func TestTCLService_TickerWriteMetrics(t *testing.T) {
	tests := []struct {
		name          string
		storeInterval int
		mockSetup     func(*mocks.MockMetricaStorage)
		expectCalls   int
	}{
		{
			name:          "ticker writes metrics on interval",
			storeInterval: 1,
			mockSetup: func(mock *mocks.MockMetricaStorage) {
				mock.EXPECT().
					WriteMetrics(gomock.Any()).
					Return(nil).
					MinTimes(1)
			},
			expectCalls: 1,
		},
		{
			name:          "ticker handles storage error",
			storeInterval: 1,
			mockSetup: func(mock *mocks.MockMetricaStorage) {
				mock.EXPECT().
					WriteMetrics(gomock.Any()).
					Return(errors.New("write error")).
					MinTimes(1)
			},
			expectCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStorage := mocks.NewMockMetricaStorage(ctrl)
			tt.mockSetup(mockStorage)

			service := &TCLService{
				config: &config.ConfigServer{
					StoreInterval: tt.storeInterval,
				},
				storage: mockStorage,
				subs:    []UpdateMetricaSubscriber{},
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			// Запускаем в горутине
			go service.TickerWriteMetrics(ctx)

			// Даем время на выполнение
			time.Sleep(1500 * time.Millisecond)
			cancel()

			// Небольшая задержка для завершения горутины
			time.Sleep(100 * time.Millisecond)
		})
	}
}

func TestTCLService_AddUpdateMetricaSubscriber(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockMetricaStorage(ctrl)
	mockSub1 := mocks.NewMockUpdateMetricaSubscriber(ctrl)
	mockSub2 := mocks.NewMockUpdateMetricaSubscriber(ctrl)

	service := &TCLService{
		config:  &config.ConfigServer{},
		storage: mockStorage,
		subs:    []UpdateMetricaSubscriber{},
	}

	// Добавляем первого подписчика
	service.AddUpdateMetricaSubscriber(mockSub1)
	assert.Len(t, service.subs, 1)

	// Добавляем второго подписчика
	service.AddUpdateMetricaSubscriber(mockSub2)
	assert.Len(t, service.subs, 2)

	// Пытаемся добавить дубликат
	service.AddUpdateMetricaSubscriber(mockSub1)
	assert.Len(t, service.subs, 2) // Длина не должна измениться

	// Пытаемся добавить nil
	service.AddUpdateMetricaSubscriber(nil)
	assert.Len(t, service.subs, 2) // Длина не должна измениться
}
