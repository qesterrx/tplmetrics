package model

import (
	"fmt"
	"strings"
)

type KindValue string

const (
	Gauge   KindValue = "gauge"
	Counter KindValue = "counter"
)

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
