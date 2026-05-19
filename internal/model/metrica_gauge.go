package model

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FormatMetricaGauge - Функция преобразования исходного значения Gauge метрики (float64) в строку
func FormatMetricaGauge(value float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.3f", value), "0"), ".")
}

// MetricaGauge - структура для хранения Gauge метрики
type MetricaGauge struct {
	name     string
	kind     KindValue
	value    float64
	newValue *float64
}

// NewMetricaGauge - Функция возвращающая ссылку на экземпляр MetricaGauge с переданными именем/значением
func NewMetricaGauge(name string, value float64) *MetricaGauge {
	return &MetricaGauge{
		name:     name,
		kind:     Gauge,
		value:    value,
		newValue: nil,
	}
}

// Name - метод возвращающий имя Gauge метрики
func (m *MetricaGauge) Name() string {
	return m.name
}

// Kind - метод возвращающий тип Gauge метрики в виде KindValue
func (m *MetricaGauge) Kind() KindValue {
	return m.kind
}

// Value - метод возвращающий значение Gauge метрики в виде строки
// Для преобразования float64 в строку используется функция [FormatMetricaGauge]
func (m *MetricaGauge) Value() string {
	return FormatMetricaGauge(m.value)
}

// SrcValue - Метод возвращающий исходное значение метрики (тип float64)
// Данный метод отсутствует в интерфейсе Metrica но необходим  для сохранения данных в Postgresql т.к данные в БД нужно сохранять в исходном типе
// Кроме того с помощью этого метода реализовывается транзационность в сервисном слое
// На вход получает переменную env указывающую какое именно значение необходимо вернуть - подтвержденное или не подтвержденное
func (m *MetricaGauge) SrcValue(env string) float64 {
	if env == "new" && m.newValue != nil {
		return *m.newValue
	}
	return m.value
}

// UpdateValueAtomic - Метод изменяющий значение Gauge метрики в синхронном режиме
func (m *MetricaGauge) UpdateValueAtomic(mtrk Metrica) error {

	if m.name != mtrk.Name() {
		return fmt.Errorf("MetricaGauge.UpdateValueAtomic поптыка обновить метрику с несовпадающим наименованием ")
	}

	if v, ok := mtrk.(*MetricaGauge); ok {
		m.value = v.value
		return nil
	} else {
		return fmt.Errorf("MetricaCounter.UpdateValue type mismatch: want MetricaCounter got %v", mtrk)
	}
}

// UpdateValueStart - Метод асинхронного измененения значение Gauge метрики
// Возвращает ошибку в случае если другой процесс обновляет выбранную метрику в асинхронном режиме
func (m *MetricaGauge) UpdateValueStart(mtrk Metrica) error {

	if m.name != mtrk.Name() {
		return fmt.Errorf("MetricaGauge.UpdateValueAtomic поптыка обновить метрику с несовпадающим наименованием ")
	}

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

// Restore - Откат изменения вызванного асинхронным изменением [UpdateValueStart]
func (m *MetricaGauge) Restore() {
	if m.newValue != nil {
		m.newValue = nil
	}
}

// Confirm - Подтверждение изменения вызванного асинхронным изменением [UpdateValueStart]
func (m *MetricaGauge) Confirm() {
	if m.newValue != nil {
		m.value = *(m.newValue)
		m.newValue = nil
	}
}

// MarshalJSON - сериализация данных в JSON через [MetricaJSONAdapter]
func (m *MetricaGauge) MarshalJSON() ([]byte, error) {
	return json.Marshal(MetricaJSONAdapter{Name: m.name, Kind: string(m.kind), Value: &m.value})
}
