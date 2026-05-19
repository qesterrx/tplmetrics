package model

import "fmt"

// MetricaJSONAdapter - Структура для общения между агентом и сервером
// Необходимость структуры обоснована исходным ТЗ
type MetricaJSONAdapter struct {
	// Имя метрики
	Name string `json:"id"`

	// Тип метрики, параметр принимающий значение gauge или counter
	Kind string `json:"type"`

	// Значение метрики в случае передачи counter
	Delta *int64 `json:"delta,omitempty"`

	//Значение метрики в случае передачи gauge
	Value *float64 `json:"value,omitempty"`
}

// Metrica - метод преобразования экземпляра MetricaJSONAdapter в экземпляр Metrica соответствующего типа
func (mtrk *MetricaJSONAdapter) Metrica() (Metrica, error) {
	kindValue, err := GetKindValue(mtrk.Kind)
	if err != nil {
		return nil, err
	}

	switch kindValue {
	case Counter:
		if mtrk.Delta == nil {
			return nil, fmt.Errorf("не передано значение для Counter метрики")
		}
		return NewMetricaCounter(mtrk.Name, *mtrk.Delta), nil
	case Gauge:
		if mtrk.Value == nil {
			return nil, fmt.Errorf("не передано значение для Gauge метрики")
		}

		return NewMetricaGauge(mtrk.Name, *mtrk.Value), nil
	}
	return nil, fmt.Errorf("NewMetrica unimplemented kind %s", mtrk.Kind)
}
