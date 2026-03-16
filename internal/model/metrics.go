package model

import (
	"fmt"
	"strconv"
)

// Интерфейс метрики исползуемый в сторадже и хендлерах
type Metrica interface {
	Name() string
	Value() string
	Kind() KindValue
	UpdateValue(Metrica) error
	Restore()
	Confirm()
}

// Фабрика метрик
func NewMetrica(name string, kind string, value string) (Metrica, error) {

	kindValue, err := GetKindValue(kind)
	if err != nil {
		return nil, err
	}

	switch kindValue {
	case Counter:
		valueInt, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, err
		}

		return NewMetricaCounter(name, valueInt), nil
	case Gauge:
		valueFlt, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, err
		}

		return NewMetricaGauge(name, valueFlt), nil
	}
	return nil, fmt.Errorf("NewMetrica unimplemented kind %s", kind)
}
