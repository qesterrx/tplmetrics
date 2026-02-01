package model

import (
	"fmt"
)

type KindValue string

const (
	Gauge   KindValue = "gauge"
	Counter KindValue = "counter"
)

type Metrica struct {
	Name  string    `json:"name"`
	Kind  KindValue `json:"kind"`
	Value any       `json:"value"`
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

func NewGaugeMetrica(name string, value float64) *Metrica {
	return &Metrica{
		Name:  name,
		Kind:  Gauge,
		Value: value,
	}
}

func NewCounterMetrica(name string, value int64) *Metrica {
	return &Metrica{
		Name:  name,
		Kind:  Counter,
		Value: value,
	}
}

func (m *Metrica) GetMetricaValue() string {
	switch m.Kind {
	case Gauge:
		return fmt.Sprintf("%.3f", m.Value)
	case Counter:
		return fmt.Sprintf("%d", m.Value)
	default:
		return fmt.Sprintf("%v", m.Value)
	}

}
