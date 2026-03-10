package model

import (
	"encoding/json"
	"fmt"
)

func FormatMetricaCounter(value int64) string {
	return fmt.Sprintf("%d", value)
}

type MetricaCounter struct {
	name  string
	kind  KindValue
	value int64
}

func NewMetricaCounter(name string, value int64) *MetricaCounter {
	return &MetricaCounter{
		name:  name,
		kind:  Counter,
		value: value,
	}
}

func (m *MetricaCounter) Name() string {
	return m.name
}

func (m *MetricaCounter) Kind() KindValue {
	return m.kind
}

func (m *MetricaCounter) Value() string {
	return FormatMetricaCounter(m.value)
}

func (m *MetricaCounter) SrcValue() int64 {
	return m.value
}

func (m *MetricaCounter) UpdateValue(mtrk Metrica) error {
	if v, ok := mtrk.(*MetricaCounter); ok {
		m.value = m.value + v.value
		return nil
	} else {
		return fmt.Errorf("MetricaCounter.UpdateValue type mismatch: want MetricaCounter got %v", mtrk)
	}
}

func (m *MetricaCounter) MarshalJSON() ([]byte, error) {
	return json.Marshal(MetricaJSONAdapter{Name: m.name, Kind: string(m.kind), Delta: &m.value})
}
