package model

import (
	"encoding/json"
	"fmt"
)

func FormatMetricaCounter(value int64) string {
	return fmt.Sprintf("%d", value)
}

type MetricaCounter struct {
	name     string
	kind     KindValue
	value    int64
	newValue *int64
}

func NewMetricaCounter(name string, value int64) *MetricaCounter {
	return &MetricaCounter{
		name:     name,
		kind:     Counter,
		value:    value,
		newValue: nil,
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

func (m *MetricaCounter) SrcValue(env string) int64 {
	if env == "new" && m.newValue != nil {
		return *m.newValue
	}
	return m.value
}

func (m *MetricaCounter) UpdateValue(mtrk Metrica) error {
	if v, ok := mtrk.(*MetricaCounter); ok {

		if m.newValue != nil {
			return fmt.Errorf("другой процесс уже обновляет метрику %s", m.name)
		} else {
			newValue := new(int64)
			*newValue = v.value + m.value
			m.newValue = newValue
			return nil
		}
	} else {
		return fmt.Errorf("MetricaCounter.UpdateValue type mismatch: want MetricaCounter got %v", mtrk)
	}
}

func (m *MetricaCounter) Restore() {
	if m.newValue != nil {
		m.newValue = nil
	}
}

func (m *MetricaCounter) Confirm() {
	if (m.newValue) != nil {
		m.value = *(m.newValue)
		m.newValue = nil
	}
}

func (m *MetricaCounter) MarshalJSON() ([]byte, error) {
	return json.Marshal(MetricaJSONAdapter{Name: m.name, Kind: string(m.kind), Delta: &m.value})
}
