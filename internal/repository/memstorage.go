package repository

import (
	"fmt"
	"strconv"
)

type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

func NewMemStorage() *MemStorage {
	mm := MemStorage{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}

	return &mm
}

func (m *MemStorage) UpdateMetric(kind KindValue, name string, value string) error {

	//А не будет ли это слабым местом, все таки при каждом сохранении будем два раза искать в мапе?
	skind, _, err := m.GetMetric(name)

	if err == nil && skind != kind {
		return fmt.Errorf("metric %s have already registred as %s", name, skind)
	}

	if kind == Gauge {

		val, err := strconv.ParseFloat(value, 64)

		if err != nil {
			return fmt.Errorf("type of value for Gauge metric is wrong (expected float64), got value %s", value)
		}

		m.gauge[name] = val
		return nil

	} else if kind == Counter {

		val, err := strconv.ParseInt(value, 10, 64)

		if err != nil {
			return fmt.Errorf("type of value for Counter metric is wrong (expected int64), got value %s", value)
		}

		m.counter[name] = m.counter[name] + val
		return nil

	} else {
		return fmt.Errorf("unimplemented type %s", kind)
	}
}

func (m *MemStorage) GetMetric(name string) (KindValue, string, error) {

	valueG, ok := m.gauge[name]

	if ok {
		return Gauge, fmt.Sprintf("%5.5f", valueG), nil
	}

	valueC, ok := m.counter[name]

	if ok {
		return Counter, fmt.Sprintf("%d", valueC), nil
	}

	return "", "", fmt.Errorf("metric %s not found", name)

}

func (m *MemStorage) Show() {
	fmt.Println("-------------GAUGE-------------")
	for k, v := range m.gauge {
		fmt.Printf("%s : %5.5f \n", k, v)
	}
	fmt.Println("-------------COUNTER-------------")
	for k, v := range m.counter {
		fmt.Printf("%s : %d \n", k, v)
	}
}
