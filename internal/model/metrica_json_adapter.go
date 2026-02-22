package model

import "fmt"

// Структура для общения между агентом и сервером
type MetricaJSONAdapter struct {
	Name  string   `json:"id"`              // имя метрики
	Kind  string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (mtrk *MetricaJSONAdapter) Metrica() (Metrica, error) {
	kindValue, err := GetKindValue(mtrk.Kind)
	if err != nil {
		return nil, err
	}

	switch kindValue {
	case Counter:
		if mtrk.Delta == nil {
			return nil, fmt.Errorf("Не передано значение для Counter метрики")
		}
		return NewMetricaCounter(mtrk.Name, *mtrk.Delta), nil
	case Gauge:
		if mtrk.Value == nil {
			return nil, fmt.Errorf("Не передано значение для Gauge метрики")
		}

		return NewMetricaGauge(mtrk.Name, *mtrk.Value), nil
	}
	return nil, fmt.Errorf("NewMetrica unimplemented kind %s", mtrk.Kind)
}
