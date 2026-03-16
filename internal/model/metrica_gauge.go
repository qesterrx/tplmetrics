package model

import (
	"encoding/json"
	"fmt"
	"strings"
)

func FormatMetricaGauge(value float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.3f", value), "0"), ".")
}

type MetricaGauge struct {
	name     string
	kind     KindValue
	value    float64
	newValue *float64
}

func NewMetricaGauge(name string, value float64) *MetricaGauge {
	return &MetricaGauge{
		name:     name,
		kind:     Gauge,
		value:    value,
		newValue: nil,
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

func (m *MetricaGauge) SrcValue(env string) float64 {
	if env == "new" && m.newValue != nil {
		return *m.newValue
	}
	return m.value
}

func (m *MetricaGauge) UpdateValue(mtrk Metrica) error {
	if v, ok := mtrk.(*MetricaGauge); ok {

		if m.newValue != nil {
			return fmt.Errorf("другой процесс уже обновляет метрику %s", m.name)
		} else {
			newValue := new(float64)
			*newValue = v.value
			m.newValue = newValue
			return nil
		}
	} else {
		return fmt.Errorf("MetricaGauge.UpdateValue type mismatch: want MetricaGauge got %v", mtrk)
	}
}

func (m *MetricaGauge) Restore() {
	if m.newValue != nil {
		m.newValue = nil
	}
}

func (m *MetricaGauge) Confirm() {
	if m.newValue != nil {
		m.value = *(m.newValue)
		m.newValue = nil
	}
}

func (m *MetricaGauge) MarshalJSON() ([]byte, error) {
	return json.Marshal(MetricaJSONAdapter{Name: m.name, Kind: string(m.kind), Value: &m.value})
}
