package model

import (
	"fmt"
	"strconv"
)

type KindValue string

const (
	Gauge   KindValue = "gauge"
	Counter KindValue = "counter"
)

type Metrica interface {
	Name() string
	Value() string
	Kind() KindValue
	UpdateValue(Metrica) error
}

func GetKindValue(kind string) (KindValue, error) {
	switch kind {
	case "gauge":
		return Gauge, nil
	case "counter":
		return Counter, nil
	default:
		return "", fmt.Errorf("unknown type metric, got %s", kind)
	}
}

// Фабрика метрик
func NewMetrica(name string, kind KindValue, value string) (Metrica, error) {
	switch {
	case kind == Counter:
		valueInt, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, err
		}

		return NewMetricaCounter(name, valueInt), nil
	case kind == Gauge:
		valueFlt, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, err
		}

		return NewMetricaGauge(name, valueFlt), nil
	}
	return nil, fmt.Errorf("NewMetrica unimplemented kind %s", kind)
}
