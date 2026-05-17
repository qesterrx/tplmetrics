package model

import (
	"fmt"
	"strconv"
)

// Metrica - Основной интерфейс используемый в приложении для обмена между структурам
type Metrica interface {

	//Name - возвращает Имя метрики
	Name() string

	//Value - Возвращает значение метрики, приведенное к String
	Value() string

	//Kind - Возвращает тип метрики в виде KindValue
	Kind() KindValue

	//UpdateValueAtomic - обновление метрики в виде атомарной операции
	UpdateValueAtomic(Metrica) error

	//UpdateValueStart - обновление метрики в виде транзакционной операции, ожидающей подтверждения (Confirm) или отката (Restore)
	UpdateValueStart(Metrica) error

	//Restore - откат изменения при транзакционной операции
	Restore()

	//Confirm - подтверждение изменения при транзакционной операции
	Confirm()
}

// NewMetrica - Фабрика метрик.
// На основе имени, типа и значения создает экземляр метрики указанного типа
func NewMetrica(name string, kind string, value string) (Metrica, error) {

	kindValue, err := GetKindValue(kind)
	if err != nil {
		return nil, err
	}

	switch kindValue {
	case Counter:
		valueInt, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, err
		}

		return NewMetricaCounter(name, valueInt), nil
	case Gauge:
		valueFlt, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, err
		}

		return NewMetricaGauge(name, valueFlt), nil
	}
	return nil, fmt.Errorf("NewMetrica unimplemented kind %s", kind)
}
