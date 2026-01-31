package repository

import "fmt"

type KindValue string

const (
	Gauge   KindValue = "gauge"
	Counter KindValue = "counter"
)

type Repository interface {
	UpdateMetric(kind KindValue, name string, value string) error
	GetMetric(name string) (KindValue, string, error)
	Show()
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
