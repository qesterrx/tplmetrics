package model

import (
	"encoding/json"
	"fmt"
)

// FormatMetricaCounter - Функция преобразования исходного значения Counter метрики (int64) в строку
func FormatMetricaCounter(value int64) string {
	return fmt.Sprintf("%d", value)
}

// MetricaCounter - структура для хранения Counter метрики
//
// generate:reset
type MetricaCounter struct {
	name     string
	kind     KindValue
	value    int64
	newValue *int64
}

// NewMetricaCounter - Функция возвращающая ссылку на экземпляр MetricaCounter с переданными именем/значением
func NewMetricaCounter(name string, value int64) *MetricaCounter {
	return &MetricaCounter{
		name:     name,
		kind:     Counter,
		value:    value,
		newValue: nil,
	}
}

// Name - метод возвращающий имя Counter метрики
func (m *MetricaCounter) Name() string {
	return m.name
}

// Kind - метод возвращающий тип Counter метрики в виде KindValue
func (m *MetricaCounter) Kind() KindValue {
	return m.kind
}

// Value - метод возвращающий значение Counter метрики в виде строки
// Для преобразования int64 в строку используется функция [FormatMetricaCounter]
func (m *MetricaCounter) Value() string {
	return FormatMetricaCounter(m.value)
}

// SrcValue - Метод возвращающий исходное значение метрики (тип int64)
// Данный метод отсутствует в интерфейсе Metrica но необходим  для сохранения данных в Postgresql т.к данные в БД нужно сохранять в исходном типе
// Кроме того с помощью этого метода реализовывается транзационность в сервисном слое
// На вход получает переменную env указывающую какое именно значение необходимо вернуть - подтвержденное или не подтвержденное
func (m *MetricaCounter) SrcValue(env string) int64 {
	if env == "new" && m.newValue != nil {
		return *m.newValue
	}
	return m.value
}

// UpdateValueAtomic - Метод изменяющий значение Counter метрики в синхронном режиме
func (m *MetricaCounter) UpdateValueAtomic(mtrk Metrica) error {

	if m.name != mtrk.Name() {
		return fmt.Errorf("MetricaCounter.UpdateValueAtomic поптыка обновить метрику с несовпадающим наименованием ")
	}

	if v, ok := mtrk.(*MetricaCounter); ok {
		m.value = v.value + m.value
		return nil
	} else {
		return fmt.Errorf("MetricaCounter.UpdateValueAtomic type mismatch: want MetricaCounter got %v", mtrk)
	}
}

// UpdateValueStart - Метод асинхронного измененения значение Counter метрики
// Возвращает ошибку в случае если другой процесс обновляет выбранную метрику в асинхронном режиме
func (m *MetricaCounter) UpdateValueStart(mtrk Metrica) error {

	if m.name != mtrk.Name() {
		return fmt.Errorf("MetricaCounter.UpdateValueStart поптыка обновить метрику с несовпадающим наименованием ")
	}

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
		return fmt.Errorf("MetricaCounter.UpdateValueStart type mismatch: want MetricaCounter got %v", mtrk)
	}
}

// Restore - Откат изменения вызванного асинхронным изменением [UpdateValueStart]
func (m *MetricaCounter) Restore() {
	if m.newValue != nil {
		m.newValue = nil
	}
}

// Confirm - Подтверждение изменения вызванного асинхронным изменением [UpdateValueStart]
func (m *MetricaCounter) Confirm() {
	if (m.newValue) != nil {
		m.value = *(m.newValue)
		m.newValue = nil
	}
}

// MarshalJSON - сериализация данных в JSON через [MetricaJSONAdapter]
func (m *MetricaCounter) MarshalJSON() ([]byte, error) {
	return json.Marshal(MetricaJSONAdapter{Name: m.name, Kind: string(m.kind), Delta: &m.value})
}
