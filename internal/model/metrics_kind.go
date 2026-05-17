package model

import (
	"fmt"
	"strings"
)

// KindValue - Тип описывающий возможные варианты (типы) метрик
type KindValue string

const (
	//Gauge - Метрика содержащая значение типа float64, перетирает значение при каждом обновлении
	Gauge KindValue = "gauge"

	//Counter - Метрика содержащая значение типа int64, инкрементирует значение при обнолвении
	Counter KindValue = "counter"
)

// GetKindValue - функция преобразующая строку содержащую тип метрики в тип KindValue
func GetKindValue(kind string) (KindValue, error) {
	switch strings.ToLower(kind) {
	case "gauge":
		return Gauge, nil
	case "counter":
		return Counter, nil
	default:
		return "", fmt.Errorf("unknown type metric, got %s", kind)
	}
}
