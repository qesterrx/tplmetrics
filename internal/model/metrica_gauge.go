package model

import (
	"encoding/json"
	"fmt"
	"strings"
)

func FormatMetricaGauge(value float64) string {
	return strings.TrimRight(fmt.Sprintf("%.3f", value), "0")
}

type MetricaGauge struct {
	name  string
	kind  KindValue
	value float64
}

func NewMetricaGauge(name string, value float64) *MetricaGauge {
	return &MetricaGauge{
		name:  name,
		kind:  Gauge,
		value: value,
	}
}

func (m *MetricaGauge) Name() string {
	return m.name
}

func (m *MetricaGauge) Kind() KindValue {
	return m.kind
}

func (m *MetricaGauge) Value() string {
	return FormatMetricaGauge(m.value)
}

func (m *MetricaGauge) SrcValue() float64 {
	return m.value
}

func (m *MetricaGauge) UpdateValue(mtrk Metrica) error {
	if v, ok := mtrk.(*MetricaGauge); ok {
		m.value = v.value
		return nil
	} else {
		return fmt.Errorf("MetricaGauge.UpdateValue type mismatch: want MetricaGauge got %v", mtrk)
	}
}

func (m *MetricaGauge) MarshalJSON() ([]byte, error) {
	return json.Marshal(MetricaJSONAdapter{Name: m.name, Kind: string(m.kind), Value: &m.value})
}
